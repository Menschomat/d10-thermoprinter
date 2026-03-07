package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/menschomat/d10-thermoprinter/internal/printer"
)

// Server handles the HTTP API endpoints.
type Server struct {
	Device     printer.Device
	Router     *mux.Router
	BlankLines int
}

// PrintRequest defines the JSON payload for printing.
type PrintRequest struct {
	Text string `json:"text"`
}

// NewServer initializes and returns a new Server instance.
func NewServer(device printer.Device, blankLines int) *Server {
	s := &Server{
		Device:     device,
		Router:     mux.NewRouter(),
		BlankLines: blankLines,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.HandleFunc("/print", s.handlePrint()).Methods("POST")
}

func (s *Server) handlePrint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit the request body size to 10KB to prevent memory exhaustion DoS
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			if err.Error() == "http: request body too large" {
				http.Error(w, "Request body too large. Maximum size is 10KB.", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
			}
			return
		}
		defer r.Body.Close()

		var req PrintRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.Text == "" {
			http.Error(w, "Text cannot be empty", http.StatusBadRequest)
			return
		}

		// Limit the text length to prevent hardware DoS (wasting paper/overheating)
		if len(req.Text) > 1000 {
			http.Error(w, "Text is too long. Maximum allowed length is 1000 characters.", http.StatusBadRequest)
			return
		}

		textToPrint := req.Text
		if s.BlankLines > 0 {
			textToPrint += strings.Repeat("\n", s.BlankLines)
		}

		// Format the text (wrap, CP850, CRLF)
		formatted, err := printer.FormatText(textToPrint)
		if err != nil {
			log.Printf("Error formatting text: %v", err)
			http.Error(w, "Error formatting text for printer. Make sure all characters are supported.", http.StatusBadRequest)
			return
		}

		// Send to the device
		if err := s.Device.Write(formatted); err != nil {
			log.Printf("Error writing to printer device: %v", err)
			http.Error(w, "Error writing to printer device", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "message": "Printed successfully"}`))
	}
}
