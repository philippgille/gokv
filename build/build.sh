#!/bin/bash

set -euxo pipefail

WORKING_DIR="$(pwd)"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Helper packages
export GO111MODULE=off
array=( encoding sql test util )
for PACKAGE_NAME in "${array[@]}"; do
    echo "building $PACKAGE_NAME"
    (cd "$SCRIPT_DIR"/../"$PACKAGE_NAME" && go get && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Implementations
export GO111MODULE=on
cat "$SCRIPT_DIR"/allmodules | while read -r MODULE_NAME; do
    echo "building $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Examples
echo "building examples"
(cd "$SCRIPT_DIR"/../examples && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
