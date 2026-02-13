#!/bin/bash
# E2E tests for prefer-single-tool prompt changes
# Runs qx with real LLM and checks output

export OPENAI_API_KEY=$(cat ~/.eliza_token)

echo "=== Test 1: find go files and count lines ==="
echo "Expected: find -exec or wc, not find | xargs wc"
./qx "find all go files and count lines in each" --no-interactive 2>&1
echo ""

echo "=== Test 2: extract name from package.json ==="
echo "Expected: jq .name, not grep name | awk"
./qx "extract name field from package.json" --no-interactive 2>&1
echo ""

echo "=== Test 3: pipe JSON - get active users ==="
echo "Expected: jq select, not grep true"
echo '{"users":[{"name":"alice","active":true},{"name":"bob","active":false}]}' | ./qx "get active users" --no-interactive 2>&1
echo ""

echo "=== Test 4: pipe tabular data - CPU processes ==="
echo "Expected: awk with condition, not grep by pattern"
ps aux | head -20 | ./qx "find processes using more than 1% CPU" --no-interactive 2>&1
echo ""
