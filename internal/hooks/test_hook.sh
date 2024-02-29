#!/usr/bin/env sh

# This is a test hook script. It is used to test the hook functionality.

if [ "$WARC_COMMAND" = "test general error" ]; then
    echo "exit status error"
    exit 1
fi

if [ "$WARC_COMMAND" = "test skip file" ]; then
    echo "skip file"
    exit 10
fi

env | grep "WARC_"
