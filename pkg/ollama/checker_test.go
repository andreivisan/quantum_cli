package ollama

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func TestChecker_IsServerRunning(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		want       bool
	}{
		{
			name: "server is running",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want: true,
		},
		{
			name: "server returns error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: true, // Still true because server responded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer ts.Close()

			// Create checker with test server URL
			c := NewChecker(ts.URL)

			// Test server status check
			if got := c.IsServerRunning(); got != tt.want {
				t.Errorf("Checker.IsServerRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChecker_CheckInstallation(t *testing.T) {
	// Store original and defer restore
	originalLookPath := lookPath
	defer func() { lookPath = originalLookPath }()

	tests := []struct {
		name    string
		mockErr error
		want    bool
	}{
		{
			name:    "ollama_not_installed",
			mockErr: exec.ErrNotFound,
			want:    false,
		},
		{
			name:    "ollama_is_installed",
			mockErr: nil,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock lookPath for this test case
			lookPath = func(file string) (string, error) {
				if tt.mockErr != nil {
					return "", tt.mockErr
				}
				return "/usr/local/bin/ollama", nil
			}

			c := NewChecker("http://localhost:11434")
			got := c.CheckInstallation()
			if got != tt.want {
				t.Errorf("CheckInstallation() = %v, want %v", got, tt.want)
			}
		})
	}
}
