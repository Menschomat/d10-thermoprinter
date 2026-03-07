package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/menschomat/d10-thermoprinter/internal/api"
	"github.com/menschomat/d10-thermoprinter/internal/printer"
)

func main() {
	devicePath := os.Getenv("PRINTER_DEVICE")
	var device printer.Device

	if devicePath == "mock" {
		log.Println("Using Mock Printer Device for development/testing.")
		device = &printer.MockDevice{}
	} else {
		if devicePath == "" {
			devicePath = "/dev/usb/lp0" // Default device
		}

		delayMs := 500
		if delayStr := os.Getenv("PRINTER_DELAY_MS"); delayStr != "" {
			if parsed, err := strconv.Atoi(delayStr); err == nil {
				delayMs = parsed
			} else {
				log.Printf("Invalid PRINTER_DELAY_MS '%s', falling back to %d ms", delayStr, delayMs)
			}
		}

		log.Printf("Using Real Printer Device at %s with %dms delay", devicePath, delayMs)
		device = &printer.RealDevice{
			DevicePath: devicePath,
			Delay:      time.Duration(delayMs) * time.Millisecond,
		}
	}

	blankLines := 4 // Default value
	if blanksStr := os.Getenv("PRINTER_BLANK_LINES"); blanksStr != "" {
		if parsed, err := strconv.Atoi(blanksStr); err == nil && parsed >= 0 {
			blankLines = parsed
		} else {
			log.Printf("Invalid PRINTER_BLANK_LINES '%s', falling back to %d", blanksStr, blankLines)
		}
	}

	maxBytes := int64(10 * 1024) // Default 10KB
	if bytesStr := os.Getenv("PRINTER_MAX_BYTES"); bytesStr != "" {
		if parsed, err := strconv.ParseInt(bytesStr, 10, 64); err == nil && parsed > 0 {
			maxBytes = parsed
		} else {
			log.Printf("Invalid PRINTER_MAX_BYTES '%s', falling back to %d", bytesStr, maxBytes)
		}
	}

	maxTextLength := 1000 // Default 1000
	if lengthStr := os.Getenv("PRINTER_MAX_TEXT_LENGTH"); lengthStr != "" {
		if parsed, err := strconv.Atoi(lengthStr); err == nil && parsed > 0 {
			maxTextLength = parsed
		} else {
			log.Printf("Invalid PRINTER_MAX_TEXT_LENGTH '%s', falling back to %d", lengthStr, maxTextLength)
		}
	}

	log.Printf("Config: Appending %d blank lines, MaxBytes: %d, MaxTextLength: %d", blankLines, maxBytes, maxTextLength)

	server := api.NewServer(device, blankLines, maxBytes, maxTextLength)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, server.Router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
