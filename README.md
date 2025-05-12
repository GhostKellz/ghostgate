# GhostGate

---

### 💡 What is GhostGate?

**GhostGate** is a modern HTTP server and reverse proxy written in Go.

Designed for flexibility and performance, it powers:

* Static file hosting (perfect for AUR mirrors or internal repos)
* Secure reverse proxying with TLS and fine-grained routing
* Auto TLS via Let's Encrypt (no certbot needed)
* Fast startup, hot reloading, and DevOps-ready deployment

---

### 🚀 Features

* Serve static files with MIME type detection
* Directory listings and custom 404/403 error pages
* Reverse proxy with path-based routing
* Header injection and request manipulation
* Rate limiting and filtering (WIP)
* Built-in TLS via Let's Encrypt using `autocert`
* YAML-based configuration (supports `gate.conf` + `conf.d/*.yaml`)
* Systemd unit file for service deployment
* Gzip compression and static file caching
* Logging with customizable levels and formats (JSON or plain)
* Welcome page when no config is present

---

### ⚙️ Getting Started

```bash
git clone https://github.com/ghostkellz/ghostgate.git
cd ghostgate
go build -o ghostgate
./ghostgate
```

---

### 📅 Roadmap

#### ✅ Core Functionality

*

#### ✅ Static Server Features

*

#### ✅ Reverse Proxy Features

*

#### ✅ DevOps/Deploy

*

---

### 📝 License

**AGPL v3** — See [LICENSE](LICENSE) for details.

GhostGate is part of the \*\*CK Technology \*\*  infrastructure tooling ecosystem.
