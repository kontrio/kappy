#!/usr/bin/env sh
set -ex

TAR_FILE="/tmp/goreleaser.tar.gz"
RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
TMPDIR="$(mktemp -d)"

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASES_URL/latest" | 
    rev | 
    cut -f1 -d'/'| 
    rev
}

install() {
  test -z "$VERSION" && VERSION="$(last_version)"
  test -z "$VERSION" && {
    echo "Unable to get goreleaser version." >&2
    exit 1
  }
  rm -f "$TAR_FILE"
  curl -s -L -o "$TAR_FILE" \
    "$RELEASES_URL/download/$VERSION/goreleaser_$(uname -s)_$(uname -m).tar.gz"
  tar -xf "$TAR_FILE" -C "$TMPDIR"
}

install
echo $TMPDIR
"${TMPDIR}/goreleaser" "$@"
