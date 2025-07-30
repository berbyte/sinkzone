FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sinkzone .

# Create final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S sinkzone && \
    adduser -u 1001 -S sinkzone -G sinkzone

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sinkzone .

# Create config directory
RUN mkdir -p /app/.sinkzone && \
    chown -R sinkzone:sinkzone /app

# Switch to non-root user
USER sinkzone

# Expose DNS port
EXPOSE 53/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nslookup google.com 127.0.0.1:53 || exit 1

# Run the DNS resolver
CMD ["./sinkzone", "resolver"] 