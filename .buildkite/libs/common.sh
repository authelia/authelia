#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

# .buildkite/libs/common.sh
#
# Shared helpers for the pipeline generator and the pre-command hook.
# Source with a BASH_SOURCE-relative path so it works regardless of CWD:
#   source "$(dirname "${BASH_SOURCE[0]}")/libs/common.sh"        # from .buildkite/
#   source "$(dirname "${BASH_SOURCE[0]}")/../libs/common.sh"     # from .buildkite/hooks/
#
# BASE_REF, BASE_REF_OK, and TAG_SUFFIX below are intentionally left as
# globals rather than returned via stdout; each caller reads them
# straight after calling the corresponding resolve_*() function. Since
# every caller lives in a different sourcing file, shellcheck can't see
# that usage when linting this file in isolation and reports SC2034 on
# each of them; disabled file-wide since it's a false positive here.
# shellcheck disable=SC2034

changed() {
  local base_ref="${1}"
  local path_pattern="${2}"
  git diff --name-only "${base_ref}" | grep -q "^${path_pattern}"
}

bypass_check() {
  local base_ref="${1}"
  local regex="${2}"
  git diff --name-only "${base_ref}" | sed -rn "${regex}" && echo true || echo false
}

# Resolves the ref to diff against. Sets BASE_REF and BASE_REF_OK as globals
# (rather than returning via stdout) so callers get both without a second
# git invocation. BASE_REF_OK="false" means fork-point resolution failed
# (e.g. shallow clone with no reflog) callers should treat that as
# "unknown" rather than "unchanged".
resolve_base_ref() {
  BASE_REF=""
  BASE_REF_OK="false"

  if [[ "${BUILDKITE_BRANCH}" == "master" ]]; then
    BASE_REF="HEAD~1"
    BASE_REF_OK="true"
    return
  fi

  if BASE_REF=$(git merge-base origin/master HEAD 2>/dev/null); then
    BASE_REF_OK="true"
  fi
}

# Computes the branch/PR-derived tag suffix shared by integration.sh,
# INTEGRATION() in hooks/pre-command, and steps/grypescans.sh.
#
# Sets TAG_SUFFIX as a global. Empty means "no case matched"; the
# caller decides what that means (pre-command treats it as "leave the
# file's existing tag alone"; integration.sh/grypescans.sh fall back to
# their own defaults).
#
# $1 - value to use for TAG_SUFFIX on a master, non-PR build. Defaults
#      to "latest", since an untagged image reference and an explicit
#      ":latest" tag resolve to the same image callers only need to
#      override this when they explicitly want the literal "master"
#      tag instead (grypescans.sh does).
# $2 - "true" (default) to collapse renovate-* branches to the literal
#      suffix "renovate", matching the image tag integration.sh actually
#      builds under. Pass "false" to fall through to the literal branch
#      name instead; grypescans.sh scans the main "authelia/authelia"
#      image, which is tagged with the full branch name even on
#      renovate branches, so it must not collapse here.
#
# Does NOT check BUILDKITE_TAG integration.sh and INTEGRATION() never
# run on tag builds (see pipeline.sh's BUILD_* gating), and
# grypescans.sh checks BUILDKITE_TAG itself before falling back here.
resolve_tag_suffix() {
  local master_value="${1:-latest}"
  local collapse_renovate="${2:-true}"
  TAG_SUFFIX=""

  if [[ "${collapse_renovate}" == "true" ]] && [[ "${BUILDKITE_BRANCH}" =~ ^renovate- ]]; then
    TAG_SUFFIX="renovate"
  elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ ! "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
    TAG_SUFFIX="${BUILDKITE_BRANCH}"
  elif [[ "${BUILDKITE_BRANCH}" != "master" ]] && [[ "${BUILDKITE_BRANCH}" =~ .*:.* ]]; then
    TAG_SUFFIX="PR${BUILDKITE_PULL_REQUEST}"
  elif [[ "${BUILDKITE_BRANCH}" == "master" ]] && [[ "${BUILDKITE_PULL_REQUEST}" == "false" ]]; then
    TAG_SUFFIX="${master_value}"
  fi
}
