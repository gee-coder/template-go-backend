$ErrorActionPreference = "Stop"

if (-not $env:APP_CONFIG) {
  $env:APP_CONFIG = "configs/config.local.yaml"
}

Write-Host "Running tests..."
go test ./...

Write-Host "Building application..."
go build ./cmd/api

Write-Host "Done."

