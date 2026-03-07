# Thermo Printer Backend

A lightweight Go backend that serves a REST API for a connected thermal printer (defaulting to `/dev/usb/lp0`). This service receives text via HTTP POST requests, formats it specifically for this thermal printer hardware, and sends the raw bytes to the device.

## Features

- **Automatic Text Wrapping:** Intelligently wraps long text to the printer's maximum width of 24 characters, preserving word boundaries where possible.
- **CP850 Encoding Support:** The printer expects characters in the CP850 DOS encodings (Code Page 850). This application automatically converts UTF-8 strings to CP850 to properly support special characters like German Umlaute (`äöüß`).
- **CRLF Conversions:** Automatically sanitizes the input replacing `\n` with `\r\n` (DOS line endings).
- **Mocking Support:** Includes a developer mock mode to print logs to the console rather than requiring physical printer hardware.
- **Docker Ready:** Provides an optimized 10MB multi-stage Docker image and a `docker-compose.yml` pre-configured to mount host USB devices.

## API Usage

**Endpoint:** `POST /print`

### Request Payload

```json
{
  "text": "Hello Printer, this string will be converted, wrapped and encoded as CP850."
}
```

### Example Usage (cURL)

```bash
curl -X POST http://localhost:8080/print \
  -H "Content-Type: application/json" \
  -d '{"text": "abcdefghijklmnopqrstuvwxyzäöüßABCDEFGHIJKLMNOPQRSTUVWXY"}'
```

## Running the Service

### 1. Local Development (Mock Mode)

To run the application locally without an attached thermal printer, use the mock device:

```bash
PRINTER_DEVICE=mock go run ./cmd/server/main.go
```
Any print requests sent to the local server will be logged with their byte sequences lossy-string conversions.

### 2. Physical Machine (Real Printer)

By default, the application writes to `/dev/usb/lp0`. Run natively compiled:

```bash
go build -o thermo-printer ./cmd/server
./thermo-printer
```
*(Optionally define a custom path with `PRINTER_DEVICE=/dev/ttyUSB0 ./thermo-printer`)*

### 3. Using Docker Compose

For easy deployment, run the included `docker-compose.yml`. This automatically mounts the host's `/dev/usb/lp0` into the container.

```bash
docker-compose up -d --build
```

You can customize the device mount inside `docker-compose.yml` if your printer resolves to a different character device on the host.

## Environment Variables

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT` | The HTTP port the server listens on | `8080` |
| `PRINTER_DEVICE` | Path to the printer device. Set to `mock` for testing. | `/dev/usb/lp0` |

## Testing

Unit tests cover the heavy lifting of CP850 mappings and intelligent text wrapping. 
To run all tests:

```bash
go test -v ./...
```
