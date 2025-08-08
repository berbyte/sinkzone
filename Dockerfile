# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sinkzone .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/sinkzone .

# Expose ports for DNS resolver and API
EXPOSE 5353 8080

# Run the resolver with API and DNS ports
CMD ["./sinkzone", "resolver", "--api-port", "8080", "--port", "5353"] 