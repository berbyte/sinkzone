# Sinkzone 🛑

A strict DNS filter to help you stay focused — or keep your kids safe.

Sinkzone is a **local DNS forwarder** that lets you block distractions at the network level. It's built for developers and parents who want more control than browser extensions or `/etc/hosts` can provide.

## ✨ Why Sinkzone?

I built Sinkzone for myself. I wanted to:

- Cut off **distractions** when I write code.
- Give my son safe, whitelisted access to chess.com and Zoom for coaching.
- Actually **enforce lockdown**, even from myself.

Most tools don’t go far enough. Sinkzone does.

---

## 🧠 How it works

- Run a **local DNS server** with a personal allow/block list.
- Use `focus` mode to block known distractions.
- Use `lockdown` mode to enforce rules with a PIN only someone else knows.
- Use `monitor` mode to just log DNS queries without blocking.
- Simple SQLite-based config — no cloud, no accounts.

---

## 🔧 Usage

```bash
sinkzone dns start            # Start DNS filter (requires root)
sinkzone web                  # Open dashboard UI
sinkzone mode focus           # Start blocking distractions
sinkzone mode lockdown        # PIN-protected lockdown
sinkzone mode monitor         # Just watch, don't block
sinkzone mode off             # Disable all filtering
````

You can also:

```bash
sinkzone allow example.com    # Add to allowlist
sinkzone block twitter.com    # Add to blocklist
sinkzone list                 # View all rules
```

---

## 🛠 Installation

### macOS (with Homebrew)

```bash
brew install sinkzone
```

### Linux

```bash
curl -sSL https://sinkzone.ber.run/install.sh | sudo bash
```

Or build from source:

```bash
go install github.com/berbyte/sinkzone@latest
```

### Docker

```bash
docker run --net=host berbyte/sinkzone dns
```

---

## 🌐 UI Preview

> Lightweight web UI lets you switch modes and manage rules.


---

## 🧪 Status

This is an MVP — alpha-quality. Expect bugs. I built it for myself, and it works for my use cases.

✅ Works on macOS, Windows and Linux
✅ Runs in Docker or Raspberry Pi
✅ Open-source and auditable
📦 Single static binary (no runtime deps)

---

## 📁 Configuration

Sinkzone stores all data locally in `~/.sinkzone/`:

- `~/.sinkzone/config.toml` - Configuration file
- `~/.sinkzone/sinkzone.db` - SQLite database

Default configuration:
```toml
upstream_dns = "8.8.8.8:53"
pin = "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"  # SHA1 hash of "1234"
```

## 🛡️ Security

Sinkzone runs a DNS server on your machine. It needs elevated privileges to bind port 53. Run it on a separate device (like a Raspberry Pi) if you're worried — or run in Docker.

You can also set up Sinkzone on a VPS and route all traffic through it.

---

## 🤖 Roadmap

* Time-based rules (e.g., focus from 9am–12pm)
* Per-device control
* Automatic domain categorization
* Shared blocklists
* Enforced lockdown from remote parent UI

---

## 📜 License

MIT. Built for hackers, parents, and ADHD coders.

---


## 👋 Contribute

Pull requests welcome. File issues, suggest ideas, or just star the repo if you like it.
