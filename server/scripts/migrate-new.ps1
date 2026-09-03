#!/usr/bin/env pwsh
param(
    [Parameter(Mandatory=$true)]
    [string]$Name
)

$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$filename = "db\migrations\$($timestamp)_$($Name).sql"

@"
-- +goose Up
-- +goose StatementBegin


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


-- +goose StatementEnd
"@ | Out-File -FilePath $filename -Encoding utf8

Write-Host "Created migration: $filename"
