[CmdletBinding(DefaultParameterSetName = "set")]
param (
  [Parameter(Mandatory = $true, ParameterSetName = "set")]
  [Parameter(Mandatory = $true, ParameterSetName = "promote")]
  [ValidateSet("staging", "production")]
  [string]$Overlay,

  [Parameter(Mandatory = $true, ParameterSetName = "set")]
  [ValidatePattern("^sha-[0-9a-fA-F]{7,40}$|^latest$")]
  [string]$Tag,

  [Parameter(Mandatory = $true, ParameterSetName = "promote")]
  [ValidateSet("staging", "production")]
  [string]$SourceOverlay
)

$ErrorActionPreference = "Stop"

if ($PSCmdlet.ParameterSetName -eq "promote" -and $SourceOverlay -eq $Overlay) {
  throw "SourceOverlay and Overlay must be different."
}

$projectRoot = Split-Path -Parent $PSScriptRoot
$overlayRoot = Join-Path $projectRoot "deploy\\k8s\\overlays"

$serviceFiles = @{
  bootstrap  = @("kustomization.yaml", "bootstrap\\kustomization.yaml")
  "public-api" = @("kustomization.yaml", "runtime\\kustomization.yaml")
  "auth-api" = @("kustomization.yaml", "runtime\\kustomization.yaml")
  "system-api" = @("kustomization.yaml", "runtime\\kustomization.yaml")
}

function Get-OverlayPath([string]$name) {
  return Join-Path $overlayRoot $name
}

function Assert-OverlayExists([string]$name) {
  $path = Get-OverlayPath $name
  if (-not (Test-Path $path)) {
    throw "Overlay '$name' does not exist at $path"
  }
}

function Read-ServiceTags([string]$name) {
  $path = Join-Path (Get-OverlayPath $name) "kustomization.yaml"
  $lines = Get-Content $path
  $tags = @{}

  foreach ($service in $serviceFiles.Keys) {
    $tag = Get-ServiceTagFromLines -lines $lines -service $service
    if ([string]::IsNullOrWhiteSpace($tag)) {
      throw "Could not find image tag for service '$service' in $path"
    }
    $tags[$service] = $tag
  }

  return $tags
}

function Get-ServiceTagFromLines([string[]]$lines, [string]$service) {
  $image = "ghcr.io/gee-coder/template-go-backend-$service"

  for ($index = 0; $index -lt $lines.Length; $index++) {
    if ($lines[$index].Trim() -eq "- name: $image") {
      for ($scan = $index + 1; $scan -lt $lines.Length; $scan++) {
        $trimmed = $lines[$scan].Trim()
        if ($trimmed -like "- name:*") {
          break
        }
        if ($trimmed -like "newTag:*") {
          return ($trimmed -replace '^newTag:\s*', '').Trim()
        }
      }
    }
  }

  return ""
}

function Update-ServiceTag([string]$filePath, [string]$service, [string]$newTag) {
  $lines = @(Get-Content $filePath)
  $image = "ghcr.io/gee-coder/template-go-backend-$service"
  $updated = $false

  for ($index = 0; $index -lt $lines.Length; $index++) {
    if ($lines[$index].Trim() -eq "- name: $image") {
      for ($scan = $index + 1; $scan -lt $lines.Length; $scan++) {
        $trimmed = $lines[$scan].Trim()
        if ($trimmed -like "- name:*") {
          break
        }
        if ($trimmed -like "newTag:*") {
          $indent = [Regex]::Match($lines[$scan], '^\s*').Value
          $lines[$scan] = "${indent}newTag: $newTag"
          $updated = $true
          break
        }
      }
    }
  }

  if (-not $updated) {
    throw "Could not update service '$service' in $filePath"
  }

  $normalized = (($lines -join [Environment]::NewLine).TrimEnd("`r", "`n")) + [Environment]::NewLine
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($filePath, $normalized, $utf8NoBom)
}

function Apply-ServiceTags([string]$name, [hashtable]$tags) {
  $overlayPath = Get-OverlayPath $name

  foreach ($service in $tags.Keys) {
    foreach ($relative in $serviceFiles[$service]) {
      $filePath = Join-Path $overlayPath $relative
      if (Test-Path $filePath) {
        Update-ServiceTag -filePath $filePath -service $service -newTag $tags[$service]
      }
    }
  }
}

Assert-OverlayExists $Overlay

if ($PSCmdlet.ParameterSetName -eq "promote") {
  Assert-OverlayExists $SourceOverlay
  $resolvedTags = Read-ServiceTags -name $SourceOverlay
}
else {
  $resolvedTags = @{
    bootstrap = $Tag
    "public-api" = $Tag
    "auth-api" = $Tag
    "system-api" = $Tag
  }
}

Apply-ServiceTags -name $Overlay -tags $resolvedTags

Write-Host "Updated overlay '$Overlay' with image tags:" -ForegroundColor Green
foreach ($service in ($resolvedTags.Keys | Sort-Object)) {
  Write-Host "  $service -> $($resolvedTags[$service])"
}
