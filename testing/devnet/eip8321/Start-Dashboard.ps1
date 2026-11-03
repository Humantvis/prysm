param([string]$KurtosisPath, [string]$PythonPath = 'python', [string]$Enclave = 'bct-eip8321')
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$workspace = Split-Path $repo -Parent
if (-not $KurtosisPath) { $KurtosisPath = Join-Path $workspace '.tools/kurtosis/kurtosis.exe' }
$urls = foreach ($index in 1,2) {
    $url = (& $KurtosisPath port print $Enclave "cl-$index-prysm-geth" http).Trim()
    if ($LASTEXITCODE -ne 0) { throw "Could not locate node $index" }
    $url
}
$runtime = Join-Path $workspace '.tools/devnet-dashboard'
New-Item -ItemType Directory -Force -Path $runtime | Out-Null
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$out = Join-Path $runtime "$stamp.log"
$err = Join-Path $runtime "$stamp-error.log"
$server = Join-Path $PSScriptRoot 'demo_server.py'
$process = Start-Process -FilePath $PythonPath -ArgumentList @('-u',('"'+$server+'"')) -WindowStyle Hidden -PassThru -RedirectStandardOutput $out -RedirectStandardError $err
$base = $null
for ($attempt = 0; $attempt -lt 50; $attempt++) {
    if ($process.HasExited) { throw "Dashboard server exited; inspect $err" }
    if (Test-Path -LiteralPath $out) { $base = Get-Content -LiteralPath $out -TotalCount 1 }
    if ($base -match '^http://127\.0\.0\.1:\d+$') { break }
    Start-Sleep -Milliseconds 100
}
if (-not $base) { throw "Dashboard server did not start; inspect $err" }
$url = "$base/dashboard.html?node1=$([uri]::EscapeDataString($urls[0]))&node2=$([uri]::EscapeDataString($urls[1]))"
[ordered]@{ pid = $process.Id; started_at = $process.StartTime.ToString('o'); url = $url } | ConvertTo-Json | Set-Content (Join-Path $runtime 'current.json')
Write-Output $url
