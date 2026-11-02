param([string]$KurtosisPath, [string]$Enclave = 'bct-eip8321', [string]$EvidencePath)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
if (-not $KurtosisPath) { $KurtosisPath = Join-Path (Split-Path $repo -Parent) '.tools/kurtosis/kurtosis.exe' }
$nodes = foreach ($index in 1,2) {
    $url = (& $KurtosisPath port print $Enclave "cl-$index-prysm-geth" http).Trim()
    if ($LASTEXITCODE -ne 0) { throw "Could not find node $index" }
    $head = Invoke-RestMethod "$url/eth/v1/beacon/headers/head"
    $fork = Invoke-RestMethod "$url/eth/v1/beacon/states/head/fork"
    $finality = Invoke-RestMethod "$url/eth/v1/beacon/states/head/finality_checkpoints"
    [pscustomobject]@{
        node = $index; url = $url; slot = $head.data.header.message.slot
        root = $head.data.root; fork_version = $fork.data.current_version
        finalized_epoch = $finality.data.finalized.epoch
        execution_optimistic = $head.execution_optimistic
    }
}
$dora = (& $KurtosisPath port print $Enclave dora http).Trim()
$report = [ordered]@{ checked_at = (Get-Date).ToString('o'); enclave = $Enclave; nodes = @($nodes); dora = $dora }
$nodes | Format-Table
Write-Output "Dora: $dora"
if ($EvidencePath) { $report | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $EvidencePath }
