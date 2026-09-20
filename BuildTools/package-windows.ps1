# Fyne increments FyneApp.toml AFTER packaging. Run in an isolated source copy
# so Windows and Linux both consume the same, explicitly chosen Version.Build.
param([switch]$CheckOnly)
$ErrorActionPreference = 'Stop'
$uiRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$temporaryRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
$buildRoot = Join-Path $temporaryRoot ('wt-ui-windows-' + [guid]::NewGuid().ToString('N'))
$originalLocation = Get-Location
$originalGoEnvironment = @{}
foreach ($name in @('GO', 'GOROOT', 'GOTOOLCHAIN', 'PATH')) {
    $originalGoEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

try {
    # Resolve the module's pinned toolchain before invoking Fyne. Its child Go
    # processes must not rediscover an older compiler from another PATH entry.
    $module = Get-Content -LiteralPath (Join-Path $uiRoot 'go.mod') -Raw
    $match = [regex]::Match($module, '(?m)^toolchain\s+(go\d+\.\d+\.\d+)\s*$')
    if (-not $match.Success) { throw 'go.mod must declare the release Go toolchain.' }
    $requiredToolchain = $match.Groups[1].Value
    $cacheRoots = @($env:GOMODCACHE)
    $goPaths = @($env:GOPATH -split [IO.Path]::PathSeparator)
    $goPaths += Join-Path ([Environment]::GetFolderPath('UserProfile')) 'go'
    foreach ($goPath in $goPaths) {
        if ($goPath) { $cacheRoots += Join-Path $goPath 'pkg\mod' }
    }
    $compilerCandidates = @()
    foreach ($cacheRoot in ($cacheRoots | Where-Object { $_ } | Select-Object -Unique)) {
        $compilerCandidates += Join-Path $cacheRoot "golang.org\toolchain@v0.0.1-$requiredToolchain.windows-amd64\bin\go.exe"
    }
    $compilerCandidates += Join-Path ([Environment]::GetFolderPath('UserProfile')) "sdk\$requiredToolchain\bin\go.exe"
    $selectedCompiler = $compilerCandidates | Where-Object { Test-Path -LiteralPath $_ -PathType Leaf } | Select-Object -First 1
    # A cached toolchain does not require the caller's old Go launcher at all.
    # Only a missing cache needs a Go 1.21+ bootstrap for automatic download.
    $env:GOROOT = $null
    $env:GOTOOLCHAIN = 'local'
    if (-not $selectedCompiler) {
        Set-Location -LiteralPath $temporaryRoot
        $bootstrap = $null
        foreach ($command in @(Get-Command go -All -ErrorAction SilentlyContinue)) {
            $candidate = if ($command.CommandType -eq 'Application') { $command.Source } else { $command.Name }
            $candidateVersion = & $candidate version
            if ($LASTEXITCODE -eq 0 -and ($candidateVersion -join '') -match '^go version go(\d+)\.(\d+)') {
                if ([int]$Matches[1] -gt 1 -or ([int]$Matches[1] -eq 1 -and [int]$Matches[2] -ge 21)) {
                    $bootstrap = $candidate
                    break
                }
            }
        }
        if (-not $bootstrap) {
            throw "No cached $requiredToolchain found, and PATH has no Go 1.21+ bootstrap. Install the UI toolchain or set GOMODCACHE to its existing cache directory."
        }
        Set-Location -LiteralPath $uiRoot
        $env:GOTOOLCHAIN = $requiredToolchain
        $goRoot = & $bootstrap env GOROOT
        if ($LASTEXITCODE -ne 0) { throw "Could not resolve $requiredToolchain using $bootstrap." }
        $selectedCompiler = Join-Path (($goRoot -join '').Trim()) 'bin\go.exe'
    }
    $env:GO = $selectedCompiler
    if (-not (Test-Path -LiteralPath $env:GO)) { throw "Missing selected Go compiler: $env:GO" }
    $env:GOROOT = Split-Path -Parent (Split-Path -Parent $env:GO)
    $env:PATH = (Split-Path -Parent $env:GO) + [IO.Path]::PathSeparator + $env:PATH
    $env:GOTOOLCHAIN = 'local'
    $goVersion = & $env:GO version
    if ($LASTEXITCODE -ne 0 -or ($goVersion -join '') -notmatch ('^go version ' + [regex]::Escape($requiredToolchain) + ' windows/amd64$')) {
        throw "Expected $requiredToolchain windows/amd64; got $goVersion"
    }
    Write-Host "Windows UI compiler: $goVersion ($env:GO)"
    if ($CheckOnly) { return }

    # Match BuildTools/build.py: current tracked files plus new Go sources,
    # embedded resources and build helpers; no local profiles or old binaries.
    $tracked = @(& git -c core.quotepath=false -C $uiRoot ls-files)
    if ($LASTEXITCODE -ne 0) { throw 'Could not list tracked UI source files.' }
    $untracked = @(& git -c core.quotepath=false -C $uiRoot ls-files --others --exclude-standard)
    if ($LASTEXITCODE -ne 0) { throw 'Could not list new UI source files.' }
    $newSources = @($untracked | Where-Object { $_ -match '\.go$|^(Resources|BuildTools)/' })
    New-Item -ItemType Directory -Path $buildRoot | Out-Null
    foreach ($name in (($tracked + $newSources) | Sort-Object -Unique)) {
        $source = Join-Path $uiRoot $name
        if (-not (Test-Path -LiteralPath $source -PathType Leaf)) { continue }
        $destination = [IO.Path]::GetFullPath((Join-Path $buildRoot $name))
        if (-not $destination.StartsWith($buildRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
            throw "Source path is outside the build copy: $name"
        }
        New-Item -ItemType Directory -Path (Split-Path -Parent $destination) -Force | Out-Null
        Copy-Item -LiteralPath $source -Destination $destination
    }

    Set-Location -LiteralPath $buildRoot
    & fyne package --release
    if ($LASTEXITCODE -ne 0) { throw "Fyne Windows packaging failed ($LASTEXITCODE)." }
    $executable = Join-Path $buildRoot 'Whispering Tiger.exe'
    if (-not (Test-Path -LiteralPath $executable -PathType Leaf)) {
        throw 'Fyne did not produce Whispering Tiger.exe.'
    }
    Copy-Item -LiteralPath $executable -Destination (Join-Path $uiRoot 'Whispering Tiger.exe') -Force
    Write-Host "Windows executable: $(Join-Path $uiRoot 'Whispering Tiger.exe')"
    Write-Host 'Version.Build comes from FyneApp.toml; the repository build number was not incremented.'
} finally {
    Set-Location -LiteralPath $originalLocation.Path
    foreach ($name in $originalGoEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $originalGoEnvironment[$name], 'Process')
    }
    # Delete only this invocation's verified temporary directory.
    $resolvedBuildRoot = [IO.Path]::GetFullPath($buildRoot)
    $temporaryPrefix = $temporaryRoot.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
    if (-not $resolvedBuildRoot.StartsWith($temporaryPrefix, [StringComparison]::OrdinalIgnoreCase) -or
        (Split-Path -Leaf $resolvedBuildRoot) -notmatch '^wt-ui-windows-[0-9a-f]{32}$') {
        throw "Refusing to clean an unexpected build directory: $resolvedBuildRoot"
    }
    if (Test-Path -LiteralPath $resolvedBuildRoot) {
        Remove-Item -LiteralPath $resolvedBuildRoot -Recurse -Force -ErrorAction Continue
    }
}
