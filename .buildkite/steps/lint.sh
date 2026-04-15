#!/usr/bin/env bash

# Usage:
#   lint.sh                  Run every linter (CI linting step entrypoint).
#   lint.sh shellcheck ...   Run shellcheck. With file args, those files are
#                            linted verbatim (used by lefthook's staged path).
#                            Otherwise, every tracked shell script is
#                            discovered via git ls-files + shebang scan and
#                            passed in. Any flag-style arguments (--format=...)
#                            are forwarded to shellcheck.
#   lint.sh -flag ...        Anything else is forwarded to reviewdog.

set -uo pipefail

discover_shell_files() {
  # A file is considered a shell script if any of:
  #   - its path matches *.sh
  #   - its path is under .buildkite/hooks/
  #   - its first line is a shell shebang (#!/bin/sh, #!/usr/bin/env bash, ...)
  local shebang_re='^#!.*(ba|da|a|k|z)?sh([[:space:]]|$)'
  local f first
  {
    git ls-files '*.sh' '.buildkite/hooks/*'
    git ls-files | while IFS= read -r f; do
      case "${f}" in
        *.sh) continue ;;
        .buildkite/hooks/*) continue ;;
      esac
      [ -f "${f}" ] || continue
      if IFS= read -r first < "${f}" 2>/dev/null && [[ "${first}" =~ ${shebang_re} ]]; then
        printf '%s\n' "${f}"
      fi
    done
  } | sort -u
}

run_shellcheck() {
  local has_files=0 arg
  for arg in "$@"; do
    case "${arg}" in
      -*) ;;
      *) has_files=1 ;;
    esac
  done
  if (( has_files )); then
    shellcheck "$@"
  else
    local files
    files=$(discover_shell_files)
    if [ -z "${files}" ]; then
      echo "no shell files found" >&2
      return 1
    fi
    # shellcheck disable=SC2086  # intentional word-splitting of the newline-separated file list
    shellcheck "$@" ${files}
  fi
}

cd "$(git rev-parse --show-toplevel)" || exit 1

if [[ $# -eq 0 ]]; then
  FAILED=0

  echo "--- :go::service_dog: Running golangci-lint"
  golangci-lint run || FAILED=1
  echo "--- :go::service_dog: Running yamllint"
  yamllint . || FAILED=1
  echo "--- :go::service_dog: Running shellcheck"
  run_shellcheck || FAILED=1
  echo "--- :go::service_dog: Running eslint"
  cd web && eslint '*/**/*.{js,ts,tsx}' || FAILED=1 && cd ..

  echo "--- :go::service_dog: Lint Runners Completed"
  if [[ ${FAILED} -ne 0 ]]; then
    echo "Linting was not successful as one or more linters returned a non-zero exit code"
    exit 1
  else
    echo "Linting was successful"
  fi
elif [[ $1 == "shellcheck" ]]; then
  shift
  run_shellcheck "$@"
else
  reviewdog "$@"
fi
