package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/arun0009/hatchetest/pkg/config"
	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/worker"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load application configuration
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Hatchet client (required)
	if cfg.HatchetToken == "" {
		log.Fatalf("HATCHET_CLIENT_TOKEN is required")
	}
	if cfg.HatchetServerURL == "" {
		log.Fatalf("HATCHET_CLIENT_SERVER_URL is required")
	}

	host, port, err := net.SplitHostPort(cfg.HatchetHostPort)
	if err != nil {
		log.Fatalf("Error splitting hatchet host and port: %v", err)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Error converting hatchet port to int: %v", err)
	}

	// Set environment variables for Hatchet client initialization
	// This is needed because defaultClientOpts loads from env vars first
	os.Setenv("HATCHET_CLIENT_TOKEN", cfg.HatchetToken)
	os.Setenv("HATCHET_CLIENT_HOST_PORT", cfg.HatchetHostPort)
	os.Setenv("HATCHET_CLIENT_SERVER_URL", cfg.HatchetServerURL)

	hatchetClient, err := client.New(
		client.WithToken(cfg.HatchetToken),
		client.WithHostPort(host, portInt),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Hatchet client: %v", err)
	}

	// Create Echo server
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "hatchetest",
		})
	})

	// Register Hatchet workflows with worker
	w, err := worker.NewWorker(
		worker.WithClient(hatchetClient),
		worker.WithName("hatchetest-worker"),
	)
	if err != nil {
		log.Fatalf("Failed to create Hatchet worker: %v", err)
	}

	// Start Hatchet worker
	go func() {
		log.Println("Starting Hatchet worker...")
		cleanup, err := w.Start()
		if err != nil {
			log.Printf("Hatchet worker error: %v", err)
		}
		defer cleanup()
	}()

	// Start server
	log.Printf("Starting unified server on port %d", cfg.Port)
	if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
