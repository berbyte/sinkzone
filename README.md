<a id="readme-top"></a>

<div align="center">
  <img src="https://share.ber.sh/sinkzone-splash.png" alt="Sinkzone: DNS-based Productivity Tool">
  <h1 align="center">Sinkzone: DNS-based Productivity Tool</h1>
  <p align="center">
    <a href="#what-is-sinkzone"><strong>Learn More »</strong></a>
    <br />
    <br />
    <a href="#quick-start">Quick Start</a>
    &middot;
    <a href="https://github.com/berbyte/sinkzone/issues/new">Report a Bug</a>
    &middot;
    <a href="#usage">Usage Guide</a>
  </p>
</div>

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)](https://github.com/berbyte/sinkzone/releases)

</div>

## What is Sinkzone?

Sinkzone is a DNS-based productivity tool that helps you stay focused by blocking distracting websites at the DNS level. When you enable focus mode, only domains in your allowlist will resolve - everything else returns NXDOMAIN.

## Why I Built This:

I constantly found myself getting interrupted by notifications from Slack and email during coding sessions - losing hours of productive time. I built Sinkzone to force myself into deep focus mode. Now I can code for hours without distractions.

My son uses it on his Windows box during his chess training sessions to stay focused without getting sidetracked.

**Sinkzone is built to solve our problems.**

<hr />

**Key Features:**
- 🔒 **DNS-level blocking** - Blocks at the network level, not just browser
- 🎯 **Focus mode** - Temporarily block non-allowlisted domains with auto-expiration
- 📊 **Real-time monitoring** - Beautiful TUI to monitor DNS traffic and manage allowlist
- ⚡ **Instant state sync** - Real-time communication between resolver and TUI
- 🛡️ **System service** - Run as background service on Linux/macOS

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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

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

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## License

MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

- Sinkzone - github@ber.run
- Project Link: [https://github.com/berbyte/sinkzone](https://github.com/berbyte/sinkzone)

<p align="right">(<a href="#readme-top">back to top</a>)</p> 