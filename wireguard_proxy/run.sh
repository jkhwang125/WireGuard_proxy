#!/bin/sh
set -e

echo "=== 1. Cleaning up any old containers ==="
docker-compose down --volumes --remove-orphans 2>/dev/null || true

echo "=== 2. Starting environment via docker-compose ==="
docker-compose up --build -d

echo "=== SUCCESS ==="

