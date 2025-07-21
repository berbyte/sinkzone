# Sinkzone

A lightweight DNS-based productivity tool that helps you stay focused by blocking distracting websites during focus sessions.

## Features

- **DNS Resolver**: Local DNS server that forwards requests to upstream nameservers
- **Focus Mode**: Block all domains except those in your allowlist
- **Real-time Monitoring**: Beautiful TUI for monitoring DNS traffic and managing allowlist
- **Interactive Forms**: Huh-based forms for focus mode configuration
- **Enhanced Styling**: Lipgloss-powered beautiful colors and layout
- **SQLite Database**: Persistent storage for DNS queries and allowlist management
- **Configurable**: Easy configuration via YAML files
- **Timeout Support**: Set focus sessions with automatic expiration
- **Lightweight**: Minimal dependencies, simple CLI interface

## Architecture

Sinkzone uses a **shared SQLite database** to enable communication between the root DNS resolver and unprivileged monitor process:

- **DNS Resolver** (runs as root): Records all DNS queries to SQLite database
- **Monitor** (runs as user): Reads DNS stats and manages allowlist via TUI
- **Database**: Shared SQLite file with WAL mode for concurrent access

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd sinkzone

# Build the binary
go build -o sinkzone .

# Install (optional)
sudo cp sinkzone /usr/local/bin/
```

## Quick Start

1. **Start the TUI interface**:
   ```bash
   ./sinkzone
   ```

2. **Start the DNS resolver** (requires root):
   ```bash
   sudo ./sinkzone resolver
   ```

3. **Configure your system to use the local DNS**:
   - Set your DNS server to `127.0.0.1`
   - Or use the resolver as a forwarder

4. **Switch to focus mode**:
   ```bash
   ./sinkzone focus 1h  # Focus for 1 hour
   ./sinkzone focus 30m # Focus for 30 minutes
   ```

## Commands

### `sinkzone` (default)
Start the TUI interface with tabs for monitoring and configuration.

**Features:**
- **Monitor Tab**: Real-time DNS traffic statistics with allowlist management
- **Config Tab**: View current configuration and focus mode status
- **Full Terminal Usage**: Automatically adjusts to terminal size
- **Tab Navigation**: Use Tab key or 1/2 keys to switch between tabs

**Controls:**
- `Tab` or `1/2`: Switch between Monitor and Config tabs
- `↑/↓`: Navigate through domains (Monitor tab)
- `a`: Add selected domain to allowlist (Monitor tab)
- `r`: Remove selected domain from allowlist (Monitor tab)
- `f`: Show focus mode form (Config tab)
- `n`: Switch to normal mode (Config tab)
- `Enter`: Submit form
- `Esc`: Cancel form
- `q`: Quit

### `sinkzone resolver`
Start the DNS resolver service. Must be run as root to bind to port 53.

**Features:**
- Records all DNS queries to SQLite database
- Applies focus mode rules based on allowlist
- Forwards requests to upstream nameservers

### `sinkzone focus <duration>`
Switch to focus mode for the specified duration. Duration can be specified as:
- `1h` - 1 hour
- `30m` - 30 minutes
- `2h30m` - 2 hours 30 minutes

### `sinkzone help`
Show help information and available commands.

## Configuration

### Main Configuration (`~/.sinkzone/sinkzone.yaml`)

```yaml
mode: normal  # or "focus"
focus_end_time: null  # automatically set when entering focus mode
upstream_nameservers:
  - "8.8.8.8:53"
  - "1.1.1.1:53"
```

### Database (`~/.sinkzone/sinkzone.db`)

The SQLite database contains:
- **dns_queries**: All DNS queries with timestamps and blocked status
- **allowlist**: Active domains that are allowed during focus mode

## TUI Interface

### Monitor Tab
- Real-time DNS traffic statistics
- Domain query counts and last seen times
- Blocked vs allowed status
- Add/remove domains from allowlist
- Sort by query count

### Config Tab
- Current mode and focus status
- Upstream nameservers configuration
- Allowlist contents
- Interactive focus mode form
- Normal mode toggle

## How It Works

1. **Normal Mode**: All DNS requests are forwarded to upstream nameservers without blocking.

2. **Focus Mode**: 
   - Only domains in the allowlist are resolved
   - All other domains return NXDOMAIN (domain not found)
   - Automatically expires after the specified duration

3. **DNS Server**: 
   - Listens on port 53 (requires root)
   - Records all queries to SQLite database
   - Forwards requests to configured upstream nameservers
   - Applies focus mode rules based on database allowlist

4. **TUI Interface**: 
   - Reads DNS statistics from SQLite database
   - Provides tabs for monitoring and configuration
   - Updates are immediately reflected in DNS resolver

## System Integration

### macOS
```bash
# Set DNS server to localhost
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

### Linux
```bash
# Edit /etc/resolv.conf
nameserver 127.0.0.1
```

### Windows
- Network Settings → Change adapter options → Properties → Internet Protocol Version 4 → Properties → Use the following DNS server addresses: 127.0.0.1

## Development

### Building
```bash
go build -o sinkzone .
```

### Running Tests
```bash
go test ./...
```

### Project Structure
```
sinkzone/
├── main.go              # Entry point
├── cmd/                 # CLI commands
│   ├── root.go         # Root command (starts TUI)
│   ├── resolver.go     # DNS resolver command
│   ├── focus.go        # Focus mode command
│   └── help.go         # Help command
├── internal/
│   ├── config/         # Configuration management
│   ├── database/       # SQLite database operations
│   ├── dns/           # DNS server implementation
│   └── monitor/       # TUI monitoring interface
└── templates/          # HTML templates (legacy)
```

## Dependencies

- `github.com/miekg/dns` - DNS server implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/huh` - Interactive forms
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/mattn/go-sqlite3` - SQLite driver
- `gopkg.in/yaml.v3` - YAML configuration parsing

## License

[Add your license here]

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request 