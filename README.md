# GhostGate
[![Go Reference](https://pkg.go.dev/badge/github.com/ghostkellz/ghostgate.svg)](https://pkg.go.dev/github.com/ghostkellz/ghostgate)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghostkellz/ghostgate)](https://goreportcard.com/report/github.com/ghostkellz/ghostgate)
[![Issues](https://img.shields.io/github/issues/ghostkellz/ghostgate)](https://github.com/ghostkellz/ghostgate/issues)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPLv3-blue.svg)](LICENSE)

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
### 🧩 What's Inside

GhostGate now includes everything you need for modern HTTP service and reverse proxying:

- 🔐 **Automatic TLS** via Let's Encrypt (autocert)
- 🔁 **Graceful reloads** with SIGHUP signal support
- 📂 **Static file server** with directory indexing, MIME type detection, and custom error pages
- 🔀 **Reverse proxy** with path routing, header injection, and basic rate limiting
- ⚙️ **Config merging** from gate.conf + conf.d/
- 🧾 **Logging** in JSON or plain formats
- 🚀 **Systemd unit file** for production deployment
- 🌐 **Welcome page fallback** if no config is loaded
---
### 🔧 Under the Hood

GhostGate isn't just fast — it's production-ready:

- Built-in TLS certificate handling via `autocert` (no external scripts)
- Hot reloads with `SIGHUP` (no downtime on config change)
- Modular configuration: `gate.conf` + `conf.d/*.yaml`
- Customizable static file server with MIME-aware handling
- Reverse proxy engine with header rewrites and rate limiting
- Clean structured logging (JSON/plain) and gzip support
- systemd integration with ready-to-deploy unit file
---
### 🌱 Next Steps (Community Wishlist)

The core is stable — here’s what you might contribute or extend:

- [ ] Dockerfile and containerized builds
- [ ] CI pipeline with GitHub Actions
- [ ] `.deb` / `.pkg.tar.zst` packaging for Linux distros
- [ ] TLS passthrough (TCP proxying)
- [ ] Dynamic config reloads from HTTP API
---

### 📝 License

**AGPL v3** — See [LICENSE](LICENSE) for details.

GhostGate is part of the \*\*CK Technology \*\*  infrastructure tooling ecosystem.
