<a id="readme-top"></a>

<div align="center">
  <img src="https://share.ber.sh/sinkzone-splash.png" alt="Sinkzone: DNS-based Productivity Tool" width="600">
  <h1 align="center">Sinkzone: Block-by-default DNS tool for deep focus</h1>
  <p align="center">
      The internet is infinite. Your focus isn’t. <br />
      <br />
      Sinkzone helps you reclaim control by flipping the default: <strong>everything is blocked</strong>, unless you explicitly allow it. <br/><br/>
      No feeds. No pings. No surprise connections.<br />
      <h2>Just a quiet, intentional internet.</h2>
    <br /><br />
    <img src="examples/demo-cli.gif" alt="Sinkzone Demo" />
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

- [What is Sinkzone?](#what-is-sinkzone)
- [Motivation](#motivation)
- [Why use Sinkzone?](#why-use-sinkzone)
- [Key Features](#key-features)
- [Demos](#demos)
  - [Terminal User Interface (TUI)](#terminal-user-interface-tui)
  - [Command Line Interface (CLI)](#command-line-interface-cli)
- [Quick Start](#quick-start)
  - [Installation by Platform](#installation-by-platform)
- [Documentation](#documentation)
  - [Manual Page](#manual-page)
- [Usage](#usage)
  - [Common Commands](#common-commands)
  - [Wildcard Patterns](#wildcard-patterns)
  - [TUI Navigation](#tui-navigation)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [API Endpoints](#api-endpoints)
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

It features a modern HTTP API, wildcard pattern support, and a beautiful terminal UI for real-time monitoring and control.

It's lightweight, cross-platform, and built for hackers, makers, and anyone serious about focus.

## Motivation

Most tools make you list what you want to block. But the internet is infinite — that list never ends. It's much easier to list the few things you actually want to allow.

Sinkzone was born from that insight. I was tired of coding sessions interrupted by Slack pings and email alerts. I needed something stronger than a browser plugin — a system-level kill switch for distractions.

Now I can code for hours uninterrupted. Even my son uses Sinkzone during chess practice to stay focused.

**Sinkzone exists because I needed it. Maybe you do too.**


---

## Why use Sinkzone?

- Avoid distraction without browser extensions or hacks
- Run locally on macOS, Linux and Windows — no cloud, no telemetry
- Perfect for deep work sessions, writing, coding, or child safety
- Terminal UI included — toggle modes, view logs, edit allowlist

---

## Key Features

- **DNS-level blocking**: Stops distractions before they reach your apps
- **Focus Mode**: Block all but allowlisted domains for a set duration
- **Wildcard Support**: Use patterns like `*github*` or `*.google.com` for flexible domain matching
- **HTTP API**: RESTful API for monitoring and control
- **Terminal UI**: Real-time DNS traffic viewer with tabbed interface
- **Memory-backed rules**: Focus mode expires automatically
- **Cross-platform**: Works on macOS and Linux

---


## Demos

### Terminal User Interface (TUI)

The TUI provides real-time DNS monitoring and allowlist management:

![TUI Demo](examples/demo-tui.gif)

*Real-time DNS traffic monitoring, allowlist management, and focus mode control*

### Command Line Interface (CLI)

The CLI offers powerful command-line tools for system management:

![CLI Demo](examples/demo-cli.gif)

*Command-line allowlist management, focus mode control, and system status monitoring*


---

## Quick Start

### Installation by Platform

<details>
<summary><b>👉 macOS Installation</b></summary>

**Homebrew (Recommended):**
```bash
brew tap berbyte/ber
brew install sinkzone
```

**Manual Setup:**
```bash
# 1. Start the DNS Resolver (default port 53, requires admin privileges)
sudo sinkzone resolver

# 2. Launch the UI (in another terminal)
sinkzone tui

# 3. Enable Focus Mode
sinkzone focus start
```

**Configure System DNS (Required):**
```bash
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

**Direct Download:**
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

</details>

---

<details>
<summary><b>👉 Linux Installation</b></summary>

**Package Managers (Recommended):**

**Debian/Ubuntu:**
```bash
# Download and install the .deb package
curl -L -O https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-amd64.deb
sudo dpkg -i sinkzone-linux-amd64.deb
```

**Red Hat/Fedora/CentOS:**
```bash
# Download and install the .rpm package
curl -L -O https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-amd64.rpm
sudo rpm -i sinkzone-linux-amd64.rpm
```

**Alpine Linux:**
```bash
# Download and install the .apk package
curl -L -O https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-amd64.apk
sudo apk add --allow-untrusted sinkzone-linux-amd64.apk
```

**Arch Linux:**
```bash
# Download and install the .pkg.tar.zst package
curl -L -O https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-linux-amd64.pkg.tar.zst
sudo pacman -U sinkzone-linux-amd64.pkg.tar.zst
```

**Manual Installation:**
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

**Manual Setup:**
```bash
# 1. Start the DNS Resolver (default port 53, requires admin privileges)
sudo sinkzone resolver

# 2. Launch the UI (in another terminal)
sinkzone tui

# 3. Enable Focus Mode
sinkzone focus start
```

**Configure System DNS (Required):**
```bash
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

**Note:** Package installations include the manual page. Run `man sinkzone` for detailed documentation.

</details>

---

<details>
<summary><b>👉 Windows Installation</b></summary>

**Direct Download:**
```powershell
# AMD64
Invoke-WebRequest -Uri "https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-windows-amd64.exe" -OutFile "sinkzone.exe"
# Move to a directory in your PATH (e.g., C:\Windows\System32 or create a custom directory)
Move-Item sinkzone.exe C:\Windows\System32\sinkzone.exe

# ARM64
Invoke-WebRequest -Uri "https://github.com/berbyte/sinkzone/releases/latest/download/sinkzone-windows-arm64.exe" -OutFile "sinkzone.exe"
# Move to a directory in your PATH (e.g., C:\Windows\System32 or create a custom directory)
Move-Item sinkzone.exe C:\Windows\System32\sinkzone.exe
```

**Manual Setup:**
```powershell
# 1. Start the DNS Resolver (run as Administrator for port 53)
sinkzone resolver

# 2. Launch the UI (in another terminal)
sinkzone tui

# 3. Enable Focus Mode
sinkzone focus start
```

**Configure System DNS (Required):**
- Open Network & Internet settings
- Change adapter options
- Right-click your network adapter → Properties
- Select "Internet Protocol Version 4 (TCP/IPv4)" → Properties
- Select "Use the following DNS server addresses"
- Enter `127.0.0.1` as the preferred DNS server

**Note:** On Windows, you may need to run the resolver as Administrator for port 53, or use an unprivileged port like 5353.

</details>

---

<details>
<summary><b>👉 Build from Source</b></summary>

```bash
# Clone and build
git clone https://github.com/berbyte/sinkzone.git
cd sinkzone
go build -o sinkzone .

# Follow the manual setup steps above for your platform
```

</details>

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
| `sinkzone resolver`      | Start DNS resolver on port 53  |
| `sinkzone focus start`   | Enable focus mode for 1 hour   |
| `sinkzone focus --disable` | Disable focus mode immediately |
| `sinkzone status`        | View current focus mode state  |
| `sinkzone allowlist add <domain>` | Add domain to allowlist |
| `sinkzone allowlist add "*github*"` | Add wildcard pattern |
| `sinkzone allowlist remove <domain>` | Remove domain from allowlist |
| `sinkzone allowlist list` | List all allowed domains |
| `sinkzone config set resolver <ip>` | Set resolver IP |
| `sinkzone man` | Show manual page |

**Note:** On Unix-like systems (macOS/Linux), you may need to run `sudo sinkzone resolver` for port 53. On Windows, run as Administrator or use an unprivileged port like 5353.

### Wildcard Patterns

Sinkzone supports wildcard patterns for flexible domain matching:

| Pattern | Matches | Examples |
|---------|---------|----------|
| `*github*` | Any domain containing "github" | `github.com`, `api.github.com`, `githubusercontent.com` |
| `*.google.com` | All subdomains of google.com | `maps.google.com`, `drive.google.com`, `docs.google.com` |
| `api.*.com` | Any api subdomain of .com domains | `api.github.com`, `api.example.com`, `api.stackoverflow.com` |
| `exact.com` | Exact domain match only | `exact.com` (not `sub.exact.com`) |

**Examples:**
```bash
# Allow all GitHub-related domains
sinkzone allowlist add "*github*"

# Allow all Google subdomains
sinkzone allowlist add "*.google.com"

# Allow all API subdomains
sinkzone allowlist add "api.*.com"

# Allow exact domain
sinkzone allowlist add "stackoverflow.com"
```

### TUI Navigation

* `←`/`→`: Switch tabs
* `f`: Enable focus mode (1 hour)
* `ESC`: Quit
* Tabs include:

  * **Monitor**: Real-time DNS traffic
  * **Allowlist**: Add or remove allowed domains
  * **Settings**: DNS resolver config


## How It Works

### Architecture

Sinkzone is composed of three parts:

* **Resolver**: A local DNS server that intercepts queries and maintains real-time data via HTTP API.
* **HTTP API Server**: Provides REST endpoints for monitoring DNS queries and controlling focus mode.
* **TUI/CLI**: User interfaces that communicate with the resolver via HTTP API.
* **TUI**: A terminal UI for interacting with and monitoring the system via HTTP API.

### API Endpoints

The resolver exposes the following HTTP endpoints:

- `GET /api/queries` - Get the last 100 DNS queries
- `GET /api/focus` - Get current focus mode state
- `POST /api/focus` - Set focus mode (enabled/disabled, duration)
- `GET /api/state` - Get complete resolver state
- `GET /health` - Health check endpoint

**API Usage Examples:**
```bash
# Start resolver with custom API port
sinkzone resolver --port 53 --api-port 8080

# Use CLI with custom API URL
sinkzone monitor --api-url http://127.0.0.1:8080
sinkzone focus --enable --api-url http://127.0.0.1:8080
sinkzone tui --api-url http://127.0.0.1:8080

# Direct API calls
curl http://127.0.0.1:8080/api/queries
curl http://127.0.0.1:8080/api/focus
curl -X POST http://127.0.0.1:8080/api/focus \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "duration": "1h"}'
```

### Normal Mode

* All DNS queries are forwarded to upstream resolvers
* You can view and manage DNS traffic and allowlist

### Focus Mode

* Only allowlisted domains resolve
* Everything else returns `NXDOMAIN`
* Automatically expires after specified duration
* Allowlist is reloaded when focus mode is enabled (changes take effect on new focus sessions)

---

## Configuration

Files are stored in `~/.sinkzone/`:

* `sinkzone.yaml`: Main config
* `allowlist.txt`: Simple text file containing allowed domains (supports wildcard patterns)
* `resolver.pid`: Process ID file for the DNS resolver

**Allowlist Format:**
```
# Comments start with #
github.com
stackoverflow.com
*github*
*.google.com
api.*.com
```

---

## Development

```bash
# Build binary
go build -o sinkzone .

# Run tests
go test ./...

# Run resolver with custom ports
sinkzone resolver --port 5353 --api-port 8080

# Test API endpoints
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/queries

# Run TUI with custom API URL
sinkzone tui --api-url http://127.0.0.1:8080
```

**Architecture:**
- DNS Server: Handles DNS resolution and blocking
- HTTP API Server: Provides REST endpoints for monitoring and control
- CLI/TUI: User interfaces that communicate via HTTP API

PRs and issues welcome. We love contributors.

---

## License

MIT License. See the [LICENSE](LICENSE) file for full details.

---

## Contact

* Email: [dominis@ber.run](mailto:dominis@ber.run)
* GitHub: [github.com/berbyte/sinkzone](https://github.com/berbyte/sinkzone)

<p align="right">(<a href="#readme-top">back to top</a>)</p>
