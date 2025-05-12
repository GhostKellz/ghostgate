# GhostGate

[![Build Status](https://github.com/ghostkellz/ghostgate/actions/workflows/go.yml/badge.svg)](https://github.com/ghostkellz/ghostgate/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/ghostkellz/ghostgate.svg)](https://pkg.go.dev/github.com/ghostkellz/ghostgate)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghostkellz/ghostgate)](https://goreportcard.com/report/github.com/ghostkellz/ghostgate)
[![Issues](https://img.shields.io/github/issues/ghostkellz/ghostgate)](https://github.com/ghostkellz/ghostgate/issues)
[![License: AGPL v3](https://img.shields.io/badge/license-AGPLv3-blue.svg)](LICENSE)

---

### 💡 What is GhostGate?

**GhostGate** is a modern HTTP server and reverse proxy written in Go.
Designed for flexibility, it aims to power:

* Static file hosting (ideal for Arch repo/AUR mirrors)
* Secure reverse proxy for web services
* Eventually: certificate handling, middleware, caching, and beyond

Whether you're serving `.pkg.tar.zst` files or routing traffic to internal services, **GhostGate** is a reliable, lightweight entrypoint.

---

### 🚀 Features

* Serve static files from any directory
* Configurable port, logging, and bind host
* Reverse proxy support (planned)
* Zero dependencies, just Go
* Fast startup and hot reloads (planned)

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

#### Core Functionality

* [ ] YAML/JSON config support
* [ ] Command-line flags and overrides
* [ ] Graceful restart and reload
* [ ] Built-in logging with access/output formats

#### Static Server Features

* [ ] Directory index support
* [ ] MIME type detection
* [ ] Custom 404/403 error pages

#### Reverse Proxy Features

* [ ] Path/host-based routing
* [ ] Header injection and manipulation
* [ ] Rate limiting and filtering
* [ ] TLS passthrough and termination

#### DevOps/Deploy

* [ ] systemd unit file and `.deb`/`.pkg.tar.zst` packaging
* [ ] Dockerfile and GitHub Container Registry publishing
* [ ] CI pipeline (GitHub Actions)

---

### 📝 License

**AGPL v3** — See [LICENSE](LICENSE) for details.

GhostGate is part of the GhostKellz infrastructure tooling ecosystem.
