package printer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Device represents an abstraction over a physical printer.
type Device interface {
	Write(data []byte) error
}

// RealDevice writes raw bytes directly to the given device path (e.g. /dev/usb/lp0).
type RealDevice struct {
	DevicePath string
	Delay      time.Duration
	mu         sync.Mutex
}

// Write writes data to real device
func (r *RealDevice) Write(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.OpenFile(r.DevicePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open device %s: %w", r.DevicePath, err)
	}
	defer file.Close()

	lines := bytes.SplitAfter(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		n, err := file.Write(line)
		if err != nil {
			return fmt.Errorf("failed to write to device %s: %w", r.DevicePath, err)
		}

		log.Printf("Successfully wrote %d bytes to %s", n, r.DevicePath)

		if r.Delay > 0 {
			time.Sleep(r.Delay)
		}
	}

	return nil
}

// MockDevice logs the data it receives, useful for development and testing.
type MockDevice struct {
	LastWritten []byte
	mu          sync.Mutex
}

// Write writes data to mock device
func (m *MockDevice) Write(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LastWritten = append(m.LastWritten, data...)

	log.Printf("MockDevice: Received %d bytes", len(data))
	log.Printf("MockDevice: Hex: %x", data)
	log.Printf("MockDevice: String (lossy): %s", string(data))

	return nil
}
