$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$utf8 = New-Object System.Text.UTF8Encoding($false)
$engineRoot = Join-Path $repoRoot 'engines/backpack'
Get-ChildItem -LiteralPath $engineRoot -Recurse -File | Where-Object {
    ($_.Extension -in '.go', '.js', '.html', '.sh', '.yml' -or $_.Name -eq 'Makefile') -and $_.FullName -notmatch '[\\/]\.git[\\/]'
} | ForEach-Object {
    $before = [IO.File]::ReadAllText($_.FullName)
    # Keep origin references, legal attribution, module identifiers and wire
    # protocol labels. Only presentation names change here.
    $after = [regex]::Replace($before, 'Backpack|BackPack', {
        param($match)
        $start = $before.LastIndexOf("`n", $match.Index)
        $end = $before.IndexOf("`n", $match.Index)
        if ($end -lt 0) { $end = $before.Length }
        $line = $before.Substring($start + 1, $end - $start - 1)
        if ($line -match 'Based on BackPack|github\.com/AminMGMT/BackPack|X-Backpack-Fix') { return $match.Value }
        return 'sober-tunnel'
    })
    $after = $after.Replace('/etc/backpack', '/etc/sober-tunnel').Replace('/run/backpack', '/run/sober-tunnel')
    $after = $after.Replace('/usr/local/bin/backpack', '/usr/local/bin/sober-tunnel').Replace('/root/sober-tunnel', '/opt/sober-tunnel')
    $after = $after.Replace('backpack-webui', 'sober-tunnel-webui').Replace('backpack-monitor', 'sober-tunnel-monitor').Replace('backpack-proxy', 'sober-tunnel-proxy')
    $after = $after.Replace('sudo backpack', 'sudo sober-tunnel').Replace('backpack node', 'sober-tunnel node').Replace('backpack -c', 'sober-tunnel -c').Replace('backpack --', 'sober-tunnel --')
    $after = $after.Replace('backpack_linux_', 'sober-tunnel_linux_')
    if ($_.Name -eq 'Makefile') { $after = $after.Replace('backpack-linux-', 'sober-tunnel-linux-').Replace('dist/backpack', 'dist/sober-tunnel').Replace('-C dist backpack', '-C dist sober-tunnel').Replace('BIN      := backpack', 'BIN      := sober-tunnel') }
    $after = $after.Replace('raw.githubusercontent.com/AminMGMT/sober-tunnel', 'raw.githubusercontent.com/amirsober-cloud/sober-tunnel')
    if ($after -cne $before) { [IO.File]::WriteAllText($_.FullName, $after, $utf8) }
}
