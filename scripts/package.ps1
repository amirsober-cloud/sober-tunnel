param(
    [ValidateSet('amd64', 'arm64')]
    [string]$Architecture = 'amd64',
    [string]$ReleasePublicKey = ''
)
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$buildDir = Join-Path $repoRoot 'build'

& (Join-Path $PSScriptRoot 'build.ps1') -Architecture $Architecture -ReleasePublicKey $ReleasePublicKey

$archiveName = "sober-tunnel_linux_$Architecture.tar.gz"
$archivePath = Join-Path $buildDir $archiveName
Copy-Item -LiteralPath (Join-Path $env:LIBPCAP_PREFIX 'LICENSE.libpcap') -Destination (Join-Path $buildDir 'LICENSE.libpcap')
tar -czf $archivePath -C $buildDir sober-tunnel sober-tunnel-mimic LICENSE.libpcap -C $repoRoot LICENSE NOTICE engines/backpack/LICENSE engines/camouflage/LICENSE engines/backpack/internal/thirdparty/kcp-go/LICENSE engines/backpack/internal/thirdparty/smux/LICENSE
if ($LASTEXITCODE -ne 0) { throw 'Release archive creation failed' }

$checksums = Get-ChildItem -LiteralPath $buildDir -Filter 'sober-tunnel_linux_*.tar.gz' |
    Sort-Object Name |
    ForEach-Object {
        $hash = (Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
        "$hash  $($_.Name)"
    }
[IO.File]::WriteAllText((Join-Path $buildDir 'SHA256SUMS'), ($checksums -join "`n") + "`n", [Text.UTF8Encoding]::new($false))
Write-Output "Created $archivePath with both binaries, licenses and SHA256SUMS."
