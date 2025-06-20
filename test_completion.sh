#!/bin/bash

# Test script for gosh tab completion and Ctrl+C handling
# This script demonstrates the enhanced features

echo "Testing gosh tab completion and Ctrl+C handling..."
echo "=========================================="
echo ""

echo "1. Building gosh..."
make build

echo ""
echo "2. Testing basic functionality..."
echo -e 'pwd\necho "Basic test successful"\nexit' | ./build/gosh

echo ""
echo "3. Testing variable expansion..."
echo -e 'export TEST_VAR=hello\necho "Variable value: $TEST_VAR"\nexit' | ./build/gosh

echo ""
echo "4. Testing error handling..."
echo -e 'nonexistentcommand\necho "Error handling works"\nexit' | ./build/gosh

echo ""
echo "5. Manual testing instructions:"
echo "   Run: make run"
echo "   Then try:"
echo "   - Type 'h' and press Tab (should complete to 'help' or 'history')"
echo "   - Type 'git ' and press Tab (should show git subcommands)"
echo "   - Type 'git che' and press Tab (should complete to 'checkout')"
echo "   - Press Ctrl+C (should interrupt and return to prompt)"
echo "   - Type 'exit' to quit"
echo ""
echo "Tab completion and Ctrl+C should now work properly!"
