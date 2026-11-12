$ErrorActionPreference = 'Stop'
$Project = Split-Path -Parent $MyInvocation.MyCommand.Path
$Dist = Join-Path $Project 'dist'
New-Item -ItemType Directory -Force $Dist | Out-Null

$Version = '1.0.0'
$Targets = @(
    @{ GOOS = 'windows'; GOARCH = 'amd64'; File = 's5levelctl-windows-amd64.exe' },
    @{ GOOS = 'linux';   GOARCH = 'amd64'; File = 's5levelctl-linux-amd64' },
    @{ GOOS = 'linux';   GOARCH = 'arm64'; File = 's5levelctl-linux-arm64' }
)

Push-Location $Project
try {
    foreach ($Target in $Targets) {
        $env:GOOS = $Target.GOOS
        $env:GOARCH = $Target.GOARCH
        $env:CGO_ENABLED = '0'
        go build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $Dist $Target.File) ./cmd/s5levelctl
    }
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}

Write-Host "Builds wurden unter $Dist erstellt."
