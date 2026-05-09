# Contributing to K3Desktop

## Before You Start

- Search [existing issues](https://github.com/k3desktop/k3desktop/issues) before opening a new one.
- For large changes, open an issue first to discuss the approach.

## Development Setup

**Requirements:** Go ≥1.21, Node.js ≥22, [Wails v3 CLI](https://v3.wails.io/), [Task](https://taskfile.dev/), Docker

```bash
git clone https://github.com/k3desktop/k3desktop.git
cd k3desktop
task dev          # hot-reload dev server
```

## Making Changes

### Backend (Go)

- One service per domain in `service/`. Each exported method on `*Service` becomes a callable TS binding.
- Add/modify DTOs in `dto/`.
- After any service method signature change, regenerate bindings:

  ```bash
  wails3 generate bindings -clean=true -ts
  ```

- Verify Go packages compile:

  ```bash
  go build ./service/... ./dto/...
  ```

### Frontend (Svelte)

- Routes live in `frontend/src/routes/` — one `.svelte` file per page.
- Type-check before submitting:

  ```bash
  cd frontend && npm run check
  ```
  
- `frontend/bindings/` is auto-generated — never edit it manually.
- Color roles: `brand` = lime `#CDF700` (button bg, use `text-gray-900`); `accent` = cyan `#0DCEFF` (focus rings, links).

### Documentation Website

```bash
cd website && npm run dev    # preview at localhost:4321
npm run build                # verify build passes
```

## Pull Request Guidelines

1. Fork the repo and create a branch from `main`.
2. Keep PRs focused — one feature or fix per PR.
3. Include a clear description of what changed and why.
4. Make sure `go build ./service/... ./dto/...` and `cd frontend && npm run check` pass.
5. For UI changes, describe what you tested manually.

## Commit Style

Use [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add node restart action
fix: prevent registry deletion while cluster is running
docs: update blueprint YAML format example
```

## Reporting Bugs

Open an issue with:

- OS and version
- Steps to reproduce
- Expected vs actual behaviour
- Relevant logs (from the log panel in the app, if applicable)

## License

By contributing you agree your changes will be licensed under the [MIT License](../LICENSE).
