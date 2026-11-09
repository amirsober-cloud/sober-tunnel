# Regenerate the panel catalog from the native CLI flag declarations.
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$catalog = [ordered]@{}
$allowed = @('setup-foreign','add-node','edit-foreign','edit-node','remove-node','probe','test','speedtest','pairing')
foreach ($file in Get-ChildItem "$repoRoot/engines/hedioum/cmd/hedioum" -Filter '*.go') {
    $text = [IO.File]::ReadAllText($file.FullName)
    foreach ($part in [regex]::Split($text, '(?m)^func ')) {
        $cmd = [regex]::Match($part, 'flag.NewFlagSet\("([^"]+)"')
        if (!$cmd.Success -or $cmd.Groups[1].Value -notin $allowed) { continue }
        $flags = @()
        foreach ($match in [regex]::Matches($part, 'fs\.(String|Int|Bool)\("([^"]+)",\s*([^,\r\n]+),\s*("(?:[^"\\]|\\.)*")\)')) {
            $name = $match.Groups[2].Value
            if ($name -eq 'move-ssh') { continue }
            $raw = $match.Groups[3].Value.Trim()
            $default = ''
            if ($raw -match '^".*"$') { $default = $raw | ConvertFrom-Json }
            elseif ($raw -match '^(-?\d+|true|false)$') { $default = $raw }
            $flags += [ordered]@{name=$name;type=$match.Groups[1].Value.ToLower();default=$default;help=($match.Groups[4].Value | ConvertFrom-Json)}
        }
        $catalog[$cmd.Groups[1].Value] = @($flags)
    }
}
foreach ($command in $allowed) { if (!$catalog.Contains($command)) { throw "Missing command: $command" } }
$catalog['setup-iran'] = $catalog['add-node']
$catalog['check-ip'] = @()
$json = ConvertTo-Json -InputObject $catalog -Depth 6
[IO.File]::WriteAllText("$repoRoot/engines/backpack/internal/mimicmgr/options.json", $json.Replace("`r`n","`n")+"`n", (New-Object Text.UTF8Encoding($false)))
