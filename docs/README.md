# Sinkzone Documentation

This directory contains additional documentation for the Sinkzone project.

## Architecture

Sinkzone follows a modular architecture with clear separation of concerns:

```
sinkzone/
├── cmd/           # CLI commands and user interface
├── config/        # Configuration management
├── database/      # Data persistence layer
├── dns/           # DNS server implementation
├── internal/      # Private application logic
├── pkg/           # Public libraries
└── docs/          # Documentation
```

## Package Structure

### cmd/
Contains all CLI commands using the Cobra framework:
- `root.go` - Main CLI structure
- `dns.go` - DNS server management
- `mode.go` - Filtering mode management
- `domains.go` - Domain rule management
- `stats.go` - Statistics and reporting
- `reset.go` - Data reset functionality
- `web.go` - Web dashboard (future)

### config/
Handles application configuration:
- TOML-based configuration files
- Default configuration generation
- PIN hashing and verification

### database/
SQLite database operations:
- Domain rules storage
- DNS query logging
- Application state management
- Statistics aggregation

### dns/
DNS server implementation:
- DNS request handling
- Upstream forwarding
- Query logging
- Response filtering

### internal/
Private application logic:
- `app/` - Core application lifecycle
- Business logic not meant for external use

### pkg/
Public libraries that could be used by other applications:
- `filter/` - Domain filtering engine

## Development Guidelines

### Adding New Commands

1. Create a new file in `cmd/`
2. Define the command using Cobra
3. Add it to the root command in `cmd/root.go`
4. Add tests in `cmd/` directory

### Adding New Features

1. Place business logic in `internal/` or `pkg/`
2. Add configuration options in `config/`
3. Update database schema if needed
4. Add CLI commands in `cmd/`

### Testing

- Unit tests should be in the same package as the code
- Integration tests can be in a separate `tests/` directory
- Use table-driven tests for better coverage

### Error Handling

- Use wrapped errors with context
- Return meaningful error messages
- Log errors appropriately
- Handle errors at the appropriate level

## API Design

### Configuration API

```go
// Load configuration
cfg, err := config.LoadConfig()

// Save configuration
err := config.SaveConfig(cfg)

// Verify PIN
valid := config.VerifyPIN(pin, hashedPIN)
```

### Database API

```go
// Open database
db, err := database.OpenDB(path)

// Add domain rule
err := db.AddDomainRule(domain, action)

// Get statistics
stats, err := db.GetStats()
```

### DNS API

```go
// Create server
server := dns.NewServer(upstream, db)

// Start server
err := server.Start()

// Stop server
err := server.Stop()
```

## Security Considerations

1. **Root Privileges**: Required for port 53 binding
2. **PIN Protection**: SHA1-hashed PINs for lockdown mode
3. **Local Storage**: All data stored locally
4. **Input Validation**: Validate all user inputs
5. **Error Handling**: Don't expose sensitive information in errors

## Performance Considerations

1. **Database Indexing**: Proper indexes on frequently queried columns
2. **Connection Pooling**: Reuse database connections
3. **Memory Usage**: Monitor memory usage for large query logs
4. **DNS Caching**: Consider implementing DNS response caching

## Deployment

### System Requirements

- Go 1.21 or later
- SQLite3
- Root privileges (for DNS server)
- 50MB disk space
- 10MB RAM

### Supported Platforms

- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

### Installation Methods

1. **Homebrew**: `brew install berbyte/tap/sinkzone`
2. **Linux**: `curl -sSL https://sinkzone.ber.run/install.sh | sudo bash`
3. **Docker**: `docker run --net=host ghcr.io/berbyte/sinkzone:latest`
4. **Manual**: Download binary from GitHub releases 