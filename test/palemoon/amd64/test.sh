#!/usr/bin/env bash

set -euo pipefail

source ../../lib/lib.sh

build_anubis_ko
mint_cert relayd

go run ../../cmd/cipra/ --compose-name $(basename $(pwd))

docker compose down -t 1 || :
docker compose rm -f || :
