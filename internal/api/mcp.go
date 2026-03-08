package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/menschomat/d10-thermoprinter/internal/printer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PrintParams struct {
	Text  string `json:"text"`
	Align string `json:"align,omitempty"`
}

func (s *Server) mcpPrint(ctx context.Context, req *mcp.CallToolRequest, params PrintParams) (*mcp.CallToolResult, any, error) {
	textToPrint := params.Text
	if s.BlankLines > 0 {
		textToPrint += strings.Repeat("\n", s.BlankLines)
	}

	align := params.Align
	if align == "" {
		align = "left"
	}
	if align != "left" && align != "center" && align != "right" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: Invalid alignment. Allowed values: left, center, right"},
			},
			IsError: true,
		}, nil, nil
	}

	formatted, err := printer.FormatText(textToPrint, align)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error formatting text: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	if err := s.Device.Write(formatted); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error writing to printer device: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Printed successfully"},
		},
	}, nil, nil
}

func (s *Server) SetupMCPServer() *mcp.Server {
	// The Version field might be required.
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "d10-thermoprinter", Version: "1.0.0"}, nil)

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "print_text",
		Description: "Print text to the thermal printer. Alignment can be left, center, or right.",
	}, s.mcpPrint)

	return mcpServer
}
