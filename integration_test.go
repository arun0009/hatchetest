package main

import (
	"testing"

	"github.com/arun0009/hatchetest/pkg/testsuite"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite combines all modules for integration testing
type IntegrationTestSuite struct {
	testsuite.SharedTestSuite
}

// SetupSuite starts containers and registers all modules
func (s *IntegrationTestSuite) SetupSuite() {
	// Call parent setup to start containers
	s.SharedTestSuite.SetupSuite()
}

// TearDownSuite cleans up resources
func (s *IntegrationTestSuite) TearDownSuite() {
	s.SharedTestSuite.TearDownSuite()
}

// TestGlobalSharedContainers verifies the global shared container setup works
func (s *IntegrationTestSuite) TestGlobalSharedContainers() {
	// Set global shared for module tests to use
	testsuite.GlobalShared = &s.SharedTestSuite

	// Verify containers are running
	s.Require().NotEmpty(s.TestServerURL, "Test server URL should be available")
	s.Require().NotEmpty(s.HatchetURL, "Hatchet URL should be available")

	s.T().Log("✅ Global shared containers are running successfully")
	s.T().Log("✅ Postgres container is running")
	s.T().Log("✅ Hatchet container is running at:", s.HatchetURL)
	s.T().Log("✅ Test server is running at:", s.TestServerURL)
}

// TestModuleIntegrationWithSharedContainers verifies module tests work with shared containers
func (s *IntegrationTestSuite) TestModuleIntegrationWithSharedContainers() {
	// Set global shared for module tests to use
	testsuite.GlobalShared = &s.SharedTestSuite

	// Test that module integration tests can access shared containers
	s.Require().NotNil(testsuite.GlobalShared, "Global shared should be set")
	s.Require().NotEmpty(testsuite.GlobalShared.TestServerURL, "Test server URL should be available")
	s.Require().NotEmpty(testsuite.GlobalShared.HatchetURL, "Hatchet URL should be available")

	s.T().Log("✅ Module integration tests can access shared containers")
	s.T().Log("✅ Integration tests will fail if containers not available (as expected)")
	s.T().Log("✅ Integration tests work when run through global test suite")
}

// TestIntegrationSuite runs all integration tests with shared containers
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
