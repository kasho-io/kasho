#!/bin/bash

# Vercel Build Check Script for npm workspace
# This script runs tests before allowing Vercel to build and deploy
# Usage: Set this as your "Build Command" in Vercel settings

set -e

# Determine repository root and app name
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
APP_NAME=$(basename "$PWD")

echo "🔍 Running pre-build checks..."
echo "📂 Repository root: $REPO_DIR"
echo "📦 Building app: $APP_NAME"

# Function to check if we're in a git repository and get the latest commit
check_latest_commit() {
    echo "📍 Checking git status..."
    (
        cd "$REPO_DIR"
        if [ -d ".git" ]; then
            echo "📍 Current commit: $(git rev-parse --short HEAD)"
            echo "📝 Commit message: $(git log -1 --pretty=%B | head -1)"
        else
            echo "⚠️  Not a git repository"
        fi
    )
}

# Function to ensure workspace dependencies are installed
install_dependencies() {
    echo "📦 Ensuring workspace dependencies are installed..."
    (
        cd "$REPO_DIR"
        if [ ! -d "node_modules" ] || [ ! -f "node_modules/.package-lock.json" ]; then
            echo "🔄 Installing workspace dependencies with npm ci..."
            npm ci
        else
            echo "✅ Workspace dependencies already installed"
        fi
    )
}

# Function to run tests for the current app using workspace
run_app_tests() {
    echo "🧪 Running tests for $APP_NAME..."
    (
        cd "$REPO_DIR"
        echo "🔍 Running test:ci for $APP_NAME..."
        npm run test:ci --workspace=apps/$APP_NAME
        echo "✅ All checks passed for $APP_NAME"
    )
}

# Function to check GitHub Actions status for the latest commit (optional)
check_github_actions() {
    echo "🔄 Checking CI status..."
    (
        if [ -n "$GITHUB_TOKEN" ]; then
            echo "🔄 GitHub token found, checking Actions status..."
            # This would require the GitHub CLI or API calls
            # You can implement this if you want to check CI status
            echo "ℹ️  GitHub Actions check skipped (implement if needed)"
        else
            echo "ℹ️  No GitHub token provided, skipping CI check"
        fi
    )
}

# Function to build the app
build_app() {
    echo "🚀 Building $APP_NAME..."
    (
        cd "$REPO_DIR"
        npm run build --workspace=apps/$APP_NAME
        echo "✅ Build completed for $APP_NAME"
    )
}

# Main execution
main() {
    check_latest_commit
    install_dependencies
    run_app_tests
    check_github_actions
    echo "🎉 Pre-build checks completed successfully!"
    
    build_app
    
    echo "🎊 Vercel build process completed successfully!"
}

# Execute main function with all arguments
main "$@"