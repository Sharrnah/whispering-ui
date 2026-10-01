#!/bin/sh
set -eu
export DEBIAN_FRONTEND=noninteractive
for attempt in 1 2 3; do
    apt-get update -o Acquire::Retries=3
    if apt-get install -y --no-install-recommends -o Acquire::Retries=3 \
        python3 build-essential gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64 \
        libgl1-mesa-dev xorg-dev libasound2-dev libpulse-dev \
        libxkbcommon-dev libwayland-dev xvfb xauth git ca-certificates; then
        break
    fi
    if [ "$attempt" -eq 3 ]; then
        exit 1
    fi
    echo "apt package fetch failed; refreshing indexes before retry $((attempt + 1))/3"
    sleep 5
done
rm -rf /var/lib/apt/lists/*
