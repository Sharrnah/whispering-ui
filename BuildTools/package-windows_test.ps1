# Runs without a Windows compiler: fake Fyne reproduces its metadata increment.
# Real Windows resource packaging is checked separately in Docker with MinGW.
$ErrorActionPreference = 'Stop'
$fixtureRoot = Join-Path ([IO.Path]::GetTempPath()) ('wt-package-test-' + [guid]::NewGuid().ToString('N'))
$helper = Join-Path $PSScriptRoot 'package-windows.ps1'
New-Item -ItemType Directory -Path (Join-Path $fixtureRoot 'BuildTools') -Force | Out-Null
Copy-Item -LiteralPath $helper -Destination (Join-Path $fixtureRoot 'BuildTools/package-windows.ps1')
$global:wtPackageCopies = @()
$global:wtPackageFailure = $false
$originalModuleCache = $env:GOMODCACHE
$env:GOMODCACHE = Join-Path $fixtureRoot 'module cache'
$goRoot = Join-Path $env:GOMODCACHE 'golang.org\toolchain@v0.0.1-go1.26.5.windows-amd64'
$selectedGo = Join-Path $goRoot 'bin\go.exe'
$originalGoEnvironment = @{}
foreach ($name in @('GO', 'GOROOT', 'GOTOOLCHAIN', 'PATH')) {
    $originalGoEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
$global:wtPackageGoRoot = $goRoot
$global:wtPackageGoVersion = 'go1.26.5'
function global:go {
    throw 'The old Go first on PATH must not be invoked when the required toolchain is cached'
}
Set-Item -LiteralPath "Function:global:$selectedGo" -Value {
    $global:LASTEXITCODE = 0
    "go version $global:wtPackageGoVersion windows/amd64"
}
function global:git {
    $global:LASTEXITCODE = 0
    if ($args -contains '--others') {
        'new.go', 'Resources/new.txt', 'profile.yaml', 'old.zip'
    } else {
        'FyneApp.toml', 'main.go', 'deleted.go', 'app-icon.png', 'LICENSE', 'BuildTools/package-windows.ps1'
    }
}
function global:fyne {
    if ($env:GO -ne (Join-Path $global:wtPackageGoRoot 'bin\go.exe') -or $env:GOTOOLCHAIN -ne 'local') {
        throw 'Fyne did not receive the selected compiler'
    }
    if (($args -join ' ') -ne 'package --release') { throw 'Unexpected Fyne command.' }
    $global:wtPackageCopies += (Get-Location).Path
    foreach ($required in @('FyneApp.toml', 'main.go', 'new.go', 'Resources/new.txt', 'app-icon.png', 'LICENSE')) {
        if (-not (Test-Path -LiteralPath $required)) { throw "Missing source: $required" }
    }
    foreach ($excluded in @('profile.yaml', 'old.zip', 'deleted.go')) {
        if (Test-Path -LiteralPath $excluded) { throw "Unexpected source: $excluded" }
    }
    $toml = Get-Content -LiteralPath 'FyneApp.toml' -Raw
    Set-Content -LiteralPath 'Whispering Tiger.exe' -Value $toml -NoNewline
    Set-Content -LiteralPath 'FyneApp.toml' -Value ($toml -replace 'Build = 1', 'Build = 2') -NoNewline
    $global:LASTEXITCODE = if ($global:wtPackageFailure) { 23 } else { 0 }
}
try {
    New-Item -ItemType Directory -Path (Split-Path $selectedGo -Parent) -Force | Out-Null
    Set-Content -LiteralPath $selectedGo -Value 'fake compiler'
    Set-Content -LiteralPath (Join-Path $fixtureRoot 'go.mod') -Value "module fixture`ngo 1.26.0`ntoolchain go1.26.5`n"
    & (Join-Path $fixtureRoot 'BuildTools/package-windows.ps1') -CheckOnly
    if ($global:wtPackageCopies.Count) { throw 'Preflight started packaging' }
    $global:wtPackageGoVersion = 'go1.20.0'
    $caught = ''
    try { & (Join-Path $fixtureRoot 'BuildTools/package-windows.ps1') -CheckOnly }
    catch { $caught = $_.Exception.Message }
    if ($caught -notlike 'Expected go1.26.5 windows/amd64*') { throw 'Old compiler was not rejected by preflight' }
    $global:wtPackageGoVersion = 'go1.26.5'
    $toml = "[Details]`nVersion = `"1.3.11`"`nBuild = 1`n"
    Set-Content -LiteralPath (Join-Path $fixtureRoot 'FyneApp.toml') -Value $toml -NoNewline
    foreach ($name in @('main.go', 'new.go', 'app-icon.png', 'LICENSE', 'profile.yaml', 'old.zip')) {
        Set-Content -LiteralPath (Join-Path $fixtureRoot $name) -Value $name
    }
    New-Item -ItemType Directory -Path (Join-Path $fixtureRoot 'Resources') | Out-Null
    Set-Content -LiteralPath (Join-Path $fixtureRoot 'Resources/new.txt') -Value 'resource'
    foreach ($fail in @($false, $false, $true)) {
        $global:wtPackageFailure = $fail
        $caught = $false
        try {
            & (Join-Path $fixtureRoot 'BuildTools/package-windows.ps1')
        } catch {
            if (-not $fail -or $_ -notmatch 'packaging failed \(23\)') { throw }
            $caught = $true
        }
        if ($caught -ne $fail) { throw 'Build failure was not propagated.' }
        foreach ($name in $originalGoEnvironment.Keys) {
            if ([Environment]::GetEnvironmentVariable($name, 'Process') -cne $originalGoEnvironment[$name]) {
                throw "Build changed the caller's $name environment"
            }
        }
        foreach ($name in @('FyneApp.toml', 'Whispering Tiger.exe')) {
            if ((Get-Content -LiteralPath (Join-Path $fixtureRoot $name) -Raw) -cne $toml) {
                throw "Repeated/failed build changed the release version in $name"
            }
        }
        foreach ($copy in $global:wtPackageCopies) {
            if (Test-Path -LiteralPath $copy) { throw "Build copy was not cleaned: $copy" }
        }
    }
    Write-Host 'PASS: source selection, repeated builds, stable version, failure propagation and temporary cleanup.'
} finally {
    $env:GOMODCACHE = $originalModuleCache
    Remove-Item Function:\git, Function:\fyne, Function:\go
    Remove-Item -LiteralPath "Function:global:$selectedGo"
    $resolvedFixture = [IO.Path]::GetFullPath($fixtureRoot)
    $tempPrefix = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
    if (-not $resolvedFixture.StartsWith($tempPrefix, [StringComparison]::OrdinalIgnoreCase) -or
        (Split-Path -Leaf $resolvedFixture) -notmatch '^wt-package-test-[0-9a-f]{32}$') {
        throw "Unexpected fixture path: $resolvedFixture"
    }
    Remove-Item -LiteralPath $resolvedFixture -Recurse -Force
    Remove-Variable wtPackageCopies, wtPackageFailure, wtPackageGoRoot, wtPackageGoVersion -Scope Global
}
