# Plivo PubSub Gateway Service

A high-performance, in-memory Pub/Sub system with WebSocket and REST API support, built with Go and designed for real-time messaging applications.

## 🚀 Features

- **Real-time Pub/Sub**: WebSocket-based messaging with fan-out delivery
- **REST API**: Complete topic management and observability endpoints
- **JWT Authentication**: Secure user authentication and authorization
- **In-Memory Storage**: Fast, low-latency message delivery
- **Message Replay**: Configurable message history with ring buffer
- **Backpressure Handling**: Drop-oldest policy for high-throughput scenarios
- **Graceful Shutdown**: Clean service termination with proper cleanup
- **Comprehensive Logging**: Structured logging with request tracing

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Environment Variables](#environment-variables)
- [API Documentation](#api-documentation)
- [WebSocket Events](#websocket-events)
- [Testing Examples](#testing-examples)
- [Architecture](#architecture)
- [Docker Deployment](#docker-deployment)
- [Assumptions & Design Decisions](#assumptions--design-decisions)

## 🚀 Quick Start

### Prerequisites

- Go 1.24.6 or later
- Docker (optional, for containerized deployment)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd plivo-pubsub-gateway
   ```

2. **Set environment variables**
   ```bash
   export JWT_SECRET_KEY="your-secret-key-here"
   export PORT="8000"
   ```

3. **Start the service**
   ```bash
   cd services/gateway
   go run main.go
   ```

4. **Verify the service is running**
   ```bash
   curl http://localhost:8000/health
   ```

## 🔧 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET_KEY` | Secret key for JWT token signing | - | ✅ Yes |
| `PORT` | HTTP server port | `8000` | ❌ No |
| `ALLOWED_CORS_ORIGIN` | CORS allowed origins (comma-separated) | `*` | ❌ No |
| `ALLOWED_CORS_METHOD` | CORS allowed methods (comma-separated) | `*` | ❌ No |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` | ❌ No |

### Example Environment Setup

```bash
# Development
export JWT_SECRET_KEY="dev-secret-key-12345"
export PORT="8000"
export LOG_LEVEL="debug"

# Production
export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export PORT="8080"
export ALLOWED_CORS_ORIGIN="https://yourdomain.com"
export LOG_LEVEL="info"
```

## 📚 API Documentation

### Health & Statistics

#### Health Check
```http
GET /health
```
**Response:**
```json
{
  "uptime_sec": 120,
  "topics": 2,
  "subscribers": 5
}
```

#### Statistics
```http
GET /stats
```
**Response:**
```json
{
  "topics": {
    "orders": {
      "messages": 42,
      "subscribers": 3
    },
    "notifications": {
      "messages": 15,
      "subscribers": 1
    }
  }
}
```

### User Management

#### Register User
```http
POST /users/register
Content-Type: application/json

{
  "username": "john_doe",
  "password": "securepassword123",
  "email": "john@example.com"
}
```

**Response:**
```json
{
  "status": "registered",
  "user": {
    "id": "abc123...",
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Login User
```http
POST /users/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "securepassword123"
}
```

#### Get User Profile
```http
GET /users/profile
Authorization: Bearer <jwt_token>
```

### Topic Management

#### Create Topic
```http
POST /topics
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "orders"
}
```

#### List Topics
```http
GET /topics
Authorization: Bearer <jwt_token>
```

#### Delete Topic
```http
DELETE /topics/{topic_name}
Authorization: Bearer <jwt_token>
```

## 🔌 WebSocket Events

### Connection

**URL:** `ws://localhost:8000/ws?token=<jwt_token>`

**Authentication:** JWT token required in query parameter

### Message Types

#### 1. Subscribe to Topic
```json
{
  "type": "subscribe",
  "topic": "orders",
  "last_n": 5,
  "request_id": "req-001"
}
```

**Response:**
```json
{
  "type": "ack",
  "request_id": "req-001",
  "topic": "orders",
  "status": "ok",
  "ts": "2024-01-15T10:30:00Z"
}
```

#### 2. Unsubscribe from Topic
```json
{
  "type": "unsubscribe",
  "topic": "orders",
  "request_id": "req-002"
}
```

#### 3. Publish Message
```json
{
  "type": "publish",
  "topic": "orders",
  "message": {
    "id": "msg-001",
    "payload": {
      "order_id": "12345",
      "status": "confirmed",
      "amount": 99.99
    }
  },
  "request_id": "req-003"
}
```

#### 4. Ping
```json
{
  "type": "ping",
  "request_id": "req-004"
}
```

**Response:**
```json
{
  "type": "pong",
  "request_id": "req-004",
  "ts": "2024-01-15T10:30:00Z"
}
```

### Event Messages

When a message is published to a topic, all subscribers receive:

```json
{
  "type": "event",
  "topic": "orders",
  "message": {
    "id": "msg-001",
    "payload": {
      "order_id": "12345",
      "status": "confirmed",
      "amount": 99.99
    },
    "topic": "orders",
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "ts": "2024-01-15T10:30:00Z"
}
```

## 🧪 Testing Examples

### 1. Complete User Flow

```bash
# 1. Register a user
curl -X POST http://localhost:8000/users/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123", "email": "test@example.com"}'

# 2. Login to get JWT token
TOKEN=$(curl -s -X POST http://localhost:8000/users/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}' | jq -r '.token')

# 3. Create a topic
curl -X POST http://localhost:8000/topics \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-topic"}'

# 4. Connect to WebSocket
wscat -c "ws://localhost:8000/ws?token=$TOKEN"
```

### 2. WebSocket Testing with JavaScript

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8000/ws?token=YOUR_JWT_TOKEN');

ws.onopen = function() {
    console.log('Connected!');
    
    // Subscribe to topic
    ws.send(JSON.stringify({
        type: 'subscribe',
        topic: 'test-topic',
        last_n: 5,
        request_id: 'test-001'
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
    
    if (data.type === 'ack') {
        // Publish a message
        ws.send(JSON.stringify({
            type: 'publish',
            topic: 'test-topic',
            message: {
                id: 'msg-001',
                payload: { text: 'Hello World!' }
            },
            request_id: 'test-002'
        }));
    }
};
```

### 3. Load Testing

```bash
# Create multiple topics
for i in {1..10}; do
  curl -X POST http://localhost:8000/topics \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"topic-$i\"}"
done

# Check statistics
curl http://localhost:8000/stats
```

## 🏗️ Architecture

### System Overview

The Plivo PubSub Gateway is built using a modular, microservice-oriented architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Plivo PubSub Gateway                     │
├─────────────────────────────────────────────────────────────┤
│  HTTP Server (Gin)                                          │
│  ├── CORS Middleware                                        │
│  ├── Auth Middleware (JWT)                                  │
│  └── Route Handlers                                         │
├─────────────────────────────────────────────────────────────┤
│  Gateway Modules                                            │
│  ├── User Module (Auth & Profile)                           │
│  ├── Topic Module (REST API)                                │
│  └── WebSocket Module (Real-time Messaging)                 │
├─────────────────────────────────────────────────────────────┤
│  Core PubSub Engine                                         │
│  ├── Topic Management                                       │
│  ├── Subscription Handling                                  │
│  ├── Message Broadcasting                                   │
│  └── Ring Buffer (Message Replay)                           │
├─────────────────────────────────────────────────────────────┤
│  Storage Layer                                              │
│  ├── In-Memory Topic Store                                  │
│  ├── In-Memory User Store                                   │
│  └── In-Memory Subscription Store                           │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. PubSub Engine (`pubsub/`)
- **Singleton Service**: Global pub/sub instance
- **Topic Management**: Create, delete, list topics
- **Subscription Handling**: Subscribe/unsubscribe with message replay
- **Message Broadcasting**: Fan-out delivery to all subscribers
- **Ring Buffer**: Configurable message history (default: 100 messages)

#### 2. User Module (`user/`)
- **Authentication**: JWT-based user authentication
- **User Management**: Register, login, profile management
- **Password Security**: bcrypt hashing with salt
- **In-Memory Storage**: Thread-safe user data storage

#### 3. Topic Module (`topic/`)
- **REST API**: Topic CRUD operations
- **Observability**: Health checks and statistics
- **Authentication**: JWT-protected endpoints
- **PubSub Integration**: Wrapper around core pub/sub engine

#### 4. WebSocket Module (`websocket/`)
- **Real-time Communication**: WebSocket upgrade and handling
- **JWT Authentication**: Token validation via query parameter
- **Message Processing**: Subscribe, unsubscribe, publish, ping
- **Event Broadcasting**: Real-time message delivery to subscribers

### Data Flow

#### Message Publishing Flow
```
1. Client → WebSocket → JWT Validation
2. WebSocket → PubSub Engine → Topic Lookup
3. PubSub Engine → Ring Buffer → Message Storage
4. PubSub Engine → All Subscribers → Message Delivery
5. Subscribers → WebSocket → Client Response
```

#### Topic Management Flow
```
1. Client → REST API → JWT Validation
2. REST API → Topic Module → PubSub Engine
3. PubSub Engine → Topic Store → CRUD Operations
4. PubSub Engine → Response → REST API → Client
```

### Concurrency & Thread Safety

- **Mutex Protection**: All shared data structures protected with `sync.RWMutex`
- **Goroutine Management**: Controlled goroutine spawning for message delivery
- **Channel Communication**: Buffered channels for message queuing
- **Context Propagation**: Request context passed through all layers

### Performance Characteristics

- **Memory Usage**: O(n) where n = number of topics + subscribers + messages
- **Message Latency**: Sub-millisecond for in-memory operations
- **Throughput**: Limited by Go's goroutine scheduler and memory bandwidth
- **Scalability**: Single-instance, designed for moderate scale (1000s of concurrent connections)

## 🐳 Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.24.6-alpine AS builder

WORKDIR /app
COPY . .

# Build the application
RUN cd services/gateway && go build -o pubsub-gateway .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/services/gateway/pubsub-gateway .

# Expose port
EXPOSE 8000

# Set environment variables
ENV PORT=8000
ENV JWT_SECRET_KEY=""

# Run the application
CMD ["./pubsub-gateway"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  pubsub-gateway:
    build: .
    ports:
      - "8000:8000"
    environment:
      - JWT_SECRET_KEY=${JWT_SECRET_KEY:-default-secret-key}
      - PORT=8000
      - LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Docker Run Commands

```bash
# Build the image
docker build -t plivo-pubsub-gateway .

# Run with environment variables
docker run -d \
  --name pubsub-gateway \
  -p 8000:8000 \
  -e JWT_SECRET_KEY="your-secret-key-here" \
  -e PORT=8000 \
  plivo-pubsub-gateway

# Run with Docker Compose
docker-compose up -d

# Check logs
docker logs pubsub-gateway

# Health check
curl http://localhost:8000/health
```

## 📋 Assumptions & Design Decisions

### Backpressure Policy
- **Drop Oldest**: When ring buffer is full, oldest messages are dropped
- **Buffer Size**: Configurable ring buffer (default: 100 messages per topic)
- **Channel Buffer**: 100 message buffer per subscriber to handle burst traffic
- **Rationale**: Prevents memory exhaustion while maintaining recent message history

### Authentication Strategy
- **JWT Tokens**: Stateless authentication for scalability
- **WebSocket Auth**: Token passed via query parameter (industry standard)
- **Password Security**: bcrypt hashing with default cost (10 rounds)
- **Token Expiry**: 24-hour expiration for security

### In-Memory Storage
- **No Persistence**: All data lost on service restart
- **Fast Access**: Sub-millisecond read/write operations
- **Memory Bounds**: Ring buffer limits prevent unbounded growth
- **Use Case**: Real-time messaging, not long-term data storage

### Message Delivery
- **Fan-out**: Every subscriber receives each message once
- **At-least-once**: Best-effort delivery (no acknowledgments)
- **Ordering**: Messages delivered in publish order per topic
- **Isolation**: No cross-topic message leakage

### Error Handling
- **Graceful Degradation**: Service continues operating despite individual failures
- **Comprehensive Logging**: Structured logging for debugging and monitoring
- **HTTP Status Codes**: Standard REST API error responses
- **WebSocket Errors**: JSON error messages with error codes

### Scalability Considerations
- **Single Instance**: Designed for moderate scale (1000s of connections)
- **Horizontal Scaling**: Multiple instances can run independently
- **Load Balancing**: Stateless design supports load balancer distribution
- **Resource Limits**: Memory usage bounded by configuration

### Security Assumptions
- **Network Security**: Assumes secure network (HTTPS/WSS in production)
- **Token Security**: JWT secret must be kept secure
- **CORS Policy**: Configurable for production deployment
- **Input Validation**: All inputs validated and sanitized

## 🔧 Development

### Project Structure
```
plivo-pubsub-gateway/
├── libraries/
│   ├── auth/           # JWT authentication library
│   └── pagination/     # Pagination utilities
├── logging/            # Structured logging
├── pubsub/             # Core pub/sub engine
├── services/
│   └── gateway/        # Main gateway service
│       ├── app/        # Application setup
│       ├── middlewares/# HTTP middlewares
│       ├── secure/     # Route security
│       ├── user/       # User management
│       ├── topic/      # Topic management
│       └── websocket/  # WebSocket handling
└── README.md
```

### Building from Source
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
cd services/gateway
go build -o pubsub-gateway .

# Run with race detection
go run -race main.go
```

## 📞 Support

For questions or issues, please refer to the code documentation or create an issue in the repository.

---

**Built with ❤️ using Go, Gin, and WebSockets**
