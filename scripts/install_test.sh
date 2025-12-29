#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PASS=0
FAIL=0

setup() {
    TEST_DIR=$(mktemp -d)
    TEST_HOME="$TEST_DIR/home"
    mkdir -p "$TEST_HOME"
    # Create dummy binary
    echo '#!/bin/bash' > "$TEST_DIR/lazywork"
    chmod +x "$TEST_DIR/lazywork"
}

teardown() {
    rm -rf "$TEST_DIR"
}

test_install_creates_binary() {
    setup
    HOME="$TEST_HOME" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    if [[ -x "$TEST_HOME/.local/bin/lazywork" ]]; then
        echo "✓ test_install_creates_binary"
        ((PASS++))
    else
        echo "✗ test_install_creates_binary"
        ((FAIL++))
    fi
    teardown
}

test_shell_integration_bash() {
    setup
    touch "$TEST_HOME/.bashrc"
    HOME="$TEST_HOME" \
    SHELL="/bin/bash" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    if grep -q 'lazywork shell init' "$TEST_HOME/.bashrc"; then
        echo "✓ test_shell_integration_bash"
        ((PASS++))
    else
        echo "✗ test_shell_integration_bash"
        ((FAIL++))
    fi
    teardown
}

test_shell_integration_zsh() {
    setup
    touch "$TEST_HOME/.zshrc"
    HOME="$TEST_HOME" \
    SHELL="/bin/zsh" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    if grep -q 'lazywork shell init zsh' "$TEST_HOME/.zshrc"; then
        echo "✓ test_shell_integration_zsh"
        ((PASS++))
    else
        echo "✗ test_shell_integration_zsh"
        ((FAIL++))
    fi
    teardown
}

test_shell_integration_fish() {
    setup
    mkdir -p "$TEST_HOME/.config/fish"
    touch "$TEST_HOME/.config/fish/config.fish"
    HOME="$TEST_HOME" \
    SHELL="/usr/bin/fish" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    if grep -q 'lazywork shell init fish | source' "$TEST_HOME/.config/fish/config.fish"; then
        echo "✓ test_shell_integration_fish"
        ((PASS++))
    else
        echo "✗ test_shell_integration_fish"
        ((FAIL++))
    fi
    teardown
}

test_idempotent() {
    setup
    touch "$TEST_HOME/.bashrc"
    HOME="$TEST_HOME" \
    SHELL="/bin/bash" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null
    # Run again
    HOME="$TEST_HOME" \
    SHELL="/bin/bash" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    local count
    count=$(grep -c 'lazywork shell init' "$TEST_HOME/.bashrc")
    if [[ "$count" -eq 1 ]]; then
        echo "✓ test_idempotent"
        ((PASS++))
    else
        echo "✗ test_idempotent (found $count occurrences, expected 1)"
        ((FAIL++))
    fi
    teardown
}

test_creates_rc_file_if_missing() {
    setup
    # Don't create .bashrc
    HOME="$TEST_HOME" \
    SHELL="/bin/bash" \
    LAZYWORK_INSTALL_DIR="$TEST_HOME/.local/bin" \
    LAZYWORK_LOCAL_BUILD="$TEST_DIR/lazywork" \
        "$SCRIPT_DIR/install.sh" > /dev/null

    if [[ -f "$TEST_HOME/.bashrc" ]] && grep -q 'lazywork shell init' "$TEST_HOME/.bashrc"; then
        echo "✓ test_creates_rc_file_if_missing"
        ((PASS++))
    else
        echo "✗ test_creates_rc_file_if_missing"
        ((FAIL++))
    fi
    teardown
}

# Run tests
echo "Running install.sh tests..."
echo ""
test_install_creates_binary
test_shell_integration_bash
test_shell_integration_zsh
test_shell_integration_fish
test_idempotent
test_creates_rc_file_if_missing

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ $FAIL -eq 0 ]]
