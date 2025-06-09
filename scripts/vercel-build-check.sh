#!/bin/bash

# Vercel Build Check Script
# This script runs tests before allowing Vercel to build and deploy
# Usage: Set this as your "Build Command" in Vercel settings

set -e

echo "🔍 Running pre-build checks..."

# Get the current directory name to determine which app we're building
APP_NAME=$(basename "$PWD")
echo "📦 Building app: $APP_NAME"

# Function to check if we're in a git repository and get the latest commit
check_latest_commit() {
    if [ -d ".git" ]; then
        echo "📍 Current commit: $(git rev-parse --short HEAD)"
        echo "📝 Commit message: $(git log -1 --pretty=%B | head -1)"
    fi
}

# Function to run tests for the current app
run_app_tests() {
    echo "🧪 Running tests for $APP_NAME..."
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "📦 Installing dependencies..."
        npm ci
    fi
    
    # Run linting
    echo "🔍 Running linter..."
    npm run lint
    
    # Run type checking
    echo "🔎 Running type checker..."
    npx tsc --noEmit
    
    # Run tests if test script exists
    if npm run | grep -q "test"; then
        echo "🧪 Running unit tests..."
        npm run test
    else
        echo "ℹ️  No test script found, skipping unit tests"
    fi
    
    echo "✅ All checks passed for $APP_NAME"
}

# Function to check GitHub Actions status for the latest commit (optional)
check_github_actions() {
    if [ -n "$GITHUB_TOKEN" ]; then
        echo "🔄 Checking GitHub Actions status..."
        # This would require the GitHub CLI or API calls
        # You can implement this if you want to check CI status
        echo "ℹ️  GitHub Actions check skipped (implement if needed)"
    fi
}

# Main execution
main() {
    check_latest_commit
    run_app_tests
    check_github_actions
    
    echo ""
    echo "🎉 Pre-build checks completed successfully!"
    echo "🚀 Proceeding with Vercel build..."
    
    # Run the actual build command
    npm run build
}

# Execute main function
main "$@"