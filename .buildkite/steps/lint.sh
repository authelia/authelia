#!/usr/bin/env bash

# Usage:
#   lint.sh                  Run every linter (CI linting step entrypoint).
#   lint.sh shellcheck ...   Run shellcheck. With file args, those files are
#                            linted verbatim (used by lefthook's staged path).
#                            Otherwise, every tracked shell script is
#                            discovered via git ls-files + shebang scan and
#                            passed in. Any flag-style arguments (--format=...)
#                            are forwarded to shellcheck.
#   lint.sh -flag ...        Anything else is forwarded to reviewdog. A reporter
#                            which cannot post because the pull request diff
#                            exceeds the GitHub API limit falls back to the
#                            local reporter, see run_reviewdog.

set -uo pipefail

# The GitHub API refuses to return a diff over 20000 lines, which makes any reporter which posts against a pull request
# fail on a large pull request even when every linter passed. This matches that refusal so it can be retried locally.
readonly REVIEWDOG_DIFF_TOO_LARGE_RE='diff exceeded the maximum number of lines'

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

run_reviewdog() {
  local output status skip arg
  local -a args

  output=$(mktemp)

  reviewdog "$@" 2>&1 | tee "${output}"
  status=${PIPESTATUS[0]}

  if (( status == 0 )); then
    rm -f "${output}"

    return 0
  fi

  if ! grep -qF "${REVIEWDOG_DIFF_TOO_LARGE_RE}" "${output}"; then
    rm -f "${output}"

    return "${status}"
  fi

  rm -f "${output}"

  echo "--- :warning: Reporter could not post as the pull request diff exceeds the GitHub API limit, retrying with the local reporter"

  args=()
  skip=0

  # The reporter is dropped in both its joined and separated forms so it can be replaced below.
  for arg in "$@"; do
    if (( skip )); then
      skip=0

      continue
    fi

    case "${arg}" in
      -reporter=*|--reporter=*) continue ;;
      -reporter|--reporter) skip=1; continue ;;
    esac

    args+=("${arg}")
  done

  reviewdog -reporter=local "${args[@]}"
}

cd "$(git rev-parse --show-toplevel)" || exit 1

if [[ $# -eq 0 ]]; then
  FAILED=0

  echo "--- :go::service_dog: Running goimports-reviser"
  goimports-reviser -rm-unused -format -excludes '*/node_modules,*/*/*/node_modules' -company-prefixes authelia.com,github.com/authelia ./... || FAILED=1
  echo "--- :go::service_dog: Running golangci-lint"
  golangci-lint run || FAILED=1
  echo "--- :yaml::service_dog: Running yamllint"
  yamllint . || FAILED=1
  echo "--- :shellcheck::service_dog: Running shellcheck"
  run_shellcheck || FAILED=1
  echo "--- :eslint::service_dog: Running eslint"
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
  run_reviewdog "$@"
fi
