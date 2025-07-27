# Sinkzone

```
  ██████  ██▓ ███▄    █  ██ ▄█▀▒███████▒ ▒█████   ███▄    █ ▓█████ 
▒██    ▒ ▓██▒ ██ ▀█   █  ██▄█▒ ▒ ▒ ▒ ▒ ▄▀░▒██▒  ██▒ ██ ▀█   █ ▓█   ▀ 
░ ▓██▄   ▒██▒▓██  ▀█ ██▒▓███▄░ ░ ▒ ▄▀▒░ ▒██░  ██▒▓██  ▀█ ██▒▒███   
  ▒   ██▒░██░▓██▒  ▐▌██▒▓██ █▄   ▄▀▒   ░▒██   ██░▓██▒  ▐▌██▒▒▓█  ▄ 
▒██████▒▒░██░▒██░   ▓██░▒██▒ █▄▒███████▒░ ████▓▒░▒██░   ▓██░░▒████▒
▒ ▒▓▒ ▒ ░░▓  ░ ▒░   ▒ ▒ ▒ ▒▒ ▓▒░▒▒ ▓░▒░▒░ ▒░▒░▒░ ░ ▒░   ▒ ▒ ░░ ▒░ ░
░ ░▒  ░ ░ ▒ ░░ ░░   ░ ▒░░ ░▒ ▒░░░▒ ▒ ░ ▒  ░ ▒ ▒░ ░ ░░   ░ ▒░ ░ ░  ░
░  ░  ░   ▒ ░   ░   ░ ░ ░ ░░ ░ ░ ░ ░ ░░ ░ ░ ▒     ░   ░ ░    ░   
      ░   ░           ░ ░  ░     ░ ░        ░ ░           ░    ░  ░
                               ░                                   
```

