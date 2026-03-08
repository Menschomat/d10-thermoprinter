# Thermo Printer Backend

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=Menschomat_d10-thermoprinter&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=Menschomat_d10-thermoprinter)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=Menschomat_d10-thermoprinter&metric=coverage)](https://sonarcloud.io/summary/new_code?id=Menschomat_d10-thermoprinter)

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
  "text": "Hello Printer, this string will be converted, wrapped and encoded as CP850.",
  "align": "center"
}
```

- **text** (string, required): The content to print.
- **align** (string, optional): Text alignment. Allowed values: `"left"`, `"center"`, `"right"`. Defaults to `"left"`.

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

For easy deployment, run the included `docker-compose.yml`. This automatically mounts the host's `/dev/usb/lp0` into the container and runs the latest main branch image.

```bash
# Start the container
docker-compose up -d

# Or force a local build instead of pulling the image
docker-compose up -d --build
```

You can customize the device mount inside `docker-compose.yml` if your printer resolves to a different character device on the host.

### 4. Using Docker Run

If you prefer not to use Docker Compose, you can run the container directly with `docker run`. You must map the printer device and provide access to the `lp` (printer) group (often GID `7` on Debian/Alpine/Raspberry Pi):

```bash
docker run -d \
  --name thermoprinter \
  -p 8080:8080 \
  --device=/dev/usb/lp0:/dev/usb/lp0 \
  --group-add 7 \
  -e PRINTER_DEVICE=/dev/usb/lp0 \
  -e PRINTER_DELAY_MS=500 \
  menschomat/d10-thermoprinter:main
```

## Environment Variables

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `PORT` | The HTTP port the server listens on | `8080` |
| `PRINTER_DEVICE` | Path to the printer device. Set to `mock` for testing. | `/dev/usb/lp0` |
| `PRINTER_DELAY_MS` | Delay in milliseconds between writing each line to the printer buffer | `500` |
| `PRINTER_BLANK_LINES` | Number of trailing blank lines to append after each print job | `4` |
| `PRINTER_MAX_BYTES` | Maximum allowed request body size in bytes | `10240` |
| `PRINTER_MAX_TEXT_LENGTH` | Maximum allowed characters in the print text | `1000` |

## Testing

Unit tests cover the heavy lifting of CP850 mappings and intelligent text wrapping.
To run all tests:

```bash
go test -v ./...
```
