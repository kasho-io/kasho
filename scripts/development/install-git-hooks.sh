#!/bin/bash

# Script to install git hooks for the project

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "🪝 Installing git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Install pre-commit hook
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash

# Pre-commit hook to run prettier on specific apps that have staged changes

# Get all staged files in apps/ directory
staged_app_files=$(git diff --cached --name-only | grep "^apps/")

if [ -n "$staged_app_files" ]; then
    # Find unique app directories that have changes
    changed_apps=$(echo "$staged_app_files" | cut -d'/' -f2 | sort -u)
    
    echo "🎨 Found changes in the following apps:"
    echo "$changed_apps" | sed 's/^/  - /'
    
    # Check if task command is available (check common locations)
    TASK_CMD=""
    if command -v task >/dev/null 2>&1; then
        TASK_CMD="task"
    elif [ -f "$HOME/.local/bin/task" ]; then
        TASK_CMD="$HOME/.local/bin/task"
    elif [ -f "$HOME/go/bin/task" ]; then
        TASK_CMD="$HOME/go/bin/task"
    elif [ -f "/usr/local/bin/task" ]; then
        TASK_CMD="/usr/local/bin/task"
    fi
    
    if [ -n "$TASK_CMD" ]; then
        # Run prettier on each changed app individually
        for app in $changed_apps; do
            if [ -d "apps/$app" ]; then
                echo "🎨 Running prettier on $app app..."
                if $TASK_CMD apps:prettier:$app; then
                    echo "✅ Prettier formatting completed for $app"
                else
                    echo "❌ Prettier formatting failed for $app"
                    exit 1
                fi
            else
                echo "⚠️  Directory apps/$app not found, skipping..."
            fi
        done
        
        # Stage any files that were formatted in the changed apps
        for app in $changed_apps; do
            echo "$staged_app_files" | grep "^apps/$app/" | while read -r file; do
                if [ -f "$file" ]; then
                    git add "$file"
                fi
            done
        done
        
        echo "✅ All prettier formatting completed successfully"
    else
        echo "⚠️  Task command not found, skipping prettier formatting"
        echo "   Install task: https://taskfile.dev/installation/"
    fi
else
    echo "ℹ️  No changes in apps/ directory, skipping prettier formatting"
fi

exit 0
EOF

# Make hook executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The following hooks have been installed:"
echo "  - pre-commit: Runs prettier on specific apps when files in those apps are changed"
echo ""
echo "To run this script on other machines:"
echo "  ./scripts/development/install-git-hooks.sh"