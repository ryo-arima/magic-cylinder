#!/bin/bash

# Test script for Magic Cylinder WebTransport experiment

echo "=== Magic Cylinder WebTransport Test ==="
echo

# Check if certificates exist
if [ ! -f "certs/server.crt" ] || [ ! -f "certs/server.key" ]; then
    echo "Generating certificates..."
    make certs
    echo
fi

# Build the applications
echo "Building applications..."
make build
echo

# Check if binaries were created
if [ ! -f "bin/repository" ] || [ ! -f "bin/controller" ]; then
    echo "❌ Build failed - binaries not found"
    exit 1
fi

echo "✅ Build successful"
echo

# Start repository server in background
echo "Starting Repository server (port 8443)..."
./bin/repository &
REPO_PID=$!

# Wait for repository server to start
sleep 3

# Start controller server in background
echo "Starting Controller server (port 8444)..."
./bin/controller &
CONTROLLER_PID=$!

echo
echo "🚀 Both servers started!"
echo "Repository PID: $REPO_PID"
echo "Controller PID: $CONTROLLER_PID"
echo
echo "Press Ctrl+C to stop both servers"
echo

# Wait for user interrupt
wait

# Cleanup
echo
echo "Stopping servers..."
kill $REPO_PID 2>/dev/null
kill $CONTROLLER_PID 2>/dev/null
echo "✅ Servers stopped"