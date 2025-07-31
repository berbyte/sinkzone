<a id="readme-top"></a>
---

<div align="center">
  <img src="https://share.ber.sh/sinkzone-splash.png" alt="Sinkzone: DNS-based Productivity Tool" width="600">
  <h1 align="center">Sinkzone: DNS-based Productivity Tool</h1>
  <p align="center">
    Stay focused by blocking distractions at the DNS level.
    <br /><br />
    <a href="#what-is-sinkzone"><strong>Learn More »</strong></a>
    &middot;
    <a href="#quick-start">Quick Start</a>
    &middot;
    <a href="https://github.com/berbyte/sinkzone/issues/new">Report a Bug</a>
    &middot;
    <a href="#usage">Usage Guide</a>
  </p>

  <p align="center">
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-blue.svg" alt="Go Version" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License" /></a>
    <a href="https://github.com/berbyte/sinkzone/releases"><img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg" alt="Platform" /></a>
  </p>
  
</div>

---

<details>
<summary><b>📚 Table of Contents</b></summary>


- [](#)
- [What is Sinkzone?](#what-is-sinkzone)
- [Motivation](#motivation)
- [Key Features](#key-features)
- [Quick Start](#quick-start)
  - [Docker](#docker)
  - [Configure System DNS (Required)](#configure-system-dns-required)
  - [Alternative Installation Methods](#alternative-installation-methods)
- [Demos](#demos)
  - [Command Line Interface (CLI)](#command-line-interface-cli)
  - [Terminal User Interface (TUI)](#terminal-user-interface-tui)
- [Documentation](#documentation)
  - [Manual Page](#manual-page)
- [Usage](#usage)
  - [Common Commands](#common-commands)
  - [TUI Navigation](#tui-navigation)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [Normal Mode](#normal-mode)
  - [Focus Mode](#focus-mode)
- [Configuration](#configuration)
- [Development](#development)
- [License](#license)
- [Contact](#contact)

</details>



---
## What is Sinkzone?

Sinkzone is a local DNS resolver that helps you eliminate distractions and get deep work done. It blocks all domains by default — only the ones you explicitly allow can get through. This means notifications, social media, news, and other time-sinks are unreachable at the network level — not just in your browser.

It's lightweight, cross-platform, and built for hackers, makers, and anyone serious about focus.

## Motivation

Most tools make you list what you want to block. But the internet is infinite — that list never ends. It's much easier to list the few things you actually want to allow.

Sinkzone was born from that insight. I was tired of coding sessions interrupted by Slack pings and email alerts. I needed something stronger than a browser plugin — a system-level kill switch for distractions.

Now I can code for hours uninterrupted. Even my son uses Sinkzone during chess practice to stay focused.

**Sinkzone exists because I needed it. Maybe you do too.**

![Sinkzone TUI](examples/tui-screenshot.png)

*The Sinkzone Terminal User Interface showing real-time DNS monitoring and allowlist management*

---

## Key Features

- **DNS-level blocking**: Stops distractions before they reach your apps
- **Focus Mode**: Block all but allowlisted domains for a set duration
- **Terminal UI**: Real-time DNS traffic viewer with tabbed interface
- **Memory-backed rules**: Focus mode expires automatically
- **Cross-platform**: Works on macOS and Linux

---

## Quick Start

### Docker

The easiest way to run Sinkzone on any platform:

```bash
# Pull and run the latest image
docker run -d \
  --name sinkzone \
  --network host \
  --cap-add NET_BIND_SERVICE \
  --restart unless-stopped \
  -v ~/.sinkzone:/app/.sinkzone \
  --platform linux/amd64 \
  ghcr.io/berbyte/sinkzone:latest resolver
```

**That's it!** Sinkzone is now running and blocking distractions at the DNS level.

**Note:** If you're on Apple Silicon (M1/M2), you may need to specify the platform explicitly:
```bash
docker run -d \
  --name sinkzone \
  --network host \
  --cap-add NET_BIND_SERVICE \
  --restart unless-stopped \
  -v ~/.sinkzone:/app/.sinkzone \
  --platform linux/amd64 \
  ghcr.io/berbyte/sinkzone:latest resolver
```

**Next steps:**
```bash
# Check status
docker exec sinkzone status

# View DNS requests
docker exec sinkzone monitor

# Add github.com
docker exec sinkzone allowlist add github.com

# Enable focus mode
docker exec sinkzone focus start

# View logs
docker logs -f sinkzone
```

### Configure System DNS (Required)

**Important:** You must configure your system to use Sinkzone as the DNS resolver for it to work.

**macOS:**
```bash
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

**Linux:**
```bash
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

**Windows:**
- Open Network & Internet settings
- Change adapter options
- Right-click your network adapter → Properties
- Select "Internet Protocol Version 4 (TCP/IPv4)" → Properties
- Select "Use the following DNS server addresses"
- Enter `127.0.0.1` as the preferred DNS server

### Alternative Installation Methods

<details>
<summary><b>📦 Package Managers</b></summary>

**Homebrew (macOS):**
```bash
brew tap berbyte/ber
brew install berbyte/ber/sinkzone
```

**Manual Setup:**
```bash
# 1. Start the DNS Resolver (requires root)
sudo sinkzone resolver

# 2. Launch the UI (in another terminal)
sinkzone tui

# 3. Enable Focus Mode
sinkzone focus start
```

</details>

<details>
<summary><b>🔨 Build from Source</b></summary>

```bash
# Clone and build
git clone https://github.com/berbyte/sinkzone.git
cd sinkzone
go build -o sinkzone .

# Follow the manual setup steps above
```

</details>

<details>
<summary><b>📥 Direct Download</b></summary>

Download the appropriate binary for your platform:

**macOS:**
```bash
# Apple Silicon (M1/M2)
curl -L -o sinkzone https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-darwin-arm64
chmod +x sinkzone
sudo mv sinkzone /usr/local/bin/

# Intel Mac
curl -L -o sinkzone https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-darwin-amd64
chmod +x sinkzone
sudo mv sinkzone /usr/local/bin/
```

**Linux:**
```bash
# AMD64
curl -L -o sinkzone https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-amd64
chmod +x sinkzone
sudo mv sinkzone /usr/local/bin/

# ARM64
curl -L -o sinkzone https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-arm64
chmod +x sinkzone
sudo mv sinkzone /usr/local/bin/
```

</details>

---

## Demos

### Command Line Interface (CLI)

The CLI offers powerful command-line tools for system management:

![CLI Demo](examples/demo-cli.gif)

*Command-line allowlist management, focus mode control, and system status monitoring*

### Terminal User Interface (TUI)

The TUI provides real-time DNS monitoring and allowlist management:

![TUI Demo](examples/demo-tui.gif)

*Real-time DNS traffic monitoring, allowlist management, and focus mode control*

---

## Documentation

### Manual Page

For detailed documentation, run:
```bash
sinkzone man
```

---

## Usage

### Common Commands

| Command                  | Description                    |
| ------------------------ | ------------------------------ |
| `sinkzone monitor`       | Show last 20 DNS requests      |
| `sinkzone tui`           | Launch the terminal UI         |
| `sudo sinkzone resolver` | Start DNS resolver on port 53  |
| `sinkzone focus start`   | Enable focus mode for 1 hour   |
| `sinkzone focus --disable` | Disable focus mode immediately |
| `sinkzone status`        | View current focus mode state  |
| `sinkzone allowlist add <domain>` | Add domain to allowlist |
| `sinkzone allowlist remove <domain>` | Remove domain from allowlist |
| `sinkzone allowlist list` | List all allowed domains |
| `sinkzone config set resolver <ip>` | Set resolver IP |
| `sinkzone man` | Show manual page |

### TUI Navigation

* `←`/`→`: Switch tabs
* `f`: Enable focus mode (1 hour)
* `q`: Quit
* Tabs include:

  * **Monitor**: Real-time DNS traffic
  * **Allowlist**: Add or remove allowed domains
  * **Settings**: DNS resolver config


## How It Works

### Architecture

Sinkzone is composed of two parts:

* **Resolver**: A local DNS server that intercepts queries and maintains real-time data via Unix socket.
* **TUI**: A terminal UI for interacting with and monitoring the system via socket communication.

### Normal Mode

* All DNS queries are forwarded to upstream resolvers
* You can view and manage DNS traffic and allowlist

### Focus Mode

* Only allowlisted domains resolve
* Everything else returns `NXDOMAIN`
* Automatically expires after specified duration

---

## Configuration

Files are stored in `~/.sinkzone/`:

* `sinkzone.yaml`: Main config
* `sinkzone.sock`: Unix socket for real-time communication between resolver and TUI
* `allowlist.txt`: Simple text file containing allowed domains

---

## Development

```bash
# Build binary
go build -o sinkzone .

# Run tests
go test ./...

# Run directly
go run main.go
```

PRs and issues welcome. We love contributors.

---

## License

MIT License. See the [LICENSE](LICENSE) file for full details.

---

## Contact

* Email: [dominis@ber.run](mailto:dominis@ber.run)
* GitHub: [github.com/berbyte/sinkzone](https://github.com/berbyte/sinkzone)

<p align="right">(<a href="#readme-top">back to top</a>)</p>