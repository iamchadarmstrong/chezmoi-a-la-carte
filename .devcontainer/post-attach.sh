#!/bin/sh
# exit on error
set -e

# Debug mode - set to 1 to enable
DEBUG=${DEBUG:-0}
# Non-interactive mode - set to 1 to skip user prompts
NONINTERACTIVE=${NONINTERACTIVE:-0}

debug() {
    if [ "$DEBUG" = "1" ]; then
        echo "DEBUG: $1" >&2
    fi
}