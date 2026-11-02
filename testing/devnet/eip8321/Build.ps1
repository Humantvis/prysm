param([string]$DockerPath = 'docker', [string]$OutputPath, [string]$CacheVolume = 'bct-prysm-linux-cache')
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
if (-not $OutputPath) { $OutputPath = Join-Path (Split-Path $repo -Parent) '.tools/devnet-build' }
New-Item -ItemType Directory -Force -Path (Join-Path $OutputPath 'bin') | Out-Null
$OutputPath = (Resolve-Path -LiteralPath $OutputPath).Path
Get-ChildItem -LiteralPath (Join-Path $OutputPath 'bin') -File | ForEach-Object { $_.IsReadOnly = $false }
& $DockerPath volume create $CacheVolume | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Could not create build cache volume' }
$image = 'golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd'
& $DockerPath run --rm --mount "type=bind,source=$repo,target=/workspace,readonly" --mount "type=bind,source=$OutputPath,target=/out" --mount "type=volume,source=$CacheVolume,target=/cache" --workdir /workspace $image bash -c 'sed -e "s/\r$//" testing/devnet/eip8321/build-linux.sh | bash'
if ($LASTEXITCODE -ne 0) { throw 'Prysm minimal build failed' }
$revision = (& git -C $repo rev-parse HEAD).Trim()
if (& git -C $repo status --porcelain) { $revision += '-dirty' }
$targets = @(
    @{ Binary = 'beacon-chain'; Base = 'ethpandaops/prysm-beacon-chain:heze-devnet-2-minimal-3cc4d4e'; Tag = 'bct/prysm-beacon:eip8321-minimal' },
    @{ Binary = 'validator'; Base = 'ethpandaops/prysm-validator:heze-devnet-2-minimal-3cc4d4e'; Tag = 'bct/prysm-validator:eip8321-minimal' }
)
foreach ($target in $targets) {
    & $DockerPath build --build-arg "BASE=$($target.Base)" --build-arg "BINARY=$($target.Binary)" --build-arg "SOURCE_REVISION=$revision" -f (Join-Path $PSScriptRoot 'Dockerfile') -t $target.Tag (Join-Path $OutputPath 'bin')
    if ($LASTEXITCODE -ne 0) { throw "Image build failed: $($target.Tag)" }
    & $DockerPath run --rm $target.Tag --version
    if ($LASTEXITCODE -ne 0) { throw "Image could not start: $($target.Tag)" }
}
Get-Content -LiteralPath (Join-Path $OutputPath 'binaries.sha256')
