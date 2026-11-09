# GitHub Actions Workflows

## Workflow Chain

The workflows are designed to run in this order:

1. **Test Suite** (`test.yml`)
   - Triggers: On push to main, on PRs
   - Runs all tests (Go services, Next.js apps)
   
2. **Publish Container Images** (`publish-containers.yml`)
   - Triggers: 
     - After Test Suite passes on main → publishes as `:develop`
     - On version tag push (v*) → publishes as `:v1.0.0` and `:latest`
   - Publishes to Docker Hub

The local demo (`environments/pg-demo/`) runs the `:develop` images, so publishing them keeps the demo current.

## Container Images

- **Base Image**: `kashoio/kasho-base:latest`
  - Contains Go dependencies and build tools
  - Rebuilt only when dependencies change

- **Production Image**: `kashoio/kasho:TAG`
  - Tags:
    - `:develop` - Latest from main branch
    - `:v1.0.0` - Semantic version releases
    - `:latest` - Latest stable release
    - `:sha-abc123` - Commit SHA for traceability

## Local Development

The development environment (`environments/pg-development/`) continues to use local builds with hot-reload via air. Only the local demo and production deployments use the published container images.