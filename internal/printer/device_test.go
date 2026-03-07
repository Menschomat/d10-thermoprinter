package printer

import (
	"os"
	"testing"
	"time"
)

func TestRealDevice_WriteWithDelay(t *testing.T) {
	// Create a temporary file to act as our "device"
	tmpFile, err := os.CreateTemp("", "mock-device-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close() // RealDevice will open it

	delay := 10 * time.Millisecond // Use a smaller delay for testing so test isn't slow
	device := &RealDevice{
		DevicePath: tmpFile.Name(),
		Delay:      delay,
	}

	data := []byte("Line 1\nLine 2\nLine 3\n")

	start := time.Now()
	err = device.Write(data)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Read what was written to the file
	writtenData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read back from temp file: %v", err)
	}

	if string(writtenData) != string(data) {
		t.Errorf("Expected written data to be %q, got %q", string(data), string(writtenData))
	}

	// We wrote 3 lines, so we should have slept 3 times.
	// Allow for some OS scheduling fuzziness (e.g. 80% of expected minimum).
	expectedMinDuration := time.Duration(float64(3*delay) * 0.8)

	// Check if the elapsed time is reasonably close to expected
	if elapsed < expectedMinDuration {
		t.Errorf("Expected execution to take at least %v due to delays, took %v", expectedMinDuration, elapsed)
	}
}
