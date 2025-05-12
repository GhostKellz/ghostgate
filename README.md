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

* TBD

#### Static Server Features

* TBD

#### Reverse Proxy Features

* TBD

#### Docker Compose

* TBD

---

### 📝 License

**GNU Affero General Public License v3.0 (AGPL-3.0)** — See [LICENSE](LICENSE) for full terms.

GhostGate is free and open software: you can use, study, share, and modify it.  
However, if you deploy a modified version of GhostGate as part of a public service, you must also publish the source code of your modifications.
