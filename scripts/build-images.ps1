[CmdletBinding()]
param (
  [Parameter(Mandatory = $true)]
  [string]$Registry,

  [string]$Tag = "latest",

  [switch]$Push
)

$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot

$services = @(
  @{ Name = "bootstrap"; Entry = "./cmd/bootstrap" },
  @{ Name = "public-api"; Entry = "./cmd/public-api" },
  @{ Name = "auth-api"; Entry = "./cmd/auth-api" },
  @{ Name = "system-api"; Entry = "./cmd/system-api" }
)

$normalizedRegistry = $Registry.Trim().TrimEnd("/")
if ([string]::IsNullOrWhiteSpace($normalizedRegistry)) {
  throw "Registry must not be empty."
}

Push-Location $projectRoot
try {
  foreach ($service in $services) {
    $image = "$normalizedRegistry/template-go-backend-$($service.Name):$Tag"
    Write-Host "Building $image" -ForegroundColor Cyan
    docker build --build-arg APP_ENTRY=$($service.Entry) -t $image .

    if ($Push) {
      Write-Host "Pushing $image" -ForegroundColor Green
      docker push $image
    }
  }
}
finally {
  Pop-Location
}
