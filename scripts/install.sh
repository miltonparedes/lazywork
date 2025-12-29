#!/usr/bin/env bash
set -euo pipefail

REPO="miltonparedes/lazywork"
INSTALL_DIR="${LAZYWORK_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="lazywork"

VERSION=""
PRERELEASE=false

usage() {
    cat <<EOF
Install LazyWork

Usage: install.sh [options]

Options:
    --version <ver>   Install specific version (e.g., v0.1.0-alpha)
    --prerelease      Include pre-releases when fetching latest
    --help            Show this help

Examples:
    curl -fsSL .../install.sh | bash
    curl -fsSL .../install.sh | bash -s -- --prerelease
    curl -fsSL .../install.sh | bash -s -- --version v0.1.0-alpha
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --prerelease)
                PRERELEASE=true
                shift
                ;;
            --help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1" >&2
                usage
                exit 1
                ;;
        esac
    done
}

detect_shell() {
    basename "${SHELL:-/bin/bash}"
}

get_rc_file() {
    case "$1" in
        fish) echo "$HOME/.config/fish/config.fish" ;;
        zsh)  echo "$HOME/.zshrc" ;;
        *)    echo "$HOME/.bashrc" ;;
    esac
}

get_init_line() {
    case "$1" in
        fish) echo 'lazywork shell init fish | source' ;;
        *)    echo "eval \"\$(lazywork shell init $1)\"" ;;
    esac
}

get_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux)  echo "linux" ;;
        darwin) echo "darwin" ;;
        *)      echo "Unsupported OS: $os" >&2; exit 1 ;;
    esac
}

get_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64)  echo "amd64" ;;
        amd64)   echo "amd64" ;;
        aarch64) echo "arm64" ;;
        arm64)   echo "arm64" ;;
        *)       echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac
}

get_latest_version() {
    if [[ "$PRERELEASE" == true ]]; then
        curl -fsSL "https://api.github.com/repos/$REPO/releases" | grep '"tag_name"' | head -1 | cut -d'"' -f4
    else
        curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4
    fi
}

download_release() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local tmpdir

    tmpdir=$(mktemp -d)

    local archive_name="lazywork_${version#v}_${os}_${arch}.tar.gz"
    local url="https://github.com/$REPO/releases/download/$version/$archive_name"

    echo "Downloading $url..."
    if ! curl -fsSL "$url" -o "$tmpdir/lazywork.tar.gz"; then
        rm -rf "$tmpdir"
        echo "Failed to download release. Check if version $version exists." >&2
        exit 1
    fi

    tar -xzf "$tmpdir/lazywork.tar.gz" -C "$tmpdir"
    mv "$tmpdir/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$tmpdir"
}

install_binary() {
    mkdir -p "$INSTALL_DIR"

    if [[ -n "${LAZYWORK_LOCAL_BUILD:-}" ]]; then
        cp "$LAZYWORK_LOCAL_BUILD" "$INSTALL_DIR/$BINARY_NAME"
    else
        local version="$VERSION"
        local os arch

        os=$(get_os)
        arch=$(get_arch)

        if [[ -z "$version" ]]; then
            echo "Fetching latest version..."
            version=$(get_latest_version)
            if [[ -z "$version" ]]; then
                echo "No stable releases found." >&2
                echo "Try: --prerelease to install pre-release" >&2
                exit 1
            fi
        fi

        echo "Installing lazywork $version ($os/$arch)..."
        download_release "$version" "$os" "$arch"
    fi

    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "Installed $BINARY_NAME to $INSTALL_DIR"
}

install_shell_integration() {
    local shell_type rc_file init_line
    shell_type=$(detect_shell)
    rc_file=$(get_rc_file "$shell_type")
    init_line=$(get_init_line "$shell_type")

    if grep -q "lazywork shell init" "$rc_file" 2>/dev/null; then
        echo "Shell integration already configured"
        return 0
    fi

    mkdir -p "$(dirname "$rc_file")"
    {
        echo ""
        echo "# LazyWork shell integration"
        echo "$init_line"
    } >> "$rc_file"

    echo "Added shell integration to $rc_file"
}

main() {
    parse_args "$@"
    echo "Installing LazyWork..."
    install_binary
    install_shell_integration
    echo ""
    echo "Done! Restart your shell or run: source $(get_rc_file "$(detect_shell)")"
}

main "$@"
