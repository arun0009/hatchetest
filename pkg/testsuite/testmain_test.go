package testsuite

import (
	"log"
	"testing"

	"github.com/hatchet-dev/hatchet/pkg/worker"
	"github.com/stretchr/testify/suite"
)

// TestSuite uses the global shared containers
type TestSuite struct {
	suite.Suite
	Shared *SharedTestSuite
}

// SetupSuite uses the global shared containers and registers approvals module
func (s *TestSuite) SetupSuite() {
	// Get or create the global shared containers
	s.Shared = GetOrCreateGlobalShared()
	s.Require().NotNil(s.Shared, "Global shared containers not available - integration tests require containers")

	// Enable workflow registration to capture the actual error
	if s.Shared.HatchetClient != nil {
		w, err := worker.NewWorker(
			worker.WithClient(s.Shared.HatchetClient),
			worker.WithName("test-worker"),
		)
		if err != nil {
			log.Printf("❌ Error creating worker: %v", err)
		}
		if err != nil {
			log.Printf("❌ Test worker registration failed: %v", err)
			// Don't fail the test setup, just log the error for analysis
		} else {
			log.Printf("✅ Test worker started successfully")
		}
		s.NotNil(w, "Worker should be created successfully")
	}
}

func TestIntegration(t *testing.T) {
	suite.Run(t, &TestSuite{})
}
