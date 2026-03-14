$ErrorActionPreference = "Stop"

if (-not $env:APP_CONFIG) {
  $env:APP_CONFIG = "configs/config.local.yaml"
}

function Get-GoHostValue {
  param (
    [Parameter(Mandatory = $true)]
    [string]$Name
  )

  return (go env $Name).Trim()
}

$originalGOOS = $env:GOOS
$originalGOARCH = $env:GOARCH

try {
  # Match the local Go toolchain host so tests and builds stay runnable
  # even when the shell exports cross-compilation defaults like GOOS=linux.
  $env:GOOS = Get-GoHostValue -Name "GOHOSTOS"
  $env:GOARCH = Get-GoHostValue -Name "GOHOSTARCH"

  Write-Host "Detected platform: $($env:GOOS)/$($env:GOARCH)"

  Write-Host "Running tests..."
  go test ./...

  Write-Host "Building application..."
  go build ./cmd/api

  Write-Host "Done."
}
finally {
  if ([string]::IsNullOrWhiteSpace($originalGOOS)) {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
  }
  else {
    $env:GOOS = $originalGOOS
  }

  if ([string]::IsNullOrWhiteSpace($originalGOARCH)) {
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
  }
  else {
    $env:GOARCH = $originalGOARCH
  }
}