> **A DNS-based productivity tool that helps you stay focused by blocking distracting websites during focus sessions.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)](https://github.com/berbyte/sinkzone/releases)

## 🚀 Features

- **🔒 DNS-based Blocking**: Blocks distracting websites at the DNS level
- **🎯 Focus Mode**: Temporarily block non-allowlisted domains with automatic expiration
- **📊 Real-time Monitoring**: Beautiful TUI for monitoring DNS traffic and managing allowlist
- **⚡ Instant State Sync**: Real-time communication between resolver and TUI
- **🛡️ Service Integration**: Run as system service on Linux/macOS
- **📱 Cross-platform**: macOS, Linux, and Windows support
- **🎨 Beautiful TUI**: Lipgloss-powered terminal interface with animations
- **💾 Persistent Storage**: SQLite database for queries and allowlist
- **⚙️ Easy Configuration**: YAML-based configuration with sensible defaults


## 🏗️ Architecture

Sinkzone uses a **shared SQLite database** to enable real-time communication between the root DNS resolver and unprivileged TUI process:

- **🔧 DNS Resolver** (runs as root): Records all DNS queries and applies focus mode rules
- **🖥️ TUI Interface** (runs as user): Monitors DNS stats and manages allowlist
- **💾 Database**: Shared SQLite file with WAL mode for concurrent access
- **⚡ State Management**: File-based state sync for instant focus mode changes

## Architecture

Sinkzone uses a **shared SQLite database** to enable communication between the root DNS resolver and unprivileged monitor process:

- **DNS Resolver** (runs as root): Records all DNS queries to SQLite database
- **Monitor** (runs as user): Reads DNS stats and manages allowlist via TUI
- **Database**: Shared SQLite file with WAL mode for concurrent access

## 📦 Installation

### Homebrew (Recommended)

```bash
# Add the tap and install
brew tap berbyte/sinkzone
brew install sinkzone
```

### Manual Installation

```bash
# Clone the repository
git clone https://github.com/berbyte/sinkzone.git
cd sinkzone

# Build the binary
go build -o sinkzone .

# Install (optional)
sudo cp sinkzone /usr/local/bin/
```

### Download Binaries

Download pre-built binaries from the [releases page](https://github.com/berbyte/sinkzone/releases).

## 🚀 Quick Start

1. **Start the DNS resolver** (requires root):
   ```bash
   sudo sinkzone resolver
   ```

2. **Open the TUI in another terminal**:
   ```bash
   sinkzone
   ```

3. **Enable focus mode**:
   - Press `f` in the TUI, or
   - Run: `sinkzone focus 1h`

4. **Configure your system DNS**:
   - Set DNS server to `127.0.0.1`
   - Or use the resolver as a forwarder

## 🎮 Usage

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

## 🔧 Service Management

### Linux (systemd)

```bash
# Enable and start the service
sudo systemctl enable sinkzone-resolver
sudo systemctl start sinkzone-resolver

# Check status
sudo systemctl status sinkzone-resolver

# Stop the service
sudo systemctl stop sinkzone-resolver
```

### macOS (launchd)

```bash
# Enable and start the service
sudo launchctl load /Library/LaunchDaemons/com.berbyte.sinkzone.resolver.plist

# Stop the service
sudo launchctl unload /Library/LaunchDaemons/com.berbyte.sinkzone.resolver.plist

# Check logs
tail -f /var/log/sinkzone-resolver.log
```

## ⚙️ Configuration

Configuration files are stored in `~/.sinkzone/`:

### Main Configuration (`~/.sinkzone/sinkzone.yaml`)

```yaml
mode: normal  # or "focus"
upstream_nameservers:
  - "8.8.8.8"
  - "1.1.1.1"
```

### State Management (`~/.sinkzone/state.json`)

```json
{
  "focus_mode": false,
  "focus_end_time": "2025-07-28T00:18:19.135892+02:00",
  "last_updated": "2025-07-27T23:18:19.135892+02:00"
}
```

### Database (`~/.sinkzone/sinkzone.db`)

The SQLite database contains:
- **dns_queries**: All DNS queries with timestamps and blocked status
- **allowlist**: Active domains that are allowed during focus mode

## 🖥️ TUI Interface

### Monitoring Tab
- **Real-time DNS traffic statistics**
- **Domain query counts and last seen times**
- **Blocked vs allowed status**
- **Add/remove domains from allowlist**
- **Sort by query count**

### Allowed Domains Tab
- **Manage your allowlist**
- **Add/remove domains**
- **View current allowlist**

### Settings Tab
- **Configure upstream DNS resolvers**
- **View current configuration**
- **Save settings**

### About Tab
- **Help information**
- **Usage instructions**

## 🔍 How It Works

### Normal Mode
- All DNS requests are forwarded to upstream nameservers without blocking
- Real-time monitoring of DNS traffic
- Allowlist management

### Focus Mode
- Only domains in the allowlist are resolved
- All other domains return `NXDOMAIN` (domain not found)
- Automatically expires after the specified duration
- Red banner indicator in TUI

### DNS Resolver
- Listens on port 53 (requires root)
- Records all queries to SQLite database
- Forwards requests to configured upstream nameservers
- Applies focus mode rules based on allowlist

### TUI Interface
- Reads DNS statistics from SQLite database
- Provides tabs for monitoring and configuration
- Real-time state synchronization
- Beautiful terminal interface with animations

## 🌐 System Integration

### macOS
```bash
# Set DNS server to localhost
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

### Linux
```bash
# Edit /etc/resolv.conf
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

### Windows
- Network Settings → Change adapter options → Properties → Internet Protocol Version 4 → Properties → Use the following DNS server addresses: 127.0.0.1

## 🛠️ Development

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
│   └── tui/           # TUI interface
└── tap/               # Homebrew tap
    └── Formula/
        └── sinkzone.rb # Homebrew formula
```

## 📚 Dependencies

- `github.com/miekg/dns` - DNS server implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/huh` - Interactive forms
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/mattn/go-sqlite3` - SQLite driver
- `gopkg.in/yaml.v3` - YAML configuration parsing

## 📄 License

MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📈 Roadmap

- [ ] Windows service support
- [ ] Web interface
- [ ] Mobile app companion
- [ ] Advanced analytics
- [ ] Custom block lists
- [ ] Integration with productivity tools

## ⭐ Support

If you find this project helpful, please consider giving it a star! ⭐

For issues and feature requests, please use the [GitHub Issues](https://github.com/berbyte/sinkzone/issues) page. 