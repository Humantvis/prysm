$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$runtime = Join-Path (Split-Path $repo -Parent) '.tools/devnet-dashboard/current.json'
if (-not (Test-Path -LiteralPath $runtime)) { return }
$record = Get-Content -LiteralPath $runtime -Raw | ConvertFrom-Json
$process = Get-Process -Id $record.pid -ErrorAction SilentlyContinue
if ($process -and $process.StartTime.ToUniversalTime().Ticks -eq ([datetime]::Parse($record.started_at)).ToUniversalTime().Ticks) {
    Stop-Process -Id $process.Id
    Write-Output 'Dashboard stopped'
}
