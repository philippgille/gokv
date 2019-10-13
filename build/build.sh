#!/bin/bash

set -euxo pipefail

WORKING_DIR="$(pwd)"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Interface
echo "building gokv"
(cd "$SCRIPT_DIR"/.. && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)

# Helper packages
array=( encoding sql test util )
for MODULE_NAME in "${array[@]}"; do
    echo "building $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Implementations
cat "$SCRIPT_DIR"/implementations | while read -r MODULE_NAME; do
    echo "building $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Examples
echo "building examples"
(cd "$SCRIPT_DIR"/../examples && go build -v) || (cd "$WORKING_DIR" && echo " failed" && exit 1)

cd "$WORKING_DIR"
