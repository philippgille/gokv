#!/bin/bash

# Updates all dependencies.
# go-cloud is doing this in a similar way:
# https://github.com/google/go-cloud/blob/be25177dcd0a4e5202ef7f32deef6fdb5261da00/internal/testing/update_deps.sh
# But we don't want to update transitive dependencies, so instead of using `go get -u` we use
# `go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)`
# as suggested in https://github.com/golang/go/issues/28424#issuecomment-1101896499.

set -euxo pipefail

WORKING_DIR="$(pwd)"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export GO111MODULE=on

get_direct_dependencies() {
  go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all
}

# Helper packages
array=( encoding sql test util )
for MODULE_NAME in "${array[@]}"; do
    echo "updating $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go get $(get_direct_dependencies) && go mod tidy) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Implementations
cat "$SCRIPT_DIR"/implementations | while read -r MODULE_NAME; do
    echo "updating $MODULE_NAME"
    (cd "$SCRIPT_DIR"/../"$MODULE_NAME" && go get $(get_direct_dependencies) && go mod tidy) || (cd "$WORKING_DIR" && echo " failed" && exit 1)
done

# Examples
(cd "$SCRIPT_DIR"/../examples/redis && go get $(get_direct_dependencies) && go mod tidy) || (cd "$WORKING_DIR" && echo "update failed" && exit 1)
(cd "$SCRIPT_DIR"/../examples/protobuf_encoding && go get $(get_direct_dependencies) && go mod tidy) || (cd "$WORKING_DIR" && echo "update failed" && exit 1)

cd "$WORKING_DIR"
