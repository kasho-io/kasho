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
   - Publishes to GitHub Container Registry (ghcr.io)
   
3. **Deploy Demo** (`deploy-demo.yml`)
   - Triggers: After Publish Container Images completes on main
   - Pulls pre-built `:develop` images from ghcr.io
   - Deploys to DigitalOcean droplet
   - Note: Currently uses `:develop` tag (pre-release phase)

## Container Images

- **Base Image**: `ghcr.io/jeffrey/kasho-base:latest`
  - Contains Go dependencies and build tools
  - Rebuilt only when dependencies change

- **Production Image**: `ghcr.io/jeffrey/kasho:TAG`
  - Tags:
    - `:develop` - Latest from main branch
    - `:v1.0.0` - Semantic version releases
    - `:latest` - Latest stable release
    - `:sha-abc123` - Commit SHA for traceability

## Local Development

The development environment (`environments/development/`) continues to use local builds with hot-reload via air. Only the demo and production deployments use the published container images.