#!/bin/bash

# Run Cart Service with Dapr
echo "Starting Cart Service with Dapr..."

# Set Dapr app ID
APP_ID="cart-service"
APP_PORT=8085
DAPR_HTTP_PORT=3500
DAPR_GRPC_PORT=50001

# Run with Dapr
dapr run \
  --app-id $APP_ID \
  --app-port $APP_PORT \
  --dapr-http-port $DAPR_HTTP_PORT \
  --dapr-grpc-port $DAPR_GRPC_PORT \
  --components-path ./.dapr/components \
  --log-level info \
  -- go run ./cmd/server/main.go
