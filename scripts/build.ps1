param([string]$Architecture = 'amd64', [string]$ReleasePublicKey = '')
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
if (!$ReleasePublicKey -and (Test-Path -LiteralPath (Join-Path $repoRoot 'docs/release-public-key.txt'))) {
    $ReleasePublicKey = [IO.File]::ReadAllText((Join-Path $repoRoot 'docs/release-public-key.txt')).Trim()
}
$oldOS = $env:GOOS; $oldArch = $env:GOARCH; $oldCGO = $env:CGO_ENABLED
try {
    $env:GOOS = 'linux'; $env:GOARCH = $Architecture; $env:CGO_ENABLED = '0'
    New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot 'build') | Out-Null
    Push-Location (Join-Path $repoRoot 'engines/backpack')
    try {
        $link = '-s -w'
        if ($ReleasePublicKey) { $link += " -X github.com/backpack/backpack/internal/app.ReleasePublicKey=$ReleasePublicKey" }
        go build -trimpath -ldflags $link -o (Join-Path $repoRoot 'build/sober-tunnel') .
        if ($LASTEXITCODE -ne 0) { throw 'sober-tunnel build failed' }
        Set-Location (Join-Path $repoRoot 'engines/camouflage')
        go build -p 1 -trimpath -ldflags '-s -w' -o (Join-Path $repoRoot 'build/sober-tunnel-mimic') ./cmd/camouflage
        if ($LASTEXITCODE -ne 0) { throw 'sober-tunnel native tools build failed' }
    } finally { Pop-Location }
} finally { $env:GOOS = $oldOS; $env:GOARCH = $oldArch; $env:CGO_ENABLED = $oldCGO }
