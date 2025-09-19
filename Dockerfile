# Build stage
FROM golang:1.24.6-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy all source code first
COPY . ./

# Download dependencies for the gateway service
WORKDIR /app/services/gateway
RUN go mod download

# Build the application
WORKDIR /app/services/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pubsub-gateway .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/services/gateway/pubsub-gateway .

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8000

# Set environment variables
ENV PORT=8000
ENV JWT_SECRET_KEY=""
ENV LOG_LEVEL="info"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8000/health || exit 1

# Run the application
CMD ["./pubsub-gateway"]
