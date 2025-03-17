# help-request-teams-notifier

This microservice integrates with the BYU OIT AV event system to listen for help requests from rooms and forward them to Microsoft Teams channels using the Teams webhook API.

## Overview

The service subscribes to the central event system and listens for specific "help-request" events. When a help request is detected, it formats a Teams message card notification and sends it to the configured webhook URL.

When running in debug mode, the service will continue to operate even if the event hub connection fails, allowing developers to work with the HTTP API endpoints. In non-debug mode, the service will exit if it cannot connect to the event hub.

## Configuration

The service can be configured using either command-line flags or environment variables:

### Command-line flags:
- `--log-level` or `-L`: Sets the logging level (default: "info")
- `--hub-address`: Address of the central event system hub
- `--webhook-url`: URL of the Microsoft Teams webhook
- `--monitoring-url`: URL of the AV Monitoring Service (used in message links)
- `--port` or `-P`: Port for the HTTP server (default: "8080")

### Environment variables:
- `LOG_LEVEL`: Sets the logging level (overrides the flag if set)
- `EVENT_HUB_ADDRESS`: Address of the event hub (overrides the flag if set)
- `TEAMS_WEBHOOK_URL`: URL of the Microsoft Teams webhook (overrides the flag if set)
- `TEAMS_MONITORING_URL`: URL of the AV Monitoring Service (overrides the flag if set)

When running on Windows, the log level automatically defaults to "debug" for easier development.

## API Endpoints

### Health Check Endpoints
- `GET /ping` - Returns "pong" if the service is running
- `GET /status` - Returns the service status
- `GET /healthz` - Returns "healthy"

### Log Management Endpoints
- `GET /logz` - Returns the current log level as plain text
- `GET /logz/:log_level` - Set the log level (debug, info, warn, error)

### API (v1)
- `GET /api/v1/health` - Returns detailed health status
- `GET /api/v1/notify/building/:building/room/:room/device/:device` - Manually trigger a notification with path parameters:
  ```
  /api/v1/notify/building/ITB/room/1101/device/CP1
  ```
  - Where:
    - `:building` - Building code (e.g., "ITB")
    - `:room` - Room number (e.g., "1101")  
    - `:device` - Device identifier (e.g., "CP1")
- `GET /api/v1/notify` - (Legacy) Manually trigger a notification with query parameters
- `POST /api/v1/notify` - (Legacy) Manually trigger a notification with JSON payload
- `GET /api/v1/config` - Returns current configuration (sensitive values are masked)

## Message Format

The Teams notification includes:
- Building name
- Room number
- Device ID
- Timestamp
- A link to view the room in the monitoring system

## Usage

```bash
./help-request-teams-notifier \
  --hub-address="localhost:7100" \
  --webhook-url="https://outlook.office.com/webhook/..." \
  --monitoring-url="https://monitoring.example.com" \
  --port="8080"
```

## Building

To build the service:

```bash
go build -o help-request-teams-notifier
```

The service uses only standard Go libraries for HTTP communication, with no external notification dependencies.