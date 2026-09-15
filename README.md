# TRANSFER-AGENT

A TCP-based file transfer tool with checksum verification. Send files between machines over a persistent TCP connection with automatic integrity validation.

## Features

- **TCP transfer** — Direct sender-to-receiver file transfer over TCP
- **Checksum verification** — SHA-256 integrity validation after transfer
- **Path traversal protection** — Sanitized filenames prevent directory attacks
- **Graceful shutdown** — Clean signal handling and connection draining
- **Connection timeouts** — 5-minute deadline prevents hung connections
- **Concurrent connections** — Receiver handles multiple simultaneous transfers

## Usage

```bash
# Receiver (server)
go run . -mode receiver -port 6789 -savedir ./received

# Sender (client)
go run . -mode sender -server localhost:6789 -file ./path/to/file.txt
```

## Build

```bash
go build -o transfer-agent .
```

## Docker

```bash
docker build -t transfer-agent .
docker run -p 6789:6789 transfer-agent -mode receiver
```

## Protocol

```
1. Sender connects to receiver via TCP
2. Sender sends JSON metadata (filename, size, SHA-256 checksum) terminated by newline
3. Sender streams file data in 32KB chunks
4. Receiver writes data and validates checksum
5. Receiver responds with JSON status
6. Connection closes
```

## Test

```bash
go test -v -race ./...
```

## Architecture

```
├── main.go                    # CLI entry point
├── models/
│   └── transfer.go            # TransferMetaData struct
├── service/
│   ├── interface.go           # Sender interface
│   ├── sender/sender.go       # TCP file sender
│   └── receiver/receiver.go   # TCP file receiver with concurrent handling
├── utils/
│   └── utils.go               # Checksum utilities
└── document_of_building/      # Design docs and protocol specs
```
