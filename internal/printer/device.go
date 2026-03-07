package printer

import (
	"fmt"
	"log"
	"os"
)

// Device represents an abstraction over a physical printer.
type Device interface {
	Write(data []byte) error
}

// RealDevice writes raw bytes directly to the given device path (e.g. /dev/usb/lp0).
type RealDevice struct {
	DevicePath string
}

// Write writes data to real device
func (r *RealDevice) Write(data []byte) error {
	file, err := os.OpenFile(r.DevicePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open device %s: %w", r.DevicePath, err)
	}
	defer file.Close()

	n, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to device %s: %w", r.DevicePath, err)
	}

	log.Printf("Successfully wrote %d bytes to %s", n, r.DevicePath)
	return nil
}

// MockDevice logs the data it receives, useful for development and testing.
type MockDevice struct {
	LastWritten []byte
}

// Write writes data to mock device
func (m *MockDevice) Write(data []byte) error {
	m.LastWritten = append(m.LastWritten, data...)

	log.Printf("MockDevice: Received %d bytes", len(data))
	log.Printf("MockDevice: Hex: %x", data)
	log.Printf("MockDevice: String (lossy): %s", string(data))

	return nil
}
