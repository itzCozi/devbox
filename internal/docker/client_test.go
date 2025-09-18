package docker

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestClientClose(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
}

func TestIsDockerAvailable(t *testing.T) {

	err := IsDockerAvailable()

	if err != nil {
		t.Logf("Docker not available (this is expected in some test environments): %v", err)

		if !contains(err.Error(), "docker") {
			t.Errorf("Expected error to mention docker, got: %v", err)
		}
	} else {
		t.Log("Docker is available")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAtIndex(s, substr)
}

func containsAtIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
