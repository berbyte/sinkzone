# Sinkzone

![Sinkzone Splash](https://share.ber.sh/sinkzone-splash.png)

> **A DNS-based productivity tool that blocks distracting websites during focus sessions.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)](https://github.com/berbyte/sinkzone/releases)

## What is Sinkzone?

Sinkzone is a DNS-based productivity tool that helps you stay focused by blocking distracting websites at the DNS level. When you enable focus mode, only domains in your allowlist will resolve - everything else returns NXDOMAIN.

**Key Features:**
- 🔒 **DNS-level blocking** - Blocks at the network level, not just browser
- 🎯 **Focus mode** - Temporarily block non-allowlisted domains with auto-expiration
- 📊 **Real-time monitoring** - Beautiful TUI to monitor DNS traffic and manage allowlist

## Quick Start

### 1. Install

**Homebrew (Recommended):**
```bash
brew tap berbyte/sinkzone
brew install sinkzone
```

**Manual:**
```bash
# Download from releases or build from source
go build -o sinkzone .
```

### 2. Start the DNS Resolver

The DNS resolver needs root privileges to bind to port 53:

```bash
sudo sinkzone resolver
```

### 3. Configure System DNS

Set your system's DNS server to `127.0.0.1`:

**macOS:**
```bash
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

**Linux:**
```bash
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

**Windows:**
- Network Settings → Change adapter options → Properties → Internet Protocol Version 4 → Properties → Use the following DNS server addresses: 127.0.0.1

### 4. Open the TUI

In another terminal:

```bash
sinkzone
```

### 5. Enable Focus Mode

- Press `f` in the TUI, or
- Run: `sinkzone focus 1h`


## Usage

### Commands

| Command | Description |
|---------|-------------|
| `sinkzone` | Start the TUI interface |
| `sudo sinkzone resolver` | Start the DNS resolver (requires root) |
| `sinkzone focus 1h` | Enable focus mode for 1 hour |
| `sinkzone status` | Check focus mode status |
| `sinkzone disable-focus` | Disable focus mode immediately |

### TUI Navigation

- **Tabs**: Use `←`/`→` to switch between tabs
- **Focus Mode**: Press `f` to enable focus mode for 1 hour
- **Monitoring**: View and manage blocked/allowed domains
- **Settings**: Configure upstream DNS resolvers
- **Quit**: Press `q` to exit

## How It Works

### Normal Mode
- All DNS requests are forwarded to upstream nameservers
- Real-time monitoring of DNS traffic
- Allowlist management

### Focus Mode
- Only domains in the allowlist are resolved
- All other domains return `NXDOMAIN` (domain not found)
- Automatically expires after the specified duration
- Red banner indicator in TUI

## Service Management

### Linux (systemd)
```bash
sudo systemctl enable sinkzone-resolver
sudo systemctl start sinkzone-resolver
```

### macOS (launchd)
```bash
sudo launchctl load /Library/LaunchDaemons/com.berbyte.sinkzone.resolver.plist
```

## Configuration

Configuration files are stored in `~/.sinkzone/`:

- `sinkzone.yaml` - Main configuration
- `sinkzone.db` - SQLite database for queries and allowlist
- `state.json` - Focus mode state

## Development

```bash
# Build
go build -o sinkzone .

# Run tests
go test ./...

# Run manually
go run main.go
```

## License

MIT License - see the [LICENSE](LICENSE) file for details. 