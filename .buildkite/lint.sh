#!/usr/bin/env bash

if [ $# -eq 0 ]; then
  FAILED=0

  echo "--- :go::service_dog: Running golangci-lint"
  golangci-lint run || FAILED=1
  echo "--- :go::service_dog: Running yamllint"
  yamllint . || FAILED=1
  echo "--- :go::service_dog: Running eslint"
  cd web && eslint '*/**/*.{js,ts,tsx}' || FAILED=1 && cd ..

  echo "--- :go::service_dog: Lint Runners Completed"
  if [ $FAILED -ne 0 ]; then
    echo "Linting was not successful as one or more linters returned a non-zero exit code"
    exit 1
  else
    echo "Linting was successful"
  fi
else
  reviewdog "$@"
fi
