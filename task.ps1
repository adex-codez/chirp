#!/usr/bin/env pwsh
param(
    [Parameter(Position=0)]
    [string]$Task,

    [Parameter(Position=1)]
    [string]$Name
)

$serverDir = Join-Path $PSScriptRoot "server"
$scriptsDir = Join-Path $serverDir "scripts"

function Invoke-ServerTask {
    param(
        [string]$Script,
        [hashtable]$Parameters = @{}
    )

    Push-Location $serverDir
    try {
        & (Join-Path $scriptsDir $Script) @Parameters
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    finally {
        Pop-Location
    }
}

function Show-Help {
    Write-Host @"
Usage: .\task.ps1 <command> [options]

Commands:
  build                   Build the server binary
  run                     Run the server
  dev                     Run with hot reload (air) + ngrok tunnel for the API
  migrate-new -Name <n>   Create a new migration
  migrate-up              Run pending migrations
  migrate-down            Rollback the last migration
  migrate-status          Check migration status
  generate                Generate sqlc code
  tidy                    Tidy go modules

Examples:
  .\task.ps1 build
  .\task.ps1 migrate-new -Name create_users
  .\task.ps1 dev
"@
}

switch ($Task) {
    "build"          { Invoke-ServerTask "build.ps1" }
    "run"            { Invoke-ServerTask "run.ps1" }
    "dev"            { Invoke-ServerTask "dev.ps1" }
    "migrate-new"    { Invoke-ServerTask "migrate-new.ps1" @{ Name = $Name } }
    "migrate-up"     { Invoke-ServerTask "migrate-up.ps1" }
    "migrate-down"   { Invoke-ServerTask "migrate-down.ps1" }
    "migrate-status" { Invoke-ServerTask "migrate-status.ps1" }
    "generate"       { Invoke-ServerTask "generate.ps1" }
    "tidy"           { Invoke-ServerTask "tidy.ps1" }
    default           { Show-Help }
}
