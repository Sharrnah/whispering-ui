# Fyne increments FyneApp.toml AFTER packaging. Run in an isolated source copy
# so Windows and Linux both consume the same, explicitly chosen Version.Build.
$ErrorActionPreference = 'Stop'
$uiRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$temporaryRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
$buildRoot = Join-Path $temporaryRoot ('wt-ui-windows-' + [guid]::NewGuid().ToString('N'))
$originalLocation = Get-Location

try {
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
