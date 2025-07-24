#!/usr/bin/env bash

export VERSION=$GITHUB_COMMIT-test
export KO_DOCKER_REPO=ko.local

set -euo pipefail

source ../../lib/lib.sh

build_anubis_ko
mint_cert relayd

go run ../../cmd/cipra/ --compose-name $(basename $(pwd))

docker compose down -t 1 || :
docker compose rm -f || :
