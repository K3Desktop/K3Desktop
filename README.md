# K3Desktop

**Desktop GUI for managing k3d Kubernetes clusters**

K3Desktop is a cross-platform desktop application built with [Wails v3](https://v3.wails.io/) (Go + Svelte). It embeds the k3d, Helm, and Docker Go packages directly — Docker runtime is the only external dependency.

![License](https://img.shields.io/github/license/k3desktop/k3desktop)
![Release](https://img.shields.io/github/v/release/k3desktop/k3desktop)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

---

## Features

| Feature | Description |
|---|---|
| **Clusters** | Create (simple or advanced YAML config), start, stop, delete k3d clusters |
| **Nodes** | Add agent nodes, start/stop/delete individual nodes |
| **Registries** | Create and manage local k3d container registries |
| **Kubeconfig** | Export to `~/.kube/config`, manage contexts |
| **Blueprints** | Reusable Helm deployment templates — deploy multi-chart stacks with one click |
| **Profiles** | Save and reuse cluster YAML configurations |

## Prerequisites

Only **Docker** is required (Desktop on macOS/Windows, Engine on Linux). No k3d CLI, Helm CLI, or kubectl needed.

## Installation

Download the latest release for your platform from the [Releases page](https://github.com/k3desktop/k3desktop/releases).

| Platform | Package |
|----------|---------|
| macOS (ARM) | `.zip` — drag `K3Desktop.app` to `/Applications` |
| Linux | `.AppImage` / `.deb` / `.rpm` / `.tar.zst` |
| Windows | `.exe` — run installer |

> **macOS first launch:** right-click → Open to bypass Gatekeeper (app is unsigned).

## Documentation

Full documentation at **[k3desktop.github.io/k3desktop](https://k3desktop.github.io/k3desktop/en/)**

Available in: English · Italiano · Español · Français · Deutsch

## Building from Source

**Requirements:** Go ≥1.25.10, Node.js ≥22, [Wails v3 CLI](https://v3.wails.io/), [Task](https://taskfile.dev/)

```bash
git clone https://github.com/k3desktop/k3desktop.git
cd k3desktop

# Dev mode (hot-reload)
task dev

# Production build
task build

# Run built binary
task run
```

### Regenerating TypeScript bindings

After changing Go service method signatures:

```bash
wails3 generate bindings -clean=true -ts
```

### Verifying Go packages (without full Wails build)

```bash
go build ./service/... ./dto/...
```

> `go build ./...` from root fails due to iOS build artifacts in the Wails module — use the above instead.

## Architecture

```
k3desktop/
├── main.go           # Wails app entry point, service registration
├── service/          # Go backend services (one file per domain)
│   ├── cluster.go    # Cluster CRUD + async creation events
│   ├── node.go       # Node management
│   ├── registry.go   # Local registry management
│   ├── kubeconfig.go # Kubeconfig export + context management
│   ├── blueprint.go  # Helm blueprint CRUD + deploy
│   ├── profile.go    # Cluster config profile CRUD + import
│   └── logger.go     # Structured log streaming to frontend
├── dto/              # Shared Go/TypeScript data types
├── frontend/         # Svelte SPA
│   ├── src/routes/   # One .svelte file per page
│   ├── src/lib/      # Shared components and stores
│   └── bindings/     # Auto-generated TS bindings (do not edit)
└── website/          # Astro Starlight docs site (GitHub Pages)
```

**Stack:** Go backend → Wails RPC → auto-generated TypeScript bindings → Svelte frontend. Long-running operations emit async Wails events instead of blocking the RPC call.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

[MIT](LICENSE) © Marco Santini
