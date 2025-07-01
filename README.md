# Sinkzone DNS Forwarder

A simple DNS forwarder written in Go that listens on port 53 and forwards all DNS requests to Google's DNS server (8.8.8.8) while logging all domain queries to stdout.

## Features

- Listens on port 53 (UDP)
- Forwards all DNS requests to 8.8.8.8
- Logs all domain queries to stdout
- Simple and lightweight

## Requirements

- Go 1.21 or later
- Root privileges (required to bind to port 53)

## Installation

1. Clone or download this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Building

```bash
go build -o dns-forwarder main.go
```

## Running

**Important**: This program must be run with root privileges to bind to port 53.

```bash
sudo ./dns-forwarder
```

Or run directly with Go:

```bash
sudo go run main.go
```

## Usage

Once running, the DNS forwarder will:

1. Start listening on port 53
2. Forward all incoming DNS requests to 8.8.8.8
3. Log all domain queries to stdout

Example output:
```
Starting DNS forwarder on port 53
Forwarding requests to: 8.8.8.8:53
Logging all domain queries to stdout
Press Ctrl+C to stop
DNS Query: google.com (Type: A)
DNS Query: example.com (Type: AAAA)
```

## Configuration

You can modify the following constants in `main.go`:

- `upstreamDNS`: The DNS server to forward requests to (default: 8.8.8.8:53)
- `localPort`: The local port to listen on (default: :53)

## Testing

To test the DNS forwarder, you can use tools like `dig` or `nslookup`:

```bash
# Test with dig
dig @127.0.0.1 google.com

# Test with nslookup
nslookup google.com 127.0.0.1
```

## Security Note

Running a DNS server requires root privileges and should be done carefully. This is a basic implementation and may need additional security measures for production use. 