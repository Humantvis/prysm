param([string]$PackagePath)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
if (-not $PackagePath) { $PackagePath = Join-Path (Split-Path $repo -Parent) 'ethereum-package' }
$revision = '5795b77e65235ad223d4d3b413b2739190bf60cc'
if (-not (Test-Path -LiteralPath $PackagePath)) {
    & git -c core.autocrlf=false clone https://github.com/ethpandaops/ethereum-package.git $PackagePath
    if ($LASTEXITCODE -ne 0) { throw 'Could not clone ethereum-package' }
    & git -C $PackagePath checkout --detach $revision
    if ($LASTEXITCODE -ne 0) { throw 'Could not select the pinned package revision' }
}
$actual = (& git -C $PackagePath rev-parse HEAD).Trim()
if ($actual -ne $revision) { throw "Expected ethereum-package $revision; found $actual. Use a separate directory." }
# Kurtosis excludes hidden directories; keep this clone outside .tools.
if ((Split-Path $PackagePath -Leaf).StartsWith('.')) { throw 'Use a non-hidden package directory' }
& git -C $PackagePath config core.autocrlf false
$extensions = @('.star','.tmpl','.yaml','.yml','.json','.md','.txt','.sh')
foreach ($file in (& git -C $PackagePath ls-files)) {
    if ([IO.Path]::GetExtension($file) -in $extensions) {
        $path = Join-Path $PackagePath $file
        $text = [IO.File]::ReadAllText($path)
        if ($text.Contains("`r`n")) { [IO.File]::WriteAllText($path, $text.Replace("`r`n", "`n")) }
    }
}
$patch = Join-Path $PSScriptRoot 'ethereum-package.patch'
& git -C $PackagePath apply --reverse --check --ignore-space-change $patch 2>$null
if ($LASTEXITCODE -ne 0) {
    & git -C $PackagePath apply --check --ignore-space-change $patch
    if ($LASTEXITCODE -ne 0) { throw 'Package has conflicting local changes; use a fresh clone' }
    & git -C $PackagePath apply --ignore-space-change $patch
    if ($LASTEXITCODE -ne 0) { throw 'Could not apply EIP-8321 package support' }
}
Write-Output $PackagePath
