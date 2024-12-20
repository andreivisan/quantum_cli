package ai

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Chat(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		serverResp string
		wantErr    bool
	}{
		{
			name:       "successful chat",
			message:    "Hello",
			serverResp: "THINKING:\nResponse text here",
			wantErr:    false,
		},
		{
			name:       "empty message",
			message:    "",
			serverResp: "THINKING:\nEmpty message received",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and content type
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Write response in chunks to simulate streaming
				flusher, ok := w.(http.Flusher)
				if !ok {
					t.Fatal("Expected http.Flusher")
				}

				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				for _, chunk := range strings.Split(tt.serverResp, " ") {
					_, err := w.Write([]byte(chunk + " "))
					if err != nil {
						t.Fatal(err)
					}
					flusher.Flush()
					time.Sleep(10 * time.Millisecond)
				}
			}))
			defer ts.Close()

			// Create client with test server URL
			client := NewClient(ts.URL)
			outputChan := make(chan string)

			// Run chat in goroutine
			done := make(chan bool)
			var chatErr error
			go func() {
				chatErr = client.Chat(tt.message, outputChan)
				close(outputChan)
				done <- true
			}()

			// Collect response
			var response []string
			for msg := range outputChan {
				response = append(response, msg)
			}

			<-done // Wait for chat to complete

			// Check for errors
			if (chatErr != nil) != tt.wantErr {
				t.Errorf("Client.Chat() error = %v, wantErr %v", chatErr, tt.wantErr)
			}

			// Verify we got some response
			if len(response) == 0 && !tt.wantErr {
				t.Error("Expected non-empty response")
			}
		})
	}
}
