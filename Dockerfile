# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o sinkzone main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

# Create sinkzone user
RUN addgroup -g 1000 sinkzone && \
    adduser -D -s /bin/sh -u 1000 -G sinkzone sinkzone

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sinkzone .

# Create data directory
RUN mkdir -p /home/sinkzone/.sinkzone && \
    chown -R sinkzone:sinkzone /home/sinkzone/.sinkzone

# Switch to sinkzone user
USER sinkzone

# Expose DNS port
EXPOSE 53/udp 53/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD nslookup google.com 127.0.0.1 || exit 1

# Default command
ENTRYPOINT ["./sinkzone"]
CMD ["dns", "start"] 