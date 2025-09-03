#!/usr/bin/env bash
set -e

# Wait for Hatchet to be ready
echo "Waiting for Hatchet to start..."
until curl -sSf http://hatchet-lite:8888/health >/dev/null 2>&1; do
    sleep 2
done

# Request a token dynamically
echo "Generating Hatchet client token..."
HATCHET_CLIENT_TOKEN=$(hatchet token create --tenant-id default --config /config | tr -d '\r\n')

export HATCHET_CLIENT_TOKEN

# Run the main application
exec "$@"
