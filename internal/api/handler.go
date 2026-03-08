package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/menschomat/d10-thermoprinter/internal/printer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server handles the HTTP API endpoints.
type Server struct {
	Device        printer.Device
	Router        *mux.Router
	BlankLines    int
	MaxBytes      int64
	MaxTextLength int
}

// PrintRequest defines the JSON payload for printing.
type PrintRequest struct {
	Text  string `json:"text"`
	Align string `json:"align,omitempty"`
}

// NewServer initializes and returns a new Server instance.
func NewServer(device printer.Device, blankLines int, maxBytes int64, maxTextLength int) *Server {
	s := &Server{
		Device:        device,
		Router:        mux.NewRouter(),
		BlankLines:    blankLines,
		MaxBytes:      maxBytes,
		MaxTextLength: maxTextLength,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.HandleFunc("/print", s.handlePrint()).Methods("POST")

	mcpServer := s.SetupMCPServer()
	mcpHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	s.Router.PathPrefix("/mcp").Handler(mcpHandler)
}

func (s *Server) parsePrintRequest(w http.ResponseWriter, r *http.Request) (PrintRequest, bool) {
	var req PrintRequest
	// Limit the request body size to prevent memory exhaustion DoS
	r.Body = http.MaxBytesReader(w, r.Body, s.MaxBytes)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, fmt.Sprintf("Request body too large. Maximum size is %d bytes.", s.MaxBytes), http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
		}
		return req, false
	}
	defer r.Body.Close()

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return req, false
	}

	if req.Text == "" {
		http.Error(w, "Text cannot be empty", http.StatusBadRequest)
		return req, false
	}

	// Limit the text length to prevent hardware DoS (wasting paper/overheating)
	if len(req.Text) > s.MaxTextLength {
		http.Error(w, fmt.Sprintf("Text is too long. Maximum allowed length is %d characters.", s.MaxTextLength), http.StatusBadRequest)
		return req, false
	}

	if req.Align == "" {
		req.Align = "left"
	}
	if req.Align != "left" && req.Align != "center" && req.Align != "right" {
		http.Error(w, "Invalid alignment. Allowed values: left, center, right", http.StatusBadRequest)
		return req, false
	}

	return req, true
}

func (s *Server) handlePrint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := s.parsePrintRequest(w, r)
		if !ok {
			return
		}

		textToPrint := req.Text
		if s.BlankLines > 0 {
			textToPrint += strings.Repeat("\n", s.BlankLines)
		}

		// Format the text (wrap, CP850, CRLF, align)
		formatted, err := printer.FormatText(textToPrint, req.Align)
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
