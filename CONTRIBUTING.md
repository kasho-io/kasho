# Contributing to Kasho

This guide helps you get set up for local development and explains the project's tooling conventions.

## Prerequisites

Before you begin, ensure you have installed:

- **Docker** - Required for running Go services and databases
- **Go 1.24+** - For Go tooling and IDE support
- **Node.js & npm** - For Next.js apps
- **Task** - Task runner for development workflows

Install Task:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

## Quick Start

```bash
# 1. Clone and enter the repo
git clone https://github.com/kasho-io/kasho.git
cd kasho

# 2. Install git hooks (for automatic formatting)
./scripts/development/install-git-hooks.sh

# 3. Install npm dependencies
npm install

# 4. Start the development environment (Docker)
task dev

# 5. In another terminal, bootstrap the replica database
task dev:bootstrap

# 6. Run an app (e.g., the demo)
task dev:app:demo
```

## When to Use `task` vs `npm`

This project uses both Task (for orchestration) and npm (for Node.js apps). Here's when to use each:

### Use `task` for:

| Task | Command |
|------|---------|
| Start Docker dev environment | `task dev` |
| Stop Docker environment | `task dev:stop` |
| Reset environment (fresh start) | `task dev:reset` |
| Bootstrap replica database | `task dev:bootstrap` |
| Build Docker images | `task build` |
| Run a specific app | `task dev:app:demo` |
| Run ALL tests (Go + apps) | `task test` |
| Run ALL linting (Go + apps) | `task lint` |
| Format all apps | `task apps:prettier` |

### Use `npm` for:

| Task | Command |
|------|---------|
| Install dependencies | `npm install` |
| Run tests for one app | `npm test --workspace=apps/demo` |
| Type-check all apps | `npm run type-check` |
| Lint one app | `npm run lint --workspace=apps/demo` |

### Rule of Thumb

- **`task`** = orchestration, Docker, Go, multi-component operations
- **`npm`** = Node.js/app-specific operations, dependency management

When in doubt, run `task` by itself to see available commands.

## Project Structure

```
kasho/
├── apps/                    # Next.js frontend applications
│   ├── homepage/            # Landing page (port 3000)
│   ├── demo/                # Interactive demo (port 3001)
│   └── docs/                # Documentation site (port 3002)
├── services/                # Go backend services
│   ├── pg-change-stream/    # Captures database changes
│   └── pg-translicator/     # Transforms and applies changes
├── tools/                   # CLI utilities
│   ├── runtime/             # Production tools (pg-bootstrap-sync, env-template)
│   └── development/         # Dev tools (generate-fake-saas-data)
├── pkg/                     # Shared Go packages
├── environments/            # Docker Compose configurations
│   ├── pg-development/      # Local dev environment
│   └── pg-demo/             # Production-like demo
├── proto/                   # Protocol buffer definitions
└── sql/                     # Database setup scripts
```

## Development Workflows

### Working on Go Services

Go services run inside Docker with hot-reload via [air](https://github.com/air-verse/air):

```bash
# Start the environment (services auto-reload on file changes)
task dev

# Run Go tests
task test:go

# Test a specific service
task test:service:pg-change-stream

# Lint Go code
task lint:go
```

### Working on Next.js Apps

Apps run locally (not in Docker) and use npm workspaces:

```bash
# Install dependencies (run from repo root)
npm install

# Start an app in dev mode
task dev:app:demo      # or: npm run dev --workspace=apps/demo

# Run tests for an app
npm test --workspace=apps/demo

# Type-check an app
npm run type-check --workspace=apps/demo

# Format code
npm run prettier --workspace=apps/demo
```

### Environment Configuration

Environment variables are managed separately for Docker services and apps:

- **Docker services**: `environments/pg-development/.env` and `environments/pg-demo/.env`
- **Next.js apps**: Use standard Next.js `.env.local` files within each app directory

See `.env.example` files in each environment directory for available variables.

## Git Hooks

The project uses git hooks for code quality. Install them with:

```bash
./scripts/development/install-git-hooks.sh
```

### Pre-commit Hook

When you commit changes to files in `apps/`, the pre-commit hook will:

1. Run Prettier to format changed files
2. Run ESLint to check for issues
3. Stage any formatting changes automatically

**Note**: The hook requires `task` to be installed. If task is not found, it will skip formatting with a warning.

### Manual Formatting

If you prefer to format manually before committing:

```bash
# Format all apps
task apps:prettier

# Format a specific app
task apps:prettier:demo

# Check formatting without writing
task apps:prettier:check
```

## Testing

### Go Services

Go services have comprehensive test coverage:

```bash
task test:go                           # All Go tests
task test:service:pg-change-stream     # Specific service
task test:pkg:kvbuffer                 # Specific package
```

### Next.js Apps

Apps currently have placeholder test scripts. Run them with:

```bash
task test:apps                         # All apps
npm test --workspace=apps/demo         # Specific app
```

## Code Style

### Go

- Follow standard Go conventions
- Run `task lint:go` before committing
- Tests live alongside code in `*_test.go` files

### TypeScript/JavaScript

- Prettier handles formatting (automated via git hooks)
- ESLint catches common issues
- Run `npm run type-check` to verify TypeScript

## Common Issues

### "Task command not found" in git hooks

Install Task and ensure it's in your PATH:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Docker services won't start

```bash
# Reset the environment (removes volumes)
task dev:reset
```

### npm install fails

```bash
# Clean and reinstall
task apps:deps:clean
npm install
```

## Pull Request Guidelines

1. Create a feature branch from `main`
2. Make your changes
3. Ensure tests pass: `task test`
4. Ensure linting passes: `task lint`
5. Commit with a descriptive message
6. Push and open a PR against `main`
7. PRs are merged with `--rebase` (not squash)
