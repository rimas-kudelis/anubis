#!/usr/bin/env bash

set -euo pipefail
set -x

(cd amd64 && ./test.sh && docker compose down -t0 && docker compose rm)
(cd i386 && ./test.sh && docker compose down -t0 && docker compose rm)
