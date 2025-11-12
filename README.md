# DIY CCTV

A WebRTC-based CCTV streaming system that allows you to connect multiple cameras and view streams through a web interface.

## Overview

DIY CCTV is a real-time video streaming application built with Go, WebRTC, and WebSockets. It enables you to:

- Connect cameras as CCTV clients to a central server
- Monitor multiple camera streams simultaneously through an admin dashboard
- View detailed stream information and statistics for individual cameras

### Application Pages

The application consists of three main pages:

#### 1. CCTV Client Page (`/`)

The client interface allows cameras to connect and stream video to the server.

- **Start Camera** button to begin streaming
- **Stop** button to end the stream
- Real-time video preview
- Connection status display ("Sent offer, waiting for answer...")
- Uses WebRTC peer connection for video streaming

#### 2. Admin Page (`/admin`)

A dashboard displaying all connected CCTV cameras in a grid layout.

- Shows live thumbnails of all active camera streams
- Click on any thumbnail to view detailed stream information
- Automatically updates when cameras connect/disconnect
- Grid layout (3 cameras per row)
- Each thumbnail links to the detailed stream page

#### 3. Stream Page (`/stream/{id}`)

Detailed view of an individual camera stream with comprehensive statistics.

- **Full-screen video display**
- **Connection Information:**
  - Status (Connected/Disconnected)
  - Stream ID
  - Connection Type (UDP/TCP)
- **Network Statistics:**
  - Bitrate (kbps)
  - Packets Lost
  - Round Trip Time (ms)
  - Jitter (ms)
- **Transfer Data:**
  - Data Received (MB)
  - Frames Received
  - Frames Dropped
- **Video Quality:**
  - Resolution (e.g., 640x480)
  - Frame Rate (fps)
  - Codec (VP8/VP9/H264)
- Back button to return to admin page

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)

## Quick Start

### Using Docker Compose (Recommended)

1. **Create environment file:**

   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` file with your configuration:**

   ```bash
   # Server Configuration
   SERVER_PORT=8081

   # Client Configuration
   CLIENT_PORT=8080
   SERVER_BASE_URL=http://server:8081
   WEB_SOCKET_URL=ws://localhost:8081/ws
   ```

3. **Start all services:**

   ```bash
   docker-compose up -d
   ```

4. **View logs:**

   ```bash
   # All services
   docker-compose logs -f

   # Specific service
   docker-compose logs -f client
   docker-compose logs -f server
   ```

5. **Stop all services:**
   ```bash
   docker-compose down
   ```

### Access the Application

- **CCTV Client:** http://localhost:8080 - Connect a camera to start streaming
- **Admin Dashboard:** http://localhost:8080/admin - View all connected cameras
- **Stream Details:** http://localhost:8080/stream/{id} - View specific camera stream
- **Server WebSocket:** ws://localhost:8081/ws
- **Server API:** http://localhost:8081/api/live - Get list of active streams

## Usage Guide

### Setting Up a Camera Stream

1. **Open CCTV Client page** at http://localhost:8080
2. Click **"Start Camera"** button
3. Allow browser to access your camera when prompted
4. Wait for "Sent offer, waiting for answer..." status
5. Once connected, your camera feed will be visible in the preview

### Monitoring All Cameras

1. **Open Admin page** at http://localhost:8080/admin
2. View grid of all active camera streams
3. Thumbnails show live preview of each camera
4. Click any thumbnail to view detailed stream information

### Viewing Stream Details

1. From admin page, **click on any camera thumbnail**
2. Or directly access: http://localhost:8080/stream/{stream-id}
3. View comprehensive statistics:
   - Real-time bitrate and network performance
   - Frame statistics and dropped frames
   - Video quality metrics (resolution, fps, codec)
4. Click **"← Back"** to return to admin dashboard

## Building Individual Services

### Build Client Only

```bash
docker build -f Dockerfile.client -t diycctv-client .
docker run -p 8080:8080 \
  -e CLIENT_PORT=8080 \
  -e SERVER_BASE_URL=http://localhost:8081 \
  -e WEB_SOCKET_URL=ws://localhost:8081/ws \
  diycctv-client
```

### Build Server Only

```bash
docker build -f Dockerfile.server -t diycctv-server .
docker run -p 8081:8081 \
  -e SERVER_PORT=8081 \
  diycctv-server
```

## Development

### Rebuild after code changes

```bash
# Rebuild all services
docker-compose up -d --build

# Rebuild specific service
docker-compose up -d --build client
docker-compose up -d --build server
```

### Access container shell

```bash
# Client container
docker-compose exec client sh

# Server container
docker-compose exec server sh
```

## Production Deployment

For production, consider:

1. **Use specific Go version in Dockerfile** (update `golang:1.21-alpine` to your required version)
2. **Set resource limits in docker-compose.yml:**

   ```yaml
   services:
     server:
       deploy:
         resources:
           limits:
             cpus: "1"
             memory: 512M
   ```

3. **Use Docker secrets for sensitive data**
4. **Enable health checks:**
   ```yaml
   healthcheck:
     test: ["CMD", "wget", "-q", "--spider", "http://localhost:8081/api/live"]
     interval: 30s
     timeout: 10s
     retries: 3
   ```

## Troubleshooting

### Port conflicts

If ports 8080 or 8081 are already in use, change them in `.env`:

```bash
CLIENT_PORT=9080
SERVER_PORT=9081
```

### WebSocket connection issues

Ensure `WEB_SOCKET_URL` in `.env` matches your server's accessible address:

- Local: `ws://localhost:8081/ws`
- Network: `ws://YOUR_IP:8081/ws`
- Production: `wss://your-domain.com/ws`

### Container logs

```bash
docker-compose logs --tail=100 -f
```

### Clean rebuild

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

## Architecture

```
┌──────────────────────┐
│  CCTV Client (/)     │  ← Camera connects and streams
│  - Start Camera      │
│  - Video Preview     │
│  - WebRTC Connection │
└──────────┬───────────┘
           │ WebSocket + WebRTC
           ▼
┌──────────────────────┐
│  Server (8081)       │  ← Central server
│  - WebSocket (/ws)   │
│  - REST API          │
│  - RTC Management    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Client Web (8080)   │
│  ┌────────────────┐  │
│  │ Admin (/admin) │  │  ← View all streams
│  │ - Grid Layout  │  │
│  │ - Thumbnails   │  │
│  └────────┬───────┘  │
│           │          │
│  ┌────────▼────────┐ │
│  │ Stream (/stream)│ │  ← View single stream
│  │ - Full Video    │ │
│  │ - Statistics    │ │
│  └─────────────────┘ │
└──────────────────────┘
```

### Data Flow

1. **CCTV Client** opens camera and creates WebRTC offer
2. **Server** receives offer via WebSocket, creates peer connection
3. **Server** sends answer back to CCTV client
4. **WebRTC** establishes direct media connection
5. **Admin Page** fetches list of active streams from server API
6. **Stream Page** connects to specific stream via WebSocket/WebRTC

## Environment Variables

| Variable          | Description           | Default                |
| ----------------- | --------------------- | ---------------------- |
| `CLIENT_PORT`     | Client service port   | 8080                   |
| `SERVER_PORT`     | Server service port   | 8081                   |
| `SERVER_BASE_URL` | Server URL for client | http://localhost:8081  |
| `WEB_SOCKET_URL`  | WebSocket endpoint    | ws://localhost:8081/ws |
