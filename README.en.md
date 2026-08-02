<div align="center">

![new-api](/web/public/logo.png)

# BetterNew-Api

🍥 **A customized fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api)**

<p align="center">
  <a href="./README.md">简体中文</a> |
  <strong>English</strong>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/QuantumNous/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/QuantumNous/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Likhixang/BetterNew-Api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Likhixang/BetterNew-Api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/QuantumNous/new-api">
    <img src="https://img.shields.io/badge/upstream-QuantumNous%2Fnew--api-blue" alt="upstream">
  </a>
</p>

</div>

## Introduction

This repository is a fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api) — a next-generation LLM gateway and AI asset management system. It retains all upstream features while introducing optimizations to channel management and model maintenance workflows.

- **Forked from**: [QuantumNous/new-api](https://github.com/QuantumNous/new-api)
- **License**: AGPL-3.0 (upstream copyright notice and license text are retained)

## Differences from Upstream

### 1. Channel Key Management

- **Inline key display**: Keys are embedded directly in the input field, hidden by default, with a toggle for plaintext view and one-click copy
- **Masked key preview**: When hidden, each line shows only the first 4 and last 4 characters (`sk-a1b2****9z8y`) instead of being fully masked with dots
- **WYSIWYG editing**: Adding a key on a new line appends it; selecting all and rewriting replaces the list — no manual mode selection required
- **Key verification removed**: Viewing keys no longer requires a verification code
- **OpenAI organization ID removed**: The organization ID field is no longer required in channel configuration
- **Mobile layout adaptation**: The key input uses a fixed width with internal scrolling for overflow, preventing form layout breakage

### 2. Upstream Model Detection

- **Model whitelist (allowed models)**: Upstream models are only considered addable when they match the whitelist — supports exact names and `regex:` patterns
- **Model blacklist (ignored models)**: Excludes upstream models that match the blacklist
- **Combined filtering**: The whitelist and blacklist can be configured independently or applied together; filtering is applied whitelist-first, then blacklist
- **Enabled by default**: Upstream model update detection is enabled by default

### 3. Fetch Models Dialog

- **Select All / Deselect All**: Bulk selection of the currently visible models
- **Only show not added**: Filters out already-added models to focus on pending additions

### 4. UI & Build

- **Interface languages**: Simplified Chinese and English only
- **Automated builds**: GitHub Actions builds and pushes images to GHCR, triggered on pushes to the `main` / `likhixang/patch` branches or on `v*` tags

### 5. PWA Support

- **Installable**: can be added to the home screen / desktop and runs in a standalone window
- **Offline-ready**: a Service Worker precaches all built static assets so the dashboard opens offline; API calls (`/api`, `/pg`, `/mj`) always hit the network and are never cached
- **Dynamic app name & icons**: the manifest is served dynamically (`/manifest.webmanifest`); the app name follows the system name configured in settings (same source as the page title), and the PWA icons follow the configured logo URL (built-in icons when unset). Chrome/Edge pick up changes automatically; on iOS, re-add the icon to the home screen
- **Implementation notes**: the Service Worker is generated at build time with Google's official [workbox-build](https://developer.chrome.com/docs/workbox/) (`web/scripts/generate-sw.mjs`), and icons are generated from `logo.png` via `web/scripts/generate-pwa-assets.mjs` (sharp). Manual integration with official libraries — not a third-party PWA plugin

## Quick Start

```bash
docker run -d --restart always --name betternew-api \
  -p 3000:3000 \
  -v ./data:/data \
  ghcr.io/likhixang/betternew-api:latest
```

Alternatively, use the `docker-compose.yml` in this repository:

```bash
docker compose up -d
```

## Documentation

For full feature and deployment documentation, refer to the upstream project: [QuantumNous/new-api](https://github.com/QuantumNous/new-api)

## License

[AGPL-3.0](LICENSE) — a modified version of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), retaining the upstream copyright notice and license text.
