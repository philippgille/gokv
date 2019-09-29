#!/bin/bash

# Updates all dependencies.
# go-cloud is doing this in a similar way:
# https://github.com/google/go-cloud/blob/be25177dcd0a4e5202ef7f32deef6fdb5261da00/internal/testing/update_deps.sh

set -euxo pipefail

WORKING_DIR="$(pwd)"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export GO111MODULE=on

# Helper packages
# TODO: Currently without modules

# Implementations
cat "$SCRIPT_DIR"/implementations | while read -r MODULE_NAME; do
    echo "updating $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go get -u && go mod tidy) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Examples
(cd "$SCRIPT_DIR"/../examples && go get -u && go mod tidy) || (cd "$WORKING_DIR" && echo "update failed" && exit 1)

cd "$WORKING_DIR"
