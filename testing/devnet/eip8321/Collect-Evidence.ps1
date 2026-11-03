param([string]$KurtosisPath, [string]$Enclave = 'bct-eip8321', [string]$OutputPath, [int]$FromSlot = 32)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$workspace = Split-Path $repo -Parent
if (-not $KurtosisPath) { $KurtosisPath = Join-Path $workspace '.tools/kurtosis/kurtosis.exe' }
if (-not $OutputPath) { $OutputPath = Join-Path $workspace ('.tools/devnet-evidence/' + (Get-Date -Format 'yyyyMMdd-HHmmss')) }
New-Item -ItemType Directory -Force -Path $OutputPath | Out-Null
$urls = foreach ($index in 1,2) {
    $url = (& $KurtosisPath port print $Enclave "cl-$index-prysm-geth" http).Trim()
    if ($LASTEXITCODE -ne 0) { throw "Could not locate node $index" }
    $url
}
$heads = foreach ($url in $urls) { Invoke-RestMethod "$url/eth/v1/beacon/headers/head" }
$slot = [math]::Min([long]$heads[0].data.header.message.slot, [long]$heads[1].data.header.message.slot)
$canonical = foreach ($url in $urls) { Invoke-RestMethod "$url/eth/v1/beacon/headers/$slot" }
if ($canonical[0].data.root -ne $canonical[1].data.root) { throw 'Nodes disagree on the comparison block' }
$finality = Invoke-RestMethod "$($urls[0])/eth/v1/beacon/states/head/finality_checkpoints"
$fork = Invoke-RestMethod "$($urls[0])/eth/v1/beacon/states/head/fork"
$finality | ConvertTo-Json -Depth 20 | Set-Content (Join-Path $OutputPath 'finality.json')
$fork | ConvertTo-Json -Depth 20 | Set-Content (Join-Path $OutputPath 'fork.json')
$rows = @()
$missing = @()
for ($n = $FromSlot; $n -le $slot; $n++) {
    try { $response = Invoke-RestMethod "$($urls[0])/eth/v2/beacon/blocks/$n" }
    catch {
        if ([int]$_.Exception.Response.StatusCode -eq 404) { $missing += $n; continue }
        throw
    }
    $response | ConvertTo-Json -Depth 100 | Set-Content (Join-Path $OutputPath "block-$n.json")
    $b = $response.data.message
    $reveal = $b.body.hash_chain_reveal
    $nonzero = $reveal -match '^0x[0-9a-fA-F]{64}$' -and $reveal -notmatch '^0x0{64}$'
    if ($nonzero -and $b.body.randao_reveal -notmatch '^0xc00{190}$') { throw "Slot $n uses a hash reveal without the required infinity BLS signature" }
    $rows += [pscustomobject]@{
        slot = $n; version = $response.version; proposer_index = $b.proposer_index
        registrations = @($b.body.randao_commitment_registrations).Count
        hash_chain_reveal = $reveal; nonzero_reveal = $nonzero
    }
}
$summary = [ordered]@{
    checked_at = (Get-Date).ToString('o'); enclave = $Enclave; compared_slot = $slot
    common_root = $canonical[0].data.root; fork_version = $fork.data.current_version
    fork_epoch = $fork.data.epoch; finalized_epoch = $finality.data.finalized.epoch
    execution_optimistic = $canonical[0].execution_optimistic
    registration_count = ($rows | Measure-Object registrations -Sum).Sum
    hash_reveal_block_count = @($rows | Where-Object nonzero_reveal).Count
    missed_slots = $missing; blocks = $rows
}
$summary | ConvertTo-Json -Depth 20 | Set-Content (Join-Path $OutputPath 'summary.json')
$rows | Format-Table slot,version,proposer_index,registrations,nonzero_reveal
Write-Output "Evidence: $OutputPath"
if ($summary.registration_count -lt 1 -or $summary.hash_reveal_block_count -lt 1 -or [long]$summary.finalized_epoch -le [long]$summary.fork_epoch -or $summary.execution_optimistic) {
    throw 'Evidence is incomplete: need registrations, nonzero reveals, execution validation and finality after the fork'
}
