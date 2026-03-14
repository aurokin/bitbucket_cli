#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/.tools/bin"

mapfile -t GO_FILES < <(cd "${ROOT_DIR}" && find cmd integration internal -name '*.go' ! -name '*_test.go' | sort)

if [[ "${#GO_FILES[@]}" -eq 0 ]]; then
  echo "no Go files found for complexity checks"
  exit 0
fi

"${BIN_DIR}/gocognit" -over 30 "${GO_FILES[@]}"
"${BIN_DIR}/gocyclo" -over 30 "${GO_FILES[@]}"
