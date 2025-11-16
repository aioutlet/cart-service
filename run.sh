#!/bin/bash
# Cart Service - Bash Run Script with Dapr
# Port: 1008, Dapr HTTP: 3508, Dapr gRPC: 50008

echo ""
echo "============================================"
echo "Starting cart-service with Dapr..."
echo "============================================"
echo ""

# Kill any existing processes on ports
echo "Cleaning up existing processes..."

# Kill processes on port 1008 (app port)
lsof -ti:1008 | xargs kill -9 2>/dev/null || true

# Kill processes on port 3508 (Dapr HTTP port)
lsof -ti:3508 | xargs kill -9 2>/dev/null || true

# Kill processes on port 50008 (Dapr gRPC port)
lsof -ti:50008 | xargs kill -9 2>/dev/null || true

sleep 2

echo ""
echo "Building cart-service..."
go build -o cart-service ./cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"
echo ""
echo "Starting with Dapr sidecar..."
echo "App ID: cart-service"
echo "App Port: 1008"
echo "Dapr HTTP Port: 3508"
echo "Dapr gRPC Port: 50008"
echo ""

dapr run \
  --app-id cart-service \
  --app-port 1008 \
  --dapr-http-port 3508 \
  --dapr-grpc-port 50008 \
  --log-level info \
  --components-path ./.dapr/components \
  --config ./.dapr/config.yaml \
  -- ./cart-service

echo ""
echo "============================================"
echo "Service stopped."
echo "============================================"
