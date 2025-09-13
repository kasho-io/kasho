#!/bin/bash
# Script to migrate v0.2.0 and develop from GHCR to Docker Hub

set -e

echo "========================================="
echo "Migrating Kasho images to Docker Hub"
echo "========================================="

# Login to GitHub Container Registry (if needed)
echo ""
echo "Step 1: Logging into GitHub Container Registry..."
echo "You may need to enter your GitHub username and personal access token"
docker login ghcr.io

# Login to Docker Hub
echo ""
echo "Step 2: Logging into Docker Hub..."
echo "Username: kashoio"
docker login -u kashoio

# Migrate v0.2.0
echo ""
echo "Step 3: Migrating v0.2.0..."
echo "Pulling v0.2.0 from GHCR..."
docker pull ghcr.io/kasho-io/kasho:v0.2.0

echo "Tagging for Docker Hub..."
docker tag ghcr.io/kasho-io/kasho:v0.2.0 kashoio/kasho:v0.2.0
docker tag ghcr.io/kasho-io/kasho:v0.2.0 kashoio/kasho:latest

echo "Pushing v0.2.0 and latest to Docker Hub..."
docker push kashoio/kasho:v0.2.0
docker push kashoio/kasho:latest

# Migrate develop
echo ""
echo "Step 4: Migrating develop..."
echo "Pulling develop from GHCR..."
docker pull ghcr.io/kasho-io/kasho:develop

echo "Tagging develop for Docker Hub..."
docker tag ghcr.io/kasho-io/kasho:develop kashoio/kasho:develop

echo "Pushing develop to Docker Hub..."
docker push kashoio/kasho:develop

echo ""
echo "========================================="
echo "✅ Migration complete!"
echo "========================================="
echo ""
echo "Images now available on Docker Hub:"
echo "  - kashoio/kasho:v0.2.0"
echo "  - kashoio/kasho:latest"
echo "  - kashoio/kasho:develop"