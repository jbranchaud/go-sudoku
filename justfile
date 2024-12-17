# List available commands
default:
    @just --list

# Check if brew is installed
_check-brew:
    #!/usr/bin/env bash
    if ! command -v brew &> /dev/null; then
        echo "Error: Homebrew is not installed"
        echo "Please install from https://brew.sh"
        exit 1
    fi

# Install all required development dependencies
setup: _check-brew _install-deps _install-go-tools

brew_deps := '''
    go
    sqlite3
'''

# Install brew dependencies
_install-deps:
    #!/usr/bin/env bash
    deps=$(echo '{{brew_deps}}' | tr -s '[:space:]' ' ' | xargs)
    for pkg in $deps; do
        if ! brew list $pkg &>/dev/null; then
            echo "Installing $pkg..."
            brew install $pkg
        else
            echo "✓ $pkg already installed"
        fi
    done

# Install Go development tools
_install-go-tools:
    go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify all required tools are installed correctly
verify:
    #!/usr/bin/env bash
    echo "Verifying development dependencies..."

    which go >/dev/null 2>&1 || { echo "Error: go is not installed"; exit 1; }
    which sqlite3 >/dev/null 2>&1 || { echo "Error: sqlite3 is not installed"; exit 1; }
    which goose >/dev/null 2>&1 || { echo "Error: goose is not installed"; exit 1; }

    echo "✓ All development tools are installed"
