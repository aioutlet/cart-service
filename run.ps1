#!/usr/bin/env pwsh
# Cart Service - PowerShell Run Script with Dapr
# Port: 1008, Dapr HTTP: 3508, Dapr gRPC: 50008

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Starting cart-service with Dapr..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Kill any existing processes on ports
Write-Host "Cleaning up existing processes..." -ForegroundColor Yellow

# Kill process on port 1008 (app port)
$process = Get-NetTCPConnection -LocalPort 1008 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($process) {
    Write-Host "Killing process on port 1008 (PID: $process)" -ForegroundColor Yellow
    Stop-Process -Id $process -Force -ErrorAction SilentlyContinue
}

# Kill process on port 3508 (Dapr HTTP port)
$process = Get-NetTCPConnection -LocalPort 3508 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($process) {
    Write-Host "Killing process on port 3508 (PID: $process)" -ForegroundColor Yellow
    Stop-Process -Id $process -Force -ErrorAction SilentlyContinue
}

# Kill process on port 50008 (Dapr gRPC port)
$process = Get-NetTCPConnection -LocalPort 50008 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($process) {
    Write-Host "Killing process on port 50008 (PID: $process)" -ForegroundColor Yellow
    Stop-Process -Id $process -Force -ErrorAction SilentlyContinue
}

Start-Sleep -Seconds 2

Write-Host ""
Write-Host "Starting with Dapr sidecar..." -ForegroundColor Green
Write-Host "App ID: cart-service" -ForegroundColor Cyan
Write-Host "App Port: 1008" -ForegroundColor Cyan
Write-Host "Dapr HTTP Port: 3508" -ForegroundColor Cyan
Write-Host "Dapr gRPC Port: 50008" -ForegroundColor Cyan
Write-Host ""

dapr run `
  --app-id cart-service `
  --app-port 1008 `
  --dapr-http-port 3508 `
  --dapr-grpc-port 50008 `
  --log-level error `
  --resources-path ./.dapr `
  --config ./.dapr/config.yaml `
  -- go run ./cmd/server/main.go

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Service stopped." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
