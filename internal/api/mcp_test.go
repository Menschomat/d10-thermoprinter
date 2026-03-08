package api

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/menschomat/d10-thermoprinter/internal/printer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// helper to build a Server wired to the given device for MCP tests.
func newMCPServer(device printer.Device, blankLines int) *Server {
	return NewServer(device, blankLines, 10*1024, 1000)
}

func TestMCPPrint_Success(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text: "Hello Printer",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success but got IsError=true: %+v", result)
	}
	if len(mock.LastWritten) == 0 {
		t.Error("expected data written to device, got none")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(textContent.Text, "success") {
		t.Errorf("expected success message, got: %+v", result.Content)
	}
}

func TestMCPPrint_DefaultAlignment(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	// Empty align should default to "left" and succeed.
	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text:  "Test",
		Align: "",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected default alignment to succeed, got error: %+v", result)
	}
}

func TestMCPPrint_CenterAlignment(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text:  "Centered",
		Align: "center",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected center alignment to succeed: %+v", result)
	}
}

func TestMCPPrint_InvalidAlignment(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text:  "Test",
		Align: "justify",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for invalid alignment")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(textContent.Text, "Invalid alignment") {
		t.Errorf("expected alignment error message, got: %+v", result.Content)
	}
}

func TestMCPPrint_BadCharset(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	// Emoji is not representable in CP850 — FormatText should fail.
	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text: "Hello 🖨️",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for unsupported characters")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(textContent.Text, "Error formatting") {
		t.Errorf("expected formatting error message, got: %+v", result.Content)
	}
}

func TestMCPPrint_DeviceWriteError(t *testing.T) {
	s := newMCPServer(&failingDevice{}, 0)

	result, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text: "Fail me",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true when device write fails")
	}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(textContent.Text, "Error writing") {
		t.Errorf("expected device write error message, got: %+v", result.Content)
	}
}

func TestMCPPrint_BlankLines(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 3)

	_, _, err := s.mcpPrint(context.Background(), &mcp.CallToolRequest{}, PrintParams{
		Text: "With Blanks",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(mock.LastWritten, []byte("With Blanks")) {
		t.Error("expected written data to contain original text")
	}
}

func TestSetupMCPServer(t *testing.T) {
	mock := &printer.MockDevice{}
	s := newMCPServer(mock, 0)

	mcpSrv := s.SetupMCPServer()
	if mcpSrv == nil {
		t.Fatal("expected a non-nil MCP server")
	}
}
