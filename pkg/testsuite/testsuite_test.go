package testsuite

import (
	"testing"
)

// This file exists solely to trigger TestMain in shared_suite.go
// when running go test ./...

// TestDummy is a placeholder test to ensure TestMain runs
func TestDummy(t *testing.T) {
	// This test does nothing but ensures TestMain is called
	// which sets up GlobalShared for other packages to use
	shared := GetOrCreateGlobalShared()
	if shared == nil {
		t.Fatal("Failed to get or create GlobalShared")
	}
	t.Log("✅ GlobalShared containers are available for integration tests")
}
