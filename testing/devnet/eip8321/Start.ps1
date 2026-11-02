param([string]$KurtosisPath, [string]$PackagePath, [string]$Enclave = 'bct-eip8321')
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
if (-not $KurtosisPath) { $KurtosisPath = Join-Path (Split-Path $repo -Parent) '.tools/kurtosis/kurtosis.exe' }
if (-not (Test-Path -LiteralPath $KurtosisPath)) { throw 'Provide -KurtosisPath pointing to the Kurtosis CLI' }
$prepared = & (Join-Path $PSScriptRoot 'Prepare-Package.ps1') -PackagePath $PackagePath
if (-not $prepared) { throw 'Package preparation did not return a path' }
$PackagePath = [string](@($prepared)[-1])
& $KurtosisPath engine start
if ($LASTEXITCODE -ne 0) { throw 'Kurtosis engine could not start' }
& $KurtosisPath run $PackagePath --enclave $Enclave --args-file (Join-Path $PSScriptRoot 'network-params.yaml') --image-download missing --verbosity brief
if ($LASTEXITCODE -ne 0) { throw 'Devnet launch failed; inspect the Kurtosis output' }
& (Join-Path $PSScriptRoot 'Status.ps1') -KurtosisPath $KurtosisPath -Enclave $Enclave
