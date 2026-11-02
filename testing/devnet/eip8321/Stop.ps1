param([string]$KurtosisPath, [string]$Enclave = 'bct-eip8321')
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
if (-not $KurtosisPath) { $KurtosisPath = Join-Path (Split-Path $repo -Parent) '.tools/kurtosis/kurtosis.exe' }
& $KurtosisPath enclave stop $Enclave
if ($LASTEXITCODE -ne 0) { throw 'Could not stop the specified enclave' }
