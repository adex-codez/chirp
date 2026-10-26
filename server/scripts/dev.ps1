#!/usr/bin/env pwsh
param(
    [int]$Port = 8080
)

$ErrorActionPreference = "Stop"
$mobileEnv = Join-Path (Join-Path (Join-Path $PSScriptRoot "..") "..") "chirp-mobile\.env"

if (-not (Get-Command ngrok -ErrorAction SilentlyContinue)) {
    Write-Error @"
ngrok CLI not found. Install it once, then re-run:
  winget install ngrok.ngrok
  ngrok config add-authtoken <your-token>   # free at https://dashboard.ngrok.com
"@
    exit 1
}

$ngrok = Start-Process ngrok -ArgumentList "http", "$Port" -NoNewWindow -PassThru

function Stop-Tunnel {
    if ($ngrok -and -not $ngrok.HasExited) {
        Stop-Process -Id $ngrok.Id -Force
    }
}

try {
    # Wait for the tunnel (ngrok exposes its agent API on :4040).
    $publicUrl = $null
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Seconds 1
        if ($ngrok.HasExited) {
            throw "ngrok exited unexpectedly. Run 'ngrok http $Port' to see why (usually a missing authtoken)."
        }
        try {
            $tunnels = Invoke-RestMethod "http://127.0.0.1:4040/api/tunnels" -TimeoutSec 2
            $publicUrl = ($tunnels.tunnels | Where-Object { $_.proto -eq "https" } | Select-Object -First 1).public_url
            if ($publicUrl) { break }
        } catch {
            # Agent not up yet; keep polling.
        }
    }
    if (-not $publicUrl) { throw "Timed out waiting for the ngrok tunnel." }

    Write-Host ""
    Write-Host "Tunnel live: $publicUrl -> http://localhost:$Port" -ForegroundColor Green
    Write-Host ""

    # Point the mobile app at the tunnel. EXPO_PUBLIC_* values bake in at
    # bundle time, so restart `expo start` after this changes. Other lines
    # in .env (e.g. Google client IDs) are left untouched.
    $key = "EXPO_PUBLIC_API_URL"
    $lines = @()
    if (Test-Path -LiteralPath $mobileEnv) {
        $lines = @(Get-Content -LiteralPath $mobileEnv)
    }
    if ($lines -contains "$key=$publicUrl") {
        Write-Host "$key already points at the tunnel; .env untouched."
    } else {
        $updated = $false
        $lines = @($lines | ForEach-Object {
                if ($_ -match "^\s*$key\s*=") { $updated = $true; "$key=$publicUrl" } else { $_ }
            })
        if (-not $updated) { $lines += "$key=$publicUrl" }
        Set-Content -LiteralPath $mobileEnv -Value $lines
        Write-Host "Wrote $key=$publicUrl to chirp-mobile/.env (restart expo start to pick it up)."
    }
    Write-Host ""

    go run github.com/air-verse/air@v1.61.7
} finally {
    Stop-Tunnel
}
