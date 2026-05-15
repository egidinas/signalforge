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

mod_hits=$(/bin/grep -n -i -E 'v[0-9.]+-private|loom-gossamer-shared|/home/' go.mod || true)
if [ -n "$mod_hits" ]; then
	printf '%s\n%s\n' "unexpected module identity terms found:" "$mod_hits" >&2
	exit 1
fi

scan_paths="README.md docs/audits docs/legacy_harvest_register.md safepath graphwall jsonfile mathutil graphsem arrowtelemetry contracts controlprogram tilebundle cantrace controlobserve dbcmeta pollqueue ringbuf stability stats web/src web/demo web/README.md web/package.json web/jest.config.cjs web/vite.config.ts web/vite.demo.config.ts web/tsconfig.json web/tsconfig.lib.json web/tsconfig.node.json"
scan_hits=$(/bin/grep -RIn -i -E 'github\.com/signalforge|signalforge/signalforge|pkg/safepath|mynaric|comet|kvaser|labview|srv25|app03|loom|gossamer|meerstetter|/home/|192\.168|(^|[^0-9.])10\.[0-9]{1,3}\.|(^|[^0-9.])172\.(1[6-9]|2[0-9]|3[0-1])\.|password|secret|auth[_-]?token|api[_-]?token|access[_-]?token|bearer|tunnel' $scan_paths || true)
if [ -n "$scan_hits" ]; then
	printf '%s\n%s\n' "unexpected identity terms found:" "$scan_hits" >&2
	exit 1
fi
