#! /bin/bash

set -euo pipefail

go mod tidy

git diff --exit-code -- go.mod go.sum &> /dev/null

if [ $? -eq 1 ]; then
    echo "go.mod or go.sum differs, please re-add it to your commit"
    exit 1
fi
