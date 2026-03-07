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

	server := api.NewServer(device)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, server.Router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
