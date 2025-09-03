package testsuite

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/arun0009/hatchetest/pkg/config"
	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// GlobalShared is the single shared test suite instance
var GlobalShared *SharedTestSuite
var globalSharedMutex sync.Mutex

// GetOrCreateGlobalShared returns the global shared instance, creating it if needed
// This ensures all tests across all packages use the same container instances
// Uses mutex to ensure thread-safe singleton creation
func GetOrCreateGlobalShared() *SharedTestSuite {
	globalSharedMutex.Lock()
	defer globalSharedMutex.Unlock()

	if GlobalShared == nil {
		log.Println("🏗️ Creating global shared test containers (first time)")
		GlobalShared = &SharedTestSuite{}
		GlobalShared.setupContainersOnly()
		log.Println("✅ Global shared test containers ready")
	} else {
		log.Println("♻️ Reusing existing global shared test containers")
	}
	return GlobalShared
}

// setupContainersOnly sets up containers without using testify suite methods
func (s *SharedTestSuite) setupContainersOnly() {
	log.Println("Setting up shared test containers...")

	ctx := context.Background()

	// Create network with unique name to avoid conflicts
	networkName := fmt.Sprintf("hatchet-test-network-%d", time.Now().UnixNano())
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: networkName,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}
	s.network = network.(*testcontainers.DockerNetwork)

	// Start Postgres container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "hatchet",
			"POSTGRES_USER":     "hatchet",
			"POSTGRES_PASSWORD": "hatchet",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		),
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"postgres"},
		},
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}
	s.postgresContainer = postgresContainer

	// Get postgres connection details
	postgresHost, err := postgresContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get postgres host: %v", err)
	}
	postgresPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("Failed to get postgres port: %v", err)
	}
	s.PostgresURL = fmt.Sprintf("postgres://hatchet:hatchet@%s:%s/hatchet?sslmode=disable", postgresHost, postgresPort.Port())

	// For Hatchet container, use internal network address for database connection
	internalPostgresURL := "postgres://hatchet:hatchet@postgres:5432/hatchet?sslmode=disable"

	// Start Hatchet container
	hatchetReq := testcontainers.ContainerRequest{
		Image:        "ghcr.io/hatchet-dev/hatchet/hatchet-lite:latest",
		ExposedPorts: []string{"8888/tcp", "7077/tcp"},
		Env: map[string]string{
			"DATABASE_URL":                                           internalPostgresURL,
			"SERVER_AUTH_COOKIE_DOMAIN":                              "localhost",
			"SERVER_AUTH_COOKIE_INSECURE":                            "t",
			"SERVER_GRPC_BIND_ADDRESS":                               "0.0.0.0",
			"SERVER_GRPC_INSECURE":                                   "t",
			"SERVER_GRPC_BROADCAST_ADDRESS":                          "localhost:7077",
			"SERVER_GRPC_PORT":                                       "7077",
			"SERVER_URL":                                             "http://localhost:8888",
			"SERVER_AUTH_SET_EMAIL_VERIFIED":                         "t",
			"SERVER_DEFAULT_ENGINE_VERSION":                          "V1",
			"SERVER_INTERNAL_CLIENT_INTERNAL_GRPC_BROADCAST_ADDRESS": "localhost:7077",
		},
		WaitingFor: wait.ForHTTP("/health").WithPort("8888/tcp").WithStartupTimeout(60 * time.Second),
		Networks:   []string{networkName},
	}

	hatchetContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: hatchetReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start hatchet container: %v", err)
	}
	s.hatchetContainer = hatchetContainer

	// Get hatchet connection details
	hatchetHost, err := hatchetContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get hatchet host: %v", err)
	}
	hatchetGRPCPort, err := hatchetContainer.MappedPort(ctx, "7077")
	if err != nil {
		log.Fatalf("Failed to get hatchet GRPC port: %v", err)
	}
	hatchetHTTPPort, err := hatchetContainer.MappedPort(ctx, "8888")
	if err != nil {
		log.Fatalf("Failed to get hatchet HTTP port: %v", err)
	}

	s.HatchetGRPCURL = fmt.Sprintf("%s:%s", hatchetHost, hatchetGRPCPort.Port())
	s.HatchetURL = fmt.Sprintf("http://%s:%s", hatchetHost, hatchetHTTPPort.Port())

	log.Printf("✅ Hatchet container available at:")
	log.Printf("   GRPC: %s", s.HatchetGRPCURL)
	log.Printf("   HTTP: %s", s.HatchetURL)

	// Verify hatchet health
	resp, err := http.Get(s.HatchetURL + "/health")
	if err != nil {
		log.Fatalf("Failed to check hatchet health: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Hatchet health check failed with status: %d", resp.StatusCode)
	}
	log.Println("✅ Hatchet container health check passed - ready for integration tests")

	// Generate fresh token from Hatchet container
	exitCode, execResp, err := s.hatchetContainer.Exec(ctx, []string{
		"/hatchet-admin", "token", "create",
		"--config", "/config",
		"--tenant-id", "707d0855-80ab-4e1f-a156-f1c4546cbf52",
	})
	if err != nil || exitCode != 0 {
		log.Fatalf("Failed to execute token creation command: %v, exit code: %d", err, exitCode)
	}
	tokenBytes, err := io.ReadAll(execResp)
	if err != nil {
		log.Fatalf("Failed to read token output: %v", err)
	}
	tokenStr := strings.TrimSpace(string(tokenBytes))

	re := regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`)
	matches := re.FindStringSubmatch(tokenStr)
	if len(matches) == 0 {
		log.Fatalf("Failed to find JWT token in output: %s", tokenStr)
	}
	token := matches[0]

	log.Printf("Generated token: %s", token)

	// Set environment variables for Hatchet client (same approach as main.go)
	os.Setenv("HATCHET_CLIENT_TOKEN", token)
	os.Setenv("HATCHET_CLIENT_HOST_PORT", s.HatchetGRPCURL)
	os.Setenv("HATCHET_CLIENT_SERVER_URL", s.HatchetURL)
	os.Setenv("HATCHET_CLIENT_TLS_STRATEGY", "none")

	// Create Hatchet client with token (same approach as main.go)
	hatchetClient, err := client.New(
		client.WithToken(token),
		client.WithHostPort(hatchetHost, hatchetGRPCPort.Int()),
	)
	if err != nil {
		log.Fatalf("Failed to create Hatchet client for tests: %v", err)
	}
	s.HatchetClient = hatchetClient
	log.Printf("✅ Hatchet client created for integration tests")

	// Test server will be created by the shared suite setup

	log.Println("Shared test containers ready!")
}

// SharedTestSuite provides shared testcontainer infrastructure for all integration tests
type SharedTestSuite struct {
	suite.Suite

	// Container infrastructure
	network           *testcontainers.DockerNetwork
	postgresContainer testcontainers.Container
	hatchetContainer  testcontainers.Container

	// Test clients and servers
	HatchetClient client.Client
	TestServer    *echo.Echo
	TestServerURL string

	// Connection details
	PostgresURL    string
	HatchetURL     string
	HatchetGRPCURL string
}

// SetupSuite runs once before any tests in the suite - starts all containers
func (s *SharedTestSuite) SetupSuite() {
	ctx := context.Background()

	log.Println("Setting up shared test containers...")

	// Create network for containers
	network, err := network.New(ctx, network.WithDriver("bridge"))
	s.Require().NoError(err, "Failed to create test network")
	s.network = network

	// Start PostgreSQL container
	s.startPostgresContainer(ctx)

	// Start Hatchet Lite container
	s.startHatchetContainer(ctx)

	// Start unified test server
	s.startTestServer()

	log.Println("Shared test containers ready!")
}

// TearDownSuite runs after all tests finish - cleans up containers
func (s *SharedTestSuite) TearDownSuite() {
	ctx := context.Background()

	log.Println("Cleaning up shared test containers...")

	if s.TestServer != nil {
		s.TestServer.Shutdown(ctx)
	}

	if s.hatchetContainer != nil {
		s.hatchetContainer.Terminate(ctx)
	}

	if s.postgresContainer != nil {
		s.postgresContainer.Terminate(ctx)
	}

	if s.network != nil {
		s.network.Remove(ctx)
	}

	log.Println("Shared test containers cleaned up!")
}

func (s *SharedTestSuite) startPostgresContainer(ctx context.Context) {
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:15-alpine",
			Env: map[string]string{
				"POSTGRES_DB":       "hatchet",
				"POSTGRES_USER":     "hatchet",
				"POSTGRES_PASSWORD": "hatchet",
			},
			ExposedPorts: []string{"5432/tcp"},
			// Wait for both port and database readiness
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("5432/tcp"),
				wait.ForLog("database system is ready to accept connections"),
			).WithStartupTimeout(60 * time.Second),
			Networks: []string{s.network.Name},
			NetworkAliases: map[string][]string{
				s.network.Name: {"postgres"},
			},
		},
		Started: true,
	})
	s.Require().NoError(err, "Failed to start postgres container")
	s.postgresContainer = postgres

	// Get external connection details
	host, err := postgres.Host(ctx)
	s.Require().NoError(err, "Failed to get postgres host")
	port, err := postgres.MappedPort(ctx, "5432")
	s.Require().NoError(err, "Failed to get postgres port")

	s.PostgresURL = fmt.Sprintf("postgres://hatchet:hatchet@%s:%s/hatchet?sslmode=disable", host, port.Port())
}

func (s *SharedTestSuite) startHatchetContainer(ctx context.Context) {
	hatchet, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/hatchet-dev/hatchet/hatchet-lite:latest",
			Env: map[string]string{
				"DATABASE_URL":                "postgres://hatchet:hatchet@postgres:5432/hatchet?sslmode=disable",
				"SERVER_GRPC_BIND_ADDRESS":    "0.0.0.0",
				"SERVER_GRPC_PORT":            "7077",
				"SERVER_HTTP_PORT":            "8888",
				"SERVER_AUTH_COOKIE_DOMAIN":   "localhost",
				"SERVER_AUTH_COOKIE_INSECURE": "true",
				"SERVER_GRPC_INSECURE":        "true",
			},
			ExposedPorts: []string{"8888/tcp", "7077/tcp"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("8888/tcp").WithStartupTimeout(120 * time.Second),
			Networks:     []string{s.network.Name},
		},
		Started: true,
	})
	s.Require().NoError(err, "Failed to start hatchet container")
	s.hatchetContainer = hatchet

	// Get external connection details
	host, err := hatchet.Host(ctx)
	s.Require().NoError(err, "Failed to get hatchet host")
	httpPort, err := hatchet.MappedPort(ctx, "8888")
	s.Require().NoError(err, "Failed to get hatchet http port")

	s.HatchetURL = fmt.Sprintf("http://%s:%s", host, httpPort.Port())
}

func (s *SharedTestSuite) startTestServer() {
	// Only create test server if it doesn't exist
	if s.TestServer == nil {
		s.TestServer = echo.New()
		s.TestServer.Use(middleware.Logger())
		s.TestServer.Use(middleware.Recover())
		s.TestServer.Use(middleware.CORS())

		// Health check endpoint
		s.TestServer.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		})

		// Use a unique port based on timestamp to avoid conflicts
		port := 8080 + (time.Now().UnixNano() % 1000)
		portStr := fmt.Sprintf(":%d", port)

		// Start server in background
		go func() {
			if err := s.TestServer.Start(portStr); err != nil && err != http.ErrServerClosed {
				log.Printf("Test server error: %v", err)
			}
		}()

		// Wait for server to be ready
		time.Sleep(2 * time.Second)
		s.TestServerURL = fmt.Sprintf("http://localhost:%d", port)
		log.Printf("Test server started on %s", s.TestServerURL)
	}
}

// RegisterModules allows test packages to register their routes
func (s *SharedTestSuite) RegisterModules(registerFuncs ...func(*echo.Echo, client.Client, *config.AppConfig)) {
	// Ensure test server is started first
	s.startTestServer()

	// Create a config for tests with Hatchet connection info
	cfg := &config.AppConfig{
		Port:             8081,
		Host:             "localhost",
		HatchetHostPort:  os.Getenv("HATCHET_CLIENT_HOST_PORT"),
		HatchetServerURL: os.Getenv("HATCHET_CLIENT_SERVER_URL"),
		HatchetToken:     os.Getenv("HATCHET_CLIENT_TOKEN"),
	}
	for _, registerFunc := range registerFuncs {
		registerFunc(s.TestServer, s.HatchetClient, cfg)
	}
}

// TearDown cleans up all shared test resources
func (s *SharedTestSuite) TearDown() error {
	ctx := context.Background()
	var errors []string

	log.Println("🧹 Tearing down shared test containers...")

	// Stop test server
	if s.TestServer != nil {
		if err := s.TestServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("test server shutdown: %v", err))
		}
	}

	// Terminate Hatchet container
	if s.hatchetContainer != nil {
		if err := s.hatchetContainer.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("hatchet container: %v", err))
		}
	}

	// Terminate Postgres container
	if s.postgresContainer != nil {
		if err := s.postgresContainer.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("postgres container: %v", err))
		}
	}

	// Remove network
	if s.network != nil {
		if err := s.network.Remove(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("network removal: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	log.Println("✅ Shared test containers cleaned up successfully")
	return nil
}
