# Repository Guidelines

## Project Structure & Module Organization
- `backend/` hosts Go services with `cmd/` entrypoints, HTTP logic in `handler/`, business flows in `usecase/`, and persistence adapters under `store/`; apply schema updates via `store/pg/migration/`.
- `web/` is a pnpm workspace: `admin/` is the Vite-powered console, `app/` is the public Next.js site, and `packages/` contains shared UI, icon, and theme libraries. Run workspace scripts from `web/`.
- `sdk/rag/` mirrors backend data contracts for retrieval tooling, and `images/` stores marketing assets used in documentation.
- Keep Go specs next to their packages (`*_test.go`) and colocate frontend tests with the component or route they cover.

## Build, Test & Development Commands
- `cd backend && make lint` regenerates Swagger/Wire code and runs `golangci-lint`.
- `cd backend && go test ./...` executes the Go suite locally; run before pushing.
- `cd web && pnpm install` installs workspace dependencies (pnpm enforced by `preinstall`).
- `cd web && pnpm dev` starts both frontends; scope with `pnpm --filter admin dev` or `pnpm --filter app dev`.
- `pnpm --filter panda-wiki-admin build` and `pnpm --filter panda-wiki-app build` produce production bundles; build the admin Docker image via `make image` inside `web/admin/`.

## Coding Style & Naming Conventions
- Go code must remain `gofmt`/`goimports` clean; exported symbols use mixedCase, migrations use snake_case filenames.
- Frontend TypeScript follows Prettier defaults (2-space indent, single quotes, trailing commas). Prefer function components and React hooks.
- Keep package and file names in kebab-case; share UI logic through `web/packages/` instead of local duplicates.

## Testing Guidelines
- Favor table-driven Go tests and add fixtures under `backend/store/...` when touching migrations.
- For UI work, add Vitest or integration coverage where practical; at minimum, document manual checks and run `pnpm --filter app lint` or package-specific linters.
- Always run `go test ./...` and pertinent lint commands before requesting review.

## Commit & Pull Request Guidelines
- Use conventional prefixes (`feat:`, `fix:`, `pref:`, `chore:`) with concise English summaries; add brief Chinese context only when it clarifies behavior.
- PRs should link issues, describe backend/frontend impacts, list executed commands, and attach screenshots or recordings for UI-visible changes. Wait for CI to pass before seeking review.
