param(
    [ValidateSet('linux', 'windows', 'all')][string]$Target = 'all',
    [ValidateSet('cu128', 'cpu')][string]$Flavor = 'cu128',
    [switch]$Release
)
$ErrorActionPreference = 'Stop'
$uiRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$outputRoot = Join-Path $uiRoot ('Build\docker\' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
New-Item -ItemType Directory -Path $outputRoot -Force | Out-Null
function Invoke-DockerChecked {
    & docker @args
    if ($LASTEXITCODE -ne 0) { throw "Docker command failed ($LASTEXITCODE)" }
}
Invoke-DockerChecked build -t whispering-tiger-ui-builder:go1.26.5 -f (Join-Path $PSScriptRoot 'Dockerfile') $PSScriptRoot
$targets = if ($Target -eq 'all') { @('linux', 'windows') } else { @($Target) }
foreach ($platform in $targets) {
    $binary = if ($platform -eq 'windows') { 'Whispering Tiger.exe' } else { 'whispering-tiger-linux-amd64' }
    $buildArgs = @('--target', $platform, '--flavor', $Flavor, '--test', '--package', '--output', "/output/$platform/$binary")
    if ($Release) { $buildArgs += '--release' }
    Invoke-DockerChecked run --rm --init `
        --mount "type=bind,source=$uiRoot,target=/source,readonly" `
        --mount "type=bind,source=$outputRoot,target=/output" `
        --mount 'type=volume,source=wt-linux-go-mod,target=/go/pkg/mod' `
        --mount 'type=volume,source=wt-linux-go-build,target=/root/.cache/go-build' `
        whispering-tiger-ui-builder:go1.26.5 @buildArgs
}
Write-Host "UI builds: $outputRoot"
