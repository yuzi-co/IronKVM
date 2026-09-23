#!/bin/sh
# Run the viewer's Go tests: the token it mints must carry the account's
# 64-bit token version exactly, and verify with the secret.
cd "$(dirname "$0")" || exit 1
command -v go >/dev/null 2>&1 || { echo "needs go on PATH"; exit 2; }
go test ./...
