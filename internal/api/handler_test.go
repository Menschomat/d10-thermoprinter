package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/menschomat/d10-thermoprinter/internal/printer"
)

func TestHandlePrint(t *testing.T) {
	mockDevice := &printer.MockDevice{}
	server := NewServer(mockDevice, 0, 10*1024, 1000)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		expectWrite    bool
		expectedError  string
	}{
		{
			name:           "Valid payload",
			payload:        `{"text": "Hello Printer"}`,
			expectedStatus: http.StatusOK,
			expectWrite:    true,
		},
		{
			name:           "Empty text",
			payload:        `{"text": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectWrite:    false,
		},
		{
			name:           "Invalid JSON",
			payload:        `{"text": `,
			expectedStatus: http.StatusBadRequest,
			expectWrite:    false,
		},
		{
			name:           "Text too long",
			payload:        `{"text": "` + string(make([]byte, 1001)) + `"}`, // 1001 characters
			expectedStatus: http.StatusBadRequest,
			expectWrite:    false,
		},
		{
			name:           "Payload too large",
			payload:        `{"text": "` + string(make([]byte, 11*1024)) + `"}`, // 11 KB
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectWrite:    false,
		},
		{
			name:           "Unsupported encoding character (emoji)",
			payload:        `{"text": "Hello 🖨️"}`, // Printer emoji
			expectedStatus: http.StatusBadRequest,
			expectWrite:    false,
		},
		{
			name:           "Print with BlankLines",
			payload:        `{"text": "With Blanks"}`,
			expectedStatus: http.StatusOK,
			expectWrite:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set BlankLines > 0 for this specific test
			if tt.name == "Print with BlankLines" {
				server.BlankLines = 2
			} else {
				server.BlankLines = 0
			}

			// Reset mock
			mockDevice.LastWritten = nil

			req, err := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(tt.payload)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.Router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectWrite {
				if len(mockDevice.LastWritten) == 0 {
					t.Errorf("Expected data to be written to device, but got 0 bytes")
				}
				if tt.name == "Print with BlankLines" {
					if !bytes.Contains(mockDevice.LastWritten, []byte("With Blanks")) {
						t.Errorf("Expected mocked output to contain text")
					}
				}
			} else {
				if len(mockDevice.LastWritten) > 0 {
					t.Errorf("Expected no data to be written, but got %d bytes", len(mockDevice.LastWritten))
				}
			}
		})
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

func TestHandlePrint_BodyReadError(t *testing.T) {
	mockDevice := &printer.MockDevice{}
	server := NewServer(mockDevice, 0, 10*1024, 1000)

	req, err := http.NewRequest("POST", "/print", errReader(0))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusBadRequest)
	}
}

type failingDevice struct{}

func (f *failingDevice) Write(data []byte) error {
	return bytes.ErrTooLarge // return some arbitrary error
}

func TestHandlePrint_DeviceError(t *testing.T) {
	device := &failingDevice{}
	server := NewServer(device, 0, 10*1024, 1000)

	req, err := http.NewRequest("POST", "/print", bytes.NewBuffer([]byte(`{"text": "Fail me"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}
}
