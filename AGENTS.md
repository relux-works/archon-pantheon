# Agent instructions for curator-runtimes

This repository is public. Everything committed here is visible to anyone.

## Workflow (Axiloom)

- Project tracking (`knowledge`, `taskboard`, `release`) lives in **Axiloom** (local workspace binding in gitignored `.axiloom.json`).
- Start every task with `axiloom guide --with-status` and follow the live project & dimension playbooks.
- Never commit `.axiloom.json` or local workspace state.

## Architecture Principles

- Plugin-oriented harness & vendor architecture (`pkg/plugin`, `pkg/agentic/systems/*`, `pkg/vendorplugin`, `pkg/runtimecatalog`).
- Safe by default with explicit, clean permission modes (including `yolo` when requested by operator policy).
- Lightweight config-driven declaration of runtimes, models, and reasoning levels without recompiling hardcoded catalogs.

## Rules

- Never commit secrets, credentials, personal paths or addresses, or the names of clients or employers.
