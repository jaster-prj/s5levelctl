$ErrorActionPreference = 'Stop'
$Project = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Exe = Join-Path $Project 'dist/s5levelctl-e2e.exe'
$TestConfig = Join-Path $env:TEMP 's5levelctl-e2e-config'

function Invoke-S5LevelCtl {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]] $Arguments)
    $output = & $Exe @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) { throw ($output -join "`n") }
    return ($output -join "`n")
}

Push-Location $Project
try {
    docker compose up -d --build
    if ($LASTEXITCODE -ne 0) { throw 'Firmware-Dummy konnte nicht gestartet werden.' }

    $ready = $false
    for ($i = 0; $i -lt 30; $i++) {
        $health = docker inspect --format '{{.State.Health.Status}}' s5levelctl-firmware-dummy 2>$null
        if ($health -eq 'healthy') { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) { throw 'Firmware-Dummy wurde nicht rechtzeitig bereit.' }

    docker exec s5levelctl-firmware-dummy reset-dummy
    if ($LASTEXITCODE -ne 0) { throw 'Dummy-Reset fehlgeschlagen.' }

    New-Item -ItemType Directory -Force (Split-Path $Exe) | Out-Null
    go build -o $Exe ./cmd/s5levelctl
    if ($LASTEXITCODE -ne 0) { throw 'Go-Build fehlgeschlagen.' }

    Remove-Item -Recurse -Force $TestConfig -ErrorAction SilentlyContinue
    $oldAppData = $env:APPDATA
    $env:APPDATA = $TestConfig
    try {
        $status = Invoke-S5LevelCtl status --host 127.0.0.1 --port 2222 --accept-new-host-key
        if ($status -notmatch 'AKTIVIERT' -or $status -notmatch '9.0.5') { throw "Status-Test fehlgeschlagen:`n$status" }

        $disabled = Invoke-S5LevelCtl disable --host 127.0.0.1 --port 2222 --restart
        if ($disabled -notmatch 'DEAKTIVIERT' -or $disabled -notmatch 'Sicherung') { throw "Disable-Test fehlgeschlagen:`n$disabled" }

        $backupCount = docker exec s5levelctl-firmware-dummy sh -c 'find /usr/share/griffin -name "*.s5levelctl-backup-*" | wc -l'
        if ([int]$backupCount -lt 1) { throw 'Es wurde keine Sicherung im Dummy gefunden.' }
        docker exec s5levelctl-firmware-dummy test -s /var/log/griffin-restarts.log
        if ($LASTEXITCODE -ne 0) { throw 'Der simulierte Griffin-Neustart wurde nicht registriert.' }

        $enabled = Invoke-S5LevelCtl enable --host 127.0.0.1 --port 2222
        if ($enabled -notmatch 'AKTIVIERT') { throw "Enable-Test fehlgeschlagen:`n$enabled" }
    } finally {
        $env:APPDATA = $oldAppData
    }

    Write-Host 'E2E-Test erfolgreich: status -> disable/backup/restart -> enable'
    Write-Host 'Der langlebige Dummy läuft weiter: s5levelctl-firmware-dummy (Port 2222).'
} finally {
    Pop-Location
}
