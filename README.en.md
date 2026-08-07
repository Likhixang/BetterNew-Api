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

### 6. Token-level Model Mapping

- **Per-key model redirect**: every API key can define model mappings (e.g. `claude-opus-4-8 → deepseek-v4-pro`). Requests are first checked against the model whitelist using the **original requested model**, then redirected to the mapped model for channel selection, upstream relay, and billing
- **Visual editor**: the key form ships a dual-view editor — Visual (table rows: original → replacement, with model dropdowns) and JSON (1:1 editing with format/copy), plus duplicate-source detection

### 7. Model Merge

- **Inbound model-name normalization**: global rules rewrite requested model names (exact match or Go regex) into a canonical target model (e.g. `cx/deepseek-v4-flash` → `deepseek-v4-flash`), applied to channel selection, upstream relay, and billing
- **Automatic channel resolution (merge reverse matching)**: during channel selection, every channel whose model normalizes to the target under the merge rules joins the candidate pool alongside exact-name channels — a request for `longcat-2.0` can route to a channel that only serves `longcat-2.0-free`, with **no per-channel model_mapping or manual model registration needed** (aligned with AxonHub model-association semantics)
- **All-candidate random selection**: exact-match and merge-match channels share one pool with weighted random selection (no exact-name priority); a failing channel is retried across the other candidates
- **Automatic upstream rename**: when a merge-match channel is picked, the request is relayed upstream under the channel's real model name (e.g. `longcat-2.0` → upstream receives `longcat-2.0-free`); an explicit per-channel model_mapping takes precedence over the automatic derivation
- **Multi-alias rules**: a single rule can carry multiple aliases (one per line), each registered independently against the shared target model
- **Live match preview**: while editing a rule, matching models and channels are previewed per channel in real time; exact matches win, regex rules are evaluated by ascending rule ID
- **Standard table framework**: the rules list uses the standard DataTablePage framework (pagination / rows-per-page / totals) and the standard drawer layout

### 8. Global Prompt Injection

- **Global system prompt**: System Settings → Models & Routing → Global Model Configuration can inject one system prompt into every relayed request
- **Three injection modes**: prepend / append / override the existing system prompt
- **All protocols covered**: OpenAI, Claude, and Gemini requests; channel-level system prompts still apply on top

### 9. Token-level Channel Limits

- **Per-key channel whitelist**: every API Key can define an allowed-channel list (e.g. `Hermes → 8 channels`); requests are routed only to whitelisted channels, everything else is rejected
- **Model + channel double restriction**: channel limits and model limits are independent and both active — the request is validated against the model whitelist first, then channel selection happens within the channel whitelist
- **Empty means allow-all**: leaving channel limits empty allows every channel (same semantics as model limits)
- **Group mechanism hidden**: the legacy Group mechanism is kept in code (as the channel candidate-set source) but fully hidden from the UI — no group traces remain in key editing, channel editing, or channel lists, so no manual group management is needed
- **Auto-disabled channels recoverable**: channels auto-disabled after consecutive failures (status=3) can be manually re-enabled by an admin

### 10. Other Improvements

- **Channel test-model dropdown**: test model is now a real dropdown sourced from the channel's configured models (incl. "Models & Groups"), with adaptive width
- **Form reset fix**: closing the create-channel drawer after a successful submission properly resets the form and advanced-settings panel state
- **Usage-log IP column**: common logs show an IP column (after enabling "Record IP Address" under Profile → Notifications, consume/error logs record and display the client IP)
- **Models page defaults to merges**: the Models page tab order is Merges → Metadata → Deployments, landing on Model Merges by default
- **Fuzzy log filters**: the model-name and username filters on usage logs now match partially — typing `gpt-4` matches `gpt-4o` / `gpt-4-turbo` without the exact name; explicit `%` wildcards still work for fine-grained control

## Quick Start

### Images

| Purpose | Image |
|---|---|
| **Production** (auto-updated) | `ghcr.io/likhixang/betternew-api:latest` |
| **Pinned release** (v1.0.0-rc.23-patch3) | `ghcr.io/likhixang/betternew-api:v1.0.0-rc.23-patch3` |
| **Testing** (patch/test branch) | `ghcr.io/likhixang/betternew-api:patch-test` |

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
