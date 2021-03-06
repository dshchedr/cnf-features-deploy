#!/bin/bash
set -e
# Setting -e is fine as we want both config and validiation to succeed
# before running the "real" tests.

suites=(validationsuite configsuite cnftests)
SUITES_PATH="${SUITES_PATH:-~/usr/bin}"


for suite in "${suites[@]}"; do
    echo running "$SUITES_PATH/$suite" "$@"
    "$SUITES_PATH/$suite" "$@"
done
