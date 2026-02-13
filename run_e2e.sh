#!/bin/bash
# Run E2E tests with API key from ~/.eliza_token
export OPENAI_API_KEY=$(cat ~/.eliza_token)
exec go test -tags=e2e -v -timeout 120s "$@" ./internal/llm/
