$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = 'barnuri/ghx'
$baseUrl = "https://github.com/$repo/releases/latest/download"
$installDir = if ($env:GHX_INSTALL_DIR) { $env:GHX_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\ghx' }
$shim = $env:GHX_SHIM -ne '0'

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "ghx install: unsupported CPU $env:PROCESSOR_ARCHITECTURE" }
}
$asset = "ghx_windows_$arch.exe"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("ghx-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "Downloading $asset..."
    Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$asset" -OutFile "$tmp\ghx.exe"
    Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/checksums.txt" -OutFile "$tmp\checksums.txt"
    $line = Get-Content "$tmp\checksums.txt" | Where-Object { ($_ -split '\s+')[1] -eq $asset }
    if (-not $line) { throw "ghx install: no checksum for $asset" }
    $expected = ($line -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 "$tmp\ghx.exe").Hash.ToLower()
    if ($actual -ne $expected) { throw "ghx install: checksum mismatch for $asset" }

    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    $names = @('ghx.exe')
    if ($shim) { $names += 'gh.exe' }
    foreach ($name in $names) {
        $target = Join-Path $installDir $name
        if (Test-Path $target) { Move-Item -Force $target "$target.old" }
        Copy-Item "$tmp\ghx.exe" $target
        Write-Host "Installed $target"
    }
}
finally {
    Remove-Item -Recurse -Force $tmp
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
    [Security.Principal.WindowsBuiltInRole]::Administrator)
$scope = if ($isAdmin) { 'Machine' } else { 'User' }
$path = [Environment]::GetEnvironmentVariable('Path', $scope)
if (($path -split ';') -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable('Path', "$installDir;$path", $scope)
    Write-Host "Added $installDir to the $scope PATH. Open a new terminal."
}
if ($shim -and -not $isAdmin) {
    $machinePath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
    $systemGh = $machinePath -split ';' | Where-Object { $_ -and (Test-Path (Join-Path $_ 'gh.exe')) }
    if ($systemGh) {
        Write-Host "gh.exe is also in the system PATH ($systemGh), which Windows searches first."
        Write-Host "Re-run this script as Administrator so 'gh' runs ghx, or add to your PowerShell profile: Set-Alias gh ghx"
    }
}
Write-Host "Next: create $env:APPDATA\ghx\config.yaml (see https://github.com/$repo#configuration), then run: gh ghx which"
