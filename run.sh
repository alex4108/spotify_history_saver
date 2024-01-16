#!/usr/bin/env bash

set -exuo pipefail
chmod +x /app/spotify-history-saver

while true; do
    /app/spotify-history-saver
    sleep 60
done