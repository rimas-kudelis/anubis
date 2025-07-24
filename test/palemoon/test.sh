export VERSION=$GITHUB_COMMIT-test
export KO_DOCKER_REPO=ko.local

set -u

(
  cd ../.. &&
    ko build --platform=all --base-import-paths --tags="latest" --image-user=1000 --image-annotation="" --image-label="" ./cmd/anubis -L
)
