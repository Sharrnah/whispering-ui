#!/bin/sh
set -eu
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends \
    python3 build-essential gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64 \
    libgl1-mesa-dev xorg-dev libasound2-dev libpulse-dev \
    libxkbcommon-dev libwayland-dev xvfb xauth git ca-certificates
rm -rf /var/lib/apt/lists/*
