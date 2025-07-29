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
  - [1. Install](#1-install)
  - [2. Start the DNS Resolver](#2-start-the-dns-resolver)
  - [3. Set System DNS to Localhost](#3-set-system-dns-to-localhost)
  - [4. Launch the UI](#4-launch-the-ui)
  - [5. Enable Focus Mode](#5-enable-focus-mode)
- [Usage](#usage)
  - [Common Commands](#common-commands)
  - [TUI Navigation](#tui-navigation)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [Normal Mode](#normal-mode)
  - [Focus Mode](#focus-mode)
- [Service Management](#service-management)
  - [Linux (systemd)](#linux-systemd)
  - [macOS (launchd)](#macos-launchd)
- [Configuration](#configuration)
- [Development](#development)
- [License](#license)
- [Contact](#contact)

</details>



---
## What is Sinkzone?

Sinkzone is a local DNS resolver and terminal UI (TUI) that helps you eliminate distractions and get deep work done. It blocks all domains by default — only the ones you explicitly allow can get through. This means notifications, social media, news, and other time-sinks are unreachable at the network level — not just in your browser.

It’s lightweight, cross-platform, and built for hackers, makers, and anyone serious about focus.

## Motivation

Most tools make you list what you want to block. But the internet is infinite — that list never ends. It’s much easier to list the few things you actually want to allow.

Sinkzone was born from that insight. I was tired of coding sessions interrupted by Slack pings and email alerts. I needed something stronger than a browser plugin — a system-level kill switch for distractions.

Now I can code for hours uninterrupted. Even my son uses Sinkzone during chess practice to stay focused.

**Sinkzone exists because I needed it. Maybe you do too.**

---

## Key Features

- **DNS-level blocking**: Stops distractions before they reach your apps
- **Focus Mode**: Block all but allowlisted domains for a set duration
- **Terminal UI**: Real-time DNS traffic viewer with tabbed interface
- **Memory-backed rules**: Focus mode expires automatically
- **Cross-platform**: Works on macOS, Linux, and Windows

---

## Quick Start

### 1. Install

**Homebrew (macOS):**
```bash
brew tap berbyte/ber
brew install berbyte/ber/sinkzone
xattr -d com.apple.quarantine $(which sinkzone)
````

**Manual (all platforms):**

```bash
# Download from https://github.com/berbyte/sinkzone/releases or build from source
go build -o sinkzone .
```

### 2. Start the DNS Resolver

Port 53 requires root privileges:

```bash
sudo sinkzone resolver
```

### 3. Set System DNS to Localhost

**macOS:**

```bash
sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

**Linux:**

```bash
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

**Windows:**

* Go to Network Settings → Adapter Options → IPv4 Settings → Use DNS `127.0.0.1`

> 💡 On some platforms, you may need to persist DNS settings with a launch script or system manager.

### 4. Launch the UI

In a second terminal:

```bash
sinkzone
```

### 5. Enable Focus Mode

```bash
sinkzone focus 1h
```

Or press `f` in the TUI.

---

## Usage

### Common Commands

| Command                  | Description                    |
| ------------------------ | ------------------------------ |
| `sinkzone`               | Launch the terminal UI         |
| `sudo sinkzone resolver` | Start DNS resolver on port 53  |
| `sinkzone focus 1h`      | Enable focus mode for 1 hour   |
| `sinkzone disable-focus` | Disable focus mode immediately |
| `sinkzone status`        | View current focus mode state  |

### TUI Navigation

* `←`/`→`: Switch tabs
* `f`: Enable focus mode (1 hour)
* `q`: Quit
* Tabs include:

  * **Monitor**: Real-time DNS traffic
  * **Allowlist**: Add or remove allowed domains
  * **Settings**: DNS resolver config

> ![TUI Screenshot Placeholder](https://share.ber.sh/sinkzone-demo.gif)
> *Demo: DNS monitoring and focus toggle*

---

## How It Works

### Architecture

Sinkzone is composed of two parts:

* **Resolver**: A local DNS server that intercepts queries.
* **TUI**: A terminal UI for interacting with and monitoring the system.

### Normal Mode

* All DNS queries are forwarded to upstream resolvers
* You can view and manage DNS traffic and allowlist

### Focus Mode

* Only allowlisted domains resolve
* Everything else returns `NXDOMAIN`
* Automatically expires after specified duration
* Red visual banner in the TUI

---

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

> 📝 Example service files are available in the [`contrib/`](contrib/) folder.

---

## Configuration

Files are stored in `~/.sinkzone/`:

* `sinkzone.yaml`: Main config
* `sinkzone.db`: SQLite database of DNS queries and allowlist
* `state.json`: Focus mode tracking

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
