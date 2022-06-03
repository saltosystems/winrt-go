#!/usr/bin/env bash

set -u
set -e

if [ -d /etc/go-builder ]; then
    diff -rq .go-builder /etc/go-builder
    exit $?
else
    echo "check-sync is not running in builder's docker image. Skipping..."
    exit 0
fi
