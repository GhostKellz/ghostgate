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

## 🛠️ Installation

### Debian/Ubuntu (or Arch) — Quick Install

```bash
# From the project directory:
chmod +x install.sh
./install.sh
```
This will:
- Build the GhostGate binary
- Install it to /usr/local/bin (Debian/Ubuntu) or /usr/bin (Arch)
- Copy config files to /etc/ghostgate/
- Install and enable the systemd service

### Arch Linux — makepkg (AUR style)

```bash
# From the project directory:
makepkg -si
```
This will use the provided PKGBUILD to build and install GhostGate as a system package.

### Systemd Service

After install, GhostGate runs as a systemd service:
```bash
sudo systemctl status ghostgate.service
sudo systemctl restart ghostgate.service
```

---

## 📦 Debian/Ubuntu: .deb Package

To build a .deb package:

```bash
sudo apt install build-essential debhelper golang
cd /path/to/ghostgate
# Build the binary first (required for packaging)
go build -o ghostgate
# Build the .deb package
sudo dpkg-buildpackage -us -uc
```

This will create a .deb file in the parent directory. Install it with:

```bash
sudo dpkg -i ../ghostgate_*.deb
```

This will install the binary, config files, and systemd service, and start GhostGate automatically.

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
## 🧹 Commands

GhostGate includes the following CLI commands:

### `ghostgate serve`

Starts the GhostGate HTTP server with the specified configuration.

```bash
ghostgate serve -config /etc/ghostgate/config.yaml
```

---

### `ghostgate check`

Validates your configuration file without starting the server. Useful for CI/CD pipelines and manual debugging.

```bash
ghostgate check -config /etc/ghostgate/config.yaml
```

If the configuration is valid, you'll see:

```
✔ Configuration is valid.
```

If errors are found, they will be printed with details.

---

### `ghostgate reload`

Sends a `SIGHUP` signal to gracefully reload GhostGate configurations without restarting the server.

```bash
ghostgate reload
```

---

### `ghostgate --version`

Displays the current version of GhostGate.

```bash
ghostgate --version
```

---

### `ghostgate status`

Displays the current systemd service status for GhostGate.

```bash
ghostgate status
```

---

### `ghostgate cert -domain <example.com>`

Issues TLS certificates via ACME for a specified domain and its wildcard.

```bash
ghostgate cert -domain example.com
```

This will:

1. Use `acme.sh` to request a certificate for the domain and its wildcard (e.g. `*.example.com`).
2. Store the resulting certificates in `/etc/ghostgate/certs/<domain>`.
3. Reload GhostGate to begin using the new certificate.

This command supports automation and can be integrated into a cron job or deployment pipeline.

---

## 🔐 TLS & Encryption

GhostGate supports full TLS configuration with:

* ACME certificate issuance using `ghostgate cert`
* Certificate reload via `ghostgate reload`
* Custom cert/key paths via configuration
* Planned support for OCSP stapling and auto-renew cron

Future improvements:

* HTTP/2 + modern TLS cipher suites
* Let's Encrypt wildcard + SAN support
* Reload-free cert hot-swap

---

## 🏷️ Multi-domain/SAN Certificates & Virtual Hosts

### Multi-domain Certificate Issuance

You can issue a certificate for multiple domains at once:

```bash
ghostgate cert -domain example.com,www.example.com,api.example.com
```

This will request a single certificate valid for all listed domains.

### Per-domain (Virtual Host) Configuration

In your `gate.conf`:

```yaml
domains:
  - domain: "example.com"
    static_dir: "/srv/example.com/static"
    proxy_routes:
      - path: "/api"
        backend: "http://localhost:3000"
  - domain: "anotherdomain.com"
    static_dir: "/srv/anotherdomain.com/static"
```

GhostGate will serve the correct static files and proxy rules based on the requested Host header.

---

## ✨ Performance Enhancements (Planned)

* **Reverse Proxy Caching:** Optional response caching via Ristretto or groupcache
* **HTTP/2 + TLS Optimization:** Enable HTTP/2 and disable insecure ciphers
* **Connection Pooling:** Reduce overhead for repeated requests
* **Static Compression:** Add Brotli/gzip compression to static file responses
* **Rate Limiting:** Support per-IP limiting in config
* **Health Endpoint:** Add `/health` route for probes and monitoring
* **Log Rotation:** Support for rotating large logs automatically

Let us know which features you'd like prioritized!

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
