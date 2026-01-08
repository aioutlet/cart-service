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

# Detect OS and use appropriate commands
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    # Windows (Git Bash, Cygwin, etc.)
    IS_WINDOWS=true
    
    # Kill processes using PowerShell
    powershell -Command "Get-NetTCPConnection -LocalPort 1008 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { Stop-Process -Id \$_ -Force -ErrorAction SilentlyContinue }" 2>/dev/null || true
    powershell -Command "Get-NetTCPConnection -LocalPort 3508 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { Stop-Process -Id \$_ -Force -ErrorAction SilentlyContinue }" 2>/dev/null || true
    powershell -Command "Get-NetTCPConnection -LocalPort 50008 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique | ForEach-Object { Stop-Process -Id \$_ -Force -ErrorAction SilentlyContinue }" 2>/dev/null || true
else
    # Linux/Mac
    IS_WINDOWS=false
    
    # Kill processes on port 1008 (app port)
    lsof -ti:1008 | xargs kill -9 2>/dev/null || true

    # Kill processes on port 3508 (Dapr HTTP port)
    lsof -ti:3508 | xargs kill -9 2>/dev/null || true

    # Kill processes on port 50008 (Dapr gRPC port)
    lsof -ti:50008 | xargs kill -9 2>/dev/null || true
fi

sleep 2

echo ""
echo "Starting Quarkus cart-service..."
echo "App ID: cart-service"
echo "App Port: 1008"
echo "Dapr HTTP Port: 3508"
echo "Dapr gRPC Port: 50008"
echo ""

if [ "$IS_WINDOWS" = true ]; then
    # MSYS_NO_PATHCONV prevents Git Bash from converting /c to C:/
    MSYS_NO_PATHCONV=1 dapr run \
      --app-id cart-service \
      --app-port 1008 \
      --dapr-http-port 3508 \
      --dapr-grpc-port 50008 \
      --log-level info \
      --resources-path ./.dapr/components \
      --config ./.dapr/config.yaml \
      -- cmd /c "mvnw.cmd quarkus:dev"
else
    dapr run \
      --app-id cart-service \
      --app-port 1008 \
      --dapr-http-port 3508 \
      --dapr-grpc-port 50008 \
      --log-level info \
      --resources-path ./.dapr/components \
      --config ./.dapr/config.yaml \
      -- ./mvnw quarkus:dev
fi

echo ""
echo "============================================"
echo "Service stopped."
echo "============================================"
