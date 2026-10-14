#!/usr/bin/env bash
set -euo pipefail

VERSAO="${1:-}"
if [[ ! "$VERSAO" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "Uso: ./release.sh v1.0.0"
	exit 1
fi

cd "$(dirname "$0")"
rm -rf dist
mkdir -p dist

for ARCH in amd64 arm64; do
	echo "Compilando linux/$ARCH..."
	CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build \
		-trimpath \
		-ldflags="-s -w -X main.versao=$VERSAO" \
		-o "dist/instalador-linux-$ARCH" .
done

(cd dist && sha256sum instalador-* > SHA256SUMS)

echo
echo "Arquivos prontos em dist/:"
ls -lh dist
