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

## 📌 About

This repository is a **fork of [QuantumNous/new-api](https://github.com/QuantumNous/new-api)** (a next-generation LLM gateway and AI asset management system), customized for a better day-to-day management experience.

**Forked from**: [QuantumNous/new-api](https://github.com/QuantumNous/new-api)
**License**: AGPL-3.0 (this fork keeps the upstream copyright notice and license text)

## ✨ Key Differences from Upstream

### 1. Channel Key Management

- **Inline key display**: Keys live directly in the input field — hidden by default, revealed with an eye toggle, copyable with one click
- **Masked preview**: When hidden, each key line shows first 4 + last 4 chars (`sk-a1b2****9z8y`) instead of a dotted field
- **WYSIWYG editing**: Add a key on a new line = append; select-all and rewrite = replace. No mode selector needed
- **No reveal verification**: Viewing a key no longer requires a verification code
- **OpenAI organization ID removed**: No longer required in channel config
- **Mobile-friendly**: Long keys no longer break the layout — fixed width with internal scroll

### 2. Upstream Model Detection

- **Allowed models (whitelist)**: Only models matching the list are considered addable — supports exact names and `regex:` patterns
- **Ignored models (blacklist)**: Models matching the list are excluded (upstream feature)
- **Combined**: Whitelist and blacklist can be used independently or together (whitelist first, then blacklist)
- **Enabled by default**: Upstream model update check is on by default

### 3. Fetch Models Dialog

- **Select All / Deselect All**: Bulk toggle the currently visible model list
- **Only show not added**: Filter out models that are already added

### 4. UI & Build

- **Interface languages**: Simplified Chinese and English only
- **Automated builds**: GitHub Actions builds and pushes images to GHCR on pushes to `main` / `likhixang/patch` and on `v*` tags

## 🚀 Quick Start

```bash
docker run -d --restart always --name betternew-api \
  -p 3000:3000 \
  -v ./data:/data \
  ghcr.io/likhixang/betternew-api:latest
```

Or use the `docker-compose.yml` in this repo:

```bash
docker compose up -d
```

## 📚 Documentation

Full feature and deployment docs live upstream: [QuantumNous/new-api](https://github.com/QuantumNous/new-api)

## 📄 License

[AGPL-3.0](LICENSE) — a modified version of [QuantumNous/new-api](https://github.com/QuantumNous/new-api), keeping the upstream copyright notice and license text.
