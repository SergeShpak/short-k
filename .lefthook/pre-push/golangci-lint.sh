#! /bin/bash
set -e

main() {
    local REPORT_FILE="./.golangci-lint/report"
    make golangci-lint-run > /dev/null 2> /dev/null

    if [ -s "$REPORT_FILE" ]; then
        # if file is not empty
        echo "golangci-lint has detected several problems, check $REPORT_FILE"
        exit 1
    fi
}

main
