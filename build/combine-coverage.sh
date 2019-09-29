#!/bin/bash

set -euxo pipefail

WORKING_DIR="$(pwd)"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

rm -f "$SCRIPT_DIR"/../coverage.txt

cat "$SCRIPT_DIR"/implementations | while read -r MODULE_NAME; do
    if [[ -f "$SCRIPT_DIR"/../"$MODULE_NAME"/coverage.txt ]]; then
        # Using grep to skip the first line of each coverage report (there's probably a more elegant way to do this)
        cat "$SCRIPT_DIR"/../"$MODULE_NAME"/coverage.txt | grep gokv >> "$SCRIPT_DIR"/../coverage.txt
    fi
done
