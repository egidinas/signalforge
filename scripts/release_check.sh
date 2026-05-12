#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

GO_BIN=${GO_BIN:-go}
GOFMT_BIN=${GOFMT_BIN:-gofmt}

if [ -n "$($GOFMT_BIN -l $(find . -name '*.go' -type f | sort))" ]; then
	printf '%s\n' "gofmt needed" >&2
	exit 1
fi

$GO_BIN test ./...
$GO_BIN vet ./...
$GO_BIN list ./...
$GO_BIN list -m all

if grep -q '^replace[[:space:]]' go.mod; then
	printf '%s\n' "replace directive present in go.mod" >&2
	exit 1
fi

scan_hits=$(/bin/grep -RIn -i -E 'github\.com/signalforge|signalforge/signalforge|pkg/safepath|mynaric|comet|kvaser|labview|srv25|app03|loom|gossamer|meerstetter|/home/|192\.168|10\.[0-9]|172\.(1[6-9]|2[0-9]|3[0-1])|password|secret|token|tunnel' README.md go.mod docs/audits docs/legacy_harvest_register.md safepath graphwall jsonfile mathutil graphsem || true)
if [ -n "$scan_hits" ]; then
	printf '%s\n%s\n' "unexpected identity terms found:" "$scan_hits" >&2
	exit 1
fi
