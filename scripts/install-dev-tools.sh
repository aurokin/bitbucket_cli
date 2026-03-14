#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/.tools/bin"
STATE_FILE="${ROOT_DIR}/.tools/versions.txt"

GOLANGCI_LINT_VERSION="v1.64.8"
GOCYCLO_VERSION="v0.6.0"
GOCOGNIT_VERSION="v1.2.1"

mkdir -p "${BIN_DIR}"

install_tool() {
  local binary="$1"
  local module="$2"
  local version="$3"

  echo "installing ${binary}@${version}"
  GOBIN="${BIN_DIR}" go install "${module}@${version}"
}

EXPECTED_VERSIONS="$(cat <<EOF
golangci-lint ${GOLANGCI_LINT_VERSION}
gocyclo ${GOCYCLO_VERSION}
gocognit ${GOCOGNIT_VERSION}
EOF
)"

if [[ -f "${STATE_FILE}" ]] && [[ "$(cat "${STATE_FILE}")" == "${EXPECTED_VERSIONS}" ]]; then
  missing=0
  for binary in golangci-lint gocyclo gocognit; do
    if [[ ! -x "${BIN_DIR}/${binary}" ]]; then
      missing=1
      break
    fi
  done
  if [[ "${missing}" -eq 0 ]]; then
    exit 0
  fi
fi

install_tool "golangci-lint" "github.com/golangci/golangci-lint/cmd/golangci-lint" "${GOLANGCI_LINT_VERSION}"
install_tool "gocyclo" "github.com/fzipp/gocyclo/cmd/gocyclo" "${GOCYCLO_VERSION}"
install_tool "gocognit" "github.com/uudashr/gocognit/cmd/gocognit" "${GOCOGNIT_VERSION}"

printf '%s\n' "${EXPECTED_VERSIONS}" > "${STATE_FILE}"
