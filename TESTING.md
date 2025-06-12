# Testing Strategy

This document outlines the task-centric testing approach for the Kasho monorepo using [Task](https://taskfile.dev/).

## Quick Start

```bash
# Run all tests
task test

# Run only Go service tests
task test:go

# Run only Next.js app tests
task test:apps

# Run linting for everything
task lint
```

## Available Test Tasks

### Core Testing Tasks

| Task | Description |
|------|-------------|
| `task test` | Run all tests (Go services + Next.js apps) |
| `task test:go` | Run tests for all Go services |
| `task test:apps` | Run tests for all Next.js apps |

### Individual Service Tests

| Task | Description |
|------|-------------|
| `task test:pg-change-stream` | Test pg-change-stream service with coverage |
| `task test:pg-translicator` | Test pg-translicator service with coverage |
| `task test:pg-bootstrap-sync` | Test pg-bootstrap-sync tool with coverage |
| `task test:kvbuffer` | Test shared kvbuffer package with coverage |
| `task test:env-template` | Test env-template tool |
| `task test:proto` | Test proto package |
| `task test:app-demo` | Test demo Next.js app |
| `task test:app-homepage` | Test homepage Next.js app |

### Linting Tasks

| Task | Description |
|------|-------------|
| `task lint` | Run linting for all code |
| `task lint:go` | Lint Go services (`go vet`) |
| `task lint:apps` | Lint all Next.js apps |
| `task lint:app-demo` | Lint demo app only |
| `task lint:app-homepage` | Lint homepage app only |

### CI/CD Tasks

| Task | Description |
|------|-------------|
| `task ci:test` | Run tests with CI-specific options |
| `task ci:test:go` | Run Go tests with coverage for CI |
| `task ci:test:apps` | Run Next.js tests for CI (includes `npm ci`) |

## Testing Philosophy

### Go Services and Tools
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test service interactions (with Redis, databases)  
- **Coverage**: Generate coverage reports with each test run
- **Race Detection**: All tests run with `-race` flag
- **Comprehensive Coverage**: Tests cover services (pg-change-stream, pg-translicator), tools (pg-bootstrap-sync, env-template), and shared packages (kvbuffer, proto)

### Next.js Apps
- **Type Checking**: Ensures TypeScript compilation succeeds
- **Linting**: Enforces code style and catches common issues
- **Build Validation**: Verifies the app can be built successfully

## Local Development Workflow

1. **Before committing**:
   ```bash
   task test
   task lint
   ```

2. **Testing specific changes**:
   ```bash
   # If you changed Go services
   task test:go
   
   # If you changed Next.js apps
   task test:apps
   ```

3. **Quick feedback loop**:
   ```bash
   # Test just one service
   task test:pg-change-stream
   
   # Test just one app
   task test:app-demo
   ```

## GitHub Actions Integration

The GitHub Actions workflows use these tasks:

```yaml
# In .github/workflows/test-go-services.yml
- name: Run Go tests
  run: task ci:test:go

# In .github/workflows/test-nextjs-app.yml  
- name: Run app tests
  run: task ci:test:apps
```

### Deployment Integration

**Demo Environment**: The `deploy-demo.yml` workflow waits for tests to pass before deploying:

1. **Test Suite** runs first (`test.yml`)
2. **Deploy Demo** waits for test completion before deploying to DigitalOcean
3. Deployment only occurs if:
   - Tests pass ✅
   - Relevant files changed (services, demo environment, tools)

This ensures the demo environment only receives tested, working code.

## Vercel Integration

### Setup for Each App

1. **Vercel Project Settings**:
   - Build Command: `../../scripts/vercel-build-check.sh`
   - Install Command: `npm ci`

2. **What the build script does**:
   ```bash
   # The script runs these task equivalents:
   npm run lint      # Same as: task lint:app-demo
   npx tsc --noEmit  # Type checking
   npm run build     # Final build step
   ```

### Vercel Configuration Files

Each app has a `vercel.json`:
```json
{
  "buildCommand": "../../scripts/vercel-build-check.sh",
  "devCommand": "npm run dev",
  "installCommand": "npm ci",
  "framework": "nextjs"
}
```

## Adding New Tests

### For Go Services

1. Create `*_test.go` files alongside your code
2. Tests will automatically be picked up by:
   - `task test:pg-change-stream`
   - `task test:pg-translicator`
   - `task test:proto`

### For Next.js Apps

1. Install testing dependencies:
   ```bash
   cd apps/demo
   npm install --save-dev jest @testing-library/react
   ```

2. Update the `test` script in `package.json`:
   ```json
   {
     "scripts": {
       "test": "npm run lint && npx tsc --noEmit && jest"
     }
   }
   ```

3. The test will be run by `task test:app-demo`

## Coverage Reports

Go services generate coverage reports automatically:

```bash
# View coverage for a specific service
task test:pg-change-stream
# This generates: services/pg-change-stream/coverage.out

# View coverage in browser (after running tests)
cd services/pg-change-stream
go tool cover -html=coverage.out
```

## Troubleshooting

### Common Issues

**Task not found**:
```bash
# Make sure Task is installed
brew install go-task/tap/go-task  # macOS
# or
go install github.com/go-task/task/v3/cmd/task@latest
```

**Tests fail locally but pass in CI**:
- Check Go version: `go version`
- Check Node version: `node --version`
- Ensure dependencies are up to date

**App tests fail on type checking**:
```bash
# Check TypeScript issues
cd apps/demo
npx tsc --noEmit
```

### Getting Help

```bash
# List all available tasks
task --list

# See what a specific task does
task --summary test:go
```

## Integration with Development Workflow

### Starting Development

```bash
# Start development environment
task dev

# Start a specific app
task dev:app-demo
task dev:app-homepage
```

### Before Pull Requests

```bash
# Full validation
task test lint

# Quick check for Go changes
task test:go lint:go

# Quick check for app changes  
task test:apps lint:apps
```

This task-centric approach ensures consistency between local development, CI/CD, and deployment processes while providing fast feedback for developers.