package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/menschomat/d10-thermoprinter/internal/printer"
)

// Server handles the HTTP API endpoints.
type Server struct {
	Device printer.Device
	Router *mux.Router
}

// PrintRequest defines the JSON payload for printing.
type PrintRequest struct {
	Text string `json:"text"`
}

// NewServer initializes and returns a new Server instance.
func NewServer(device printer.Device) *Server {
	s := &Server{
		Device: device,
		Router: mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.HandleFunc("/print", s.handlePrint()).Methods("POST")
}

func (s *Server) handlePrint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
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

		// Format the text (wrap, CP850, CRLF)
		formatted, err := printer.FormatText(req.Text)
		if err != nil {
			log.Printf("Error formatting text: %v", err)
			http.Error(w, "Error formatting text for printer", http.StatusInternalServerError)
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
