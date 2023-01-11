#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright © 2021 Roberto Hidalgo <joao@un.rob.mx>

cd "$MILPA_REPO_ROOT" || @milpa.fail "could not cd into $MILPA_REPO_ROOT"
@milpa.log info "Running unit tests"
args=()
after_run=complete
if [[ "${MILPA_OPT_COVERAGE}" ]]; then
  after_run=success
  args=( -coverprofile=coverage.out --coverpkg=./...)
fi
gotestsum --format testname -- "$MILPA_ARG_SPEC" "${args[@]}" || exit 2
@milpa.log "$after_run" "Unit tests passed"

[[ ! "${MILPA_OPT_COVERAGE}" ]] && exit
@milpa.log info "Building coverage report"
go tool cover -html=coverage.out -o coverage.html || @milpa.fail "could not build reports"
@milpa.log complete "Coverage report ready at coverage.html"
