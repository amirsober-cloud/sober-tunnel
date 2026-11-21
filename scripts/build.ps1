param([string]$Architecture = 'amd64', [string]$ReleasePublicKey = '')
$ErrorActionPreference = 'Stop'
if (!$env:CC -or !$env:LIBPCAP_PREFIX) { throw 'SoberLinkPack v2 needs a Linux-target C compiler and static libpcap. Build on Linux/CI, or set CC and LIBPCAP_PREFIX for cross-compilation.' }
$repoRoot = Split-Path -Parent $PSScriptRoot
if (!$ReleasePublicKey -and (Test-Path -LiteralPath (Join-Path $repoRoot 'docs/release-public-key.txt'))) {
    $ReleasePublicKey = [IO.File]::ReadAllText((Join-Path $repoRoot 'docs/release-public-key.txt')).Trim()
}
$oldOS = $env:GOOS; $oldArch = $env:GOARCH; $oldCGO = $env:CGO_ENABLED
try {
    $env:GOOS = 'linux'; $env:GOARCH = $Architecture; $env:CGO_ENABLED = '1'
    New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot 'build') | Out-Null
    Push-Location (Join-Path $repoRoot 'engines/backpack')
    try {
        $link = '-s -w'
        if ($ReleasePublicKey) { $link += " -X github.com/backpack/backpack/internal/app.ReleasePublicKey=$ReleasePublicKey" }
        $env:CGO_CFLAGS = "-I$env:LIBPCAP_PREFIX/include"
        $env:CGO_LDFLAGS = "-L$env:LIBPCAP_PREFIX/lib -lpcap"
        go build -tags netgo,osusergo -trimpath -ldflags ($link + ' -linkmode external -extldflags -static') -o (Join-Path $repoRoot 'build/sober-tunnel') .
        if ($LASTEXITCODE -ne 0) { throw 'sober-tunnel build failed' }
        $env:CGO_ENABLED = '0'
        Set-Location (Join-Path $repoRoot 'engines/camouflage')
        go build -p 1 -trimpath -ldflags '-s -w' -o (Join-Path $repoRoot 'build/sober-tunnel-mimic') ./cmd/camouflage
        if ($LASTEXITCODE -ne 0) { throw 'sober-tunnel native tools build failed' }
    } finally { Pop-Location }
} finally { $env:GOOS = $oldOS; $env:GOARCH = $oldArch; $env:CGO_ENABLED = $oldCGO }
