# RTSP Streaming System - Technical Documentation

## Project Overview

This system efficiently manages real-time streaming from thousands of IP cameras connected to various indicators. The system activates camera streams on-demand and automatically shuts them down when users stop viewing to minimize resource usage.

**Status: ✅ Go Stream Manager Implementation Complete**

## System Architecture

```
+------------+         +------------------+         +----------------+
|  Browser   | <-----> | Laravel Backend  | <-----> | Stream Manager |
+------------+         +------------------+         +----------------+
                                |                             |
                                |                             v
                                |                      [ffmpeg Process]
                                |                             |
                                v                             v
                          [Nginx Web Server] ------>  HLS files (.m3u8 / .ts)
```

## Core Components

### ✅ 1. Laravel Backend
- Handles user requests and authentication
- Manages camera-to-indicator mapping with RTSP links
- Communicates with Stream Manager for on-demand streaming
- Returns stream URLs to frontend

### ✅ 2. Stream Manager (Go Microservice) - **IMPLEMENTED**
- **Port:** 8088 (configurable)
- Checks for existing active ffmpeg processes
- Converts RTSP streams to HLS format using ffmpeg
- Automatically kills processes after inactivity timeout
- Concurrent stream management with mutex protection

### ✅ 3. Media Server (Nginx)
- Serves .m3u8 and .ts files via HTTP
- CORS enabled for browser access
- Caching disabled for real-time streaming

## Streaming Workflow

1. **User Action**: User navigates to indicator page and clicks "view camera"
2. **Laravel Processing**: Retrieves RTSP link for the indicator
3. **Stream Request**: Laravel sends GET request to Stream Manager:
   ```
   GET /api/stream/start?indicator_id=123&rtsp_link=rtsp://example.com/stream
   ```
4. **Stream Manager Response**:
   - If process active: returns existing stream path `/streams/123/stream.m3u8`
   - If inactive: starts new ffmpeg process and generates .m3u8 file
5. **Laravel Response**: Returns stream URL to frontend
6. **Browser Playback**: Connects to stream using HLS.js
7. **Auto Cleanup**: Stream Manager stops process after 30s of inactivity

## Technical Implementation Details

### RTSP → HLS Conversion
```bash
ffmpeg -rtsp_transport tcp -i [RTSP_LINK] \
    -c:v libx264 -preset ultrafast -f hls \
    -hls_time 2 -hls_list_size 5 -hls_flags delete_segments \
    -y /streams/{indicator_id}/stream.m3u8
```

### Stream Files Structure
```
/streams/{indicator_id}/
├── stream.m3u8
├── segment0.ts
├── segment1.ts
└── ...
```

### Stream Manager Logic (Implemented)
```go
func (m *Manager) StartStream(indicatorID, rtspLink string) (string, error) {
    if existingStream := m.getActiveStream(indicatorID); existingStream != nil {
        existingStream.LastAccess = time.Now()
        return existingStream.StreamPath, nil
    }
    return m.createNewStream(indicatorID, rtspLink)
}
```

### Auto-Cleanup System
- **Cleanup Worker**: Runs every 10 seconds
- **Inactivity Timeout**: 30 seconds (configurable)
- **Actions**: Kill ffmpeg process → Remove stream directory → Update internal state

## API Endpoints (Implemented)

### Start/Get Stream
```http
GET /api/stream/start?indicator_id=123&rtsp_link=rtsp://example.com/stream
```
**Response:**
```json
{
  "stream_path": "/streams/123/stream.m3u8",
  "status": "started"
}
```

### Stop Stream
```http
POST /api/stream/stop
Content-Type: application/json

{"indicator_id": "123"}
```

### Check Stream Status
```http
GET /api/stream/status?indicator_id=123
```

### Health Check
```http
GET /api/health
```
**Response:**
```json
{
  "status": "healthy",
  "active_streams": 5
}
```

## Test Scenarios

| Scenario | Expected Result |
|----------|----------------|
| User starts viewing stream | Video appears in 3-5 seconds |
| User closes page | Process stops after 30 seconds |
| 10 parallel users view same camera | 1 ffmpeg process, shared .m3u8 file |
| Invalid RTSP link | Error response with details |
| Check health endpoint | Returns active stream count |

## Installation & Usage

### Quick Start
```bash
# Clone and setup
git clone <repository>
cd go-rtsp
cp .env.example .env

# Run with Go
go mod tidy
go run main.go

# Or use Docker
docker-compose up -d
```

### Configuration
Environment variables in `.env`:
```env
PORT=8088
STREAM_DIR=./streams
CLEANUP_DELAY_SECONDS=30
HLS_SEGMENT_TIME=2
HLS_LIST_SIZE=5
```

### Project Structure
```
go-rtsp/
├── main.go                 # Entry point
├── internal/
│   ├── server/            # HTTP server & handlers
│   ├── stream/            # Stream management logic
│   └── config/            # Configuration
├── docker-compose.yml     # Docker setup with Nginx
├── Dockerfile            # Go app container
└── nginx.conf            # Nginx config for HLS serving
```

## Scalability & Performance

- **Go Concurrency**: Handles thousands of parallel processes
- **Memory Efficient**: Minimal RAM and CPU usage
- **Mutex Protection**: Thread-safe stream management
- **Resource Cleanup**: Automatic cleanup prevents memory leaks
- **Future Extensions**: Easy to add WebRTC, Snapshot, MJPEG support

## Security Features

- **Access Control**: Stream access limited to authorized users
- **CORS Enabled**: Proper browser integration
- **Process Isolation**: Each stream runs in separate ffmpeg process
- **Auto Cleanup**: Prevents resource exhaustion

## Implementation Status

### ✅ Completed
- [x] Go Stream Manager microservice
- [x] RTSP to HLS conversion
- [x] Concurrent stream management
- [x] Auto-cleanup system
- [x] REST API endpoints
- [x] Docker containerization
- [x] Nginx configuration

### 🔄 Next Steps
1. **Laravel Integration**: Update Laravel backend to use new API
2. **Frontend Integration**: HLS.js video player implementation
3. **Authentication**: Token-based stream access
4. **Monitoring**: Metrics and logging
5. **Load Testing**: Performance optimization