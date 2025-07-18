#!/usr/bin/env bash

set -euo pipefail

function cleanup() {
  pkill -P $$
}

trap cleanup EXIT SIGINT

# Build static assets
(cd ../.. && npm ci && npm run assets)

# Spawn three jobs:

# HTTP daemon that listens over a unix socket (implicitly ./unixhttpd.sock)
go run ../cmd/unixhttpd &

go tool anubis \
  --policy-fname ./anubis.yaml \
  --use-remote-address \
  --target=unix://$(pwd)/unixhttpd.sock &

go run ../cmd/backoff-retry node ./test.mjs
