# S5LevelCtl

`s5levelctl` ist ein kleines Go-CLI zum kontrollierten Deaktivieren und erneuten
Aktivieren des automatischen Bed-Levelings eines UltiMaker S5 R2. Es orientiert
sich am Leveling-Verfahren von UltiTuner und ist für Firmware 9.0.5 vorgesehen.

> **Achtung:** Dies ist keine offiziell von UltiMaker unterstützte Änderung.
> Die Nutzung erfolgt auf eigene Verantwortung. Nach deaktiviertem Active
> Leveling muss das Druckbett manuell nivelliert werden. Ein Firmware-Update
> kann die Änderung überschreiben.

## Sicherheitsmechanismen

- `status` und `diagnose` verändern nichts.
- Änderungen erfolgen nur, wenn exakt die bekannte Griffin-Zuweisung
  `ProbeMode.DETAILED` oder `ProbeMode.NEVER` gefunden wird.
- Vor jeder tatsächlichen Änderung wird eine datierte Sicherung direkt neben
  der Originaldatei auf dem Drucker angelegt.
- Die Datei wird atomar ersetzt; Rechte und Besitzer bleiben erhalten.
- SSH-Hostkeys werden lokal gepinnt. Ein geänderter Fingerprint wird abgelehnt.
- Der Griffin-Dienst wird nur mit dem expliziten Schalter `--restart` neu gestartet.

## Voraussetzungen

1. Am Drucker den **Developer Mode** aktivieren.
2. Drucker und Rechner müssen sich im selben vertrauenswürdigen Netzwerk befinden.
3. Während `disable`, `enable` oder eines Dienstneustarts darf kein Druck laufen.

## Verwendung

Beim ersten Kontakt zunächst den auf dem Bildschirm ausgegebenen Fingerprint
prüfen. Danach darf er bewusst gespeichert werden:

```powershell
.\dist\s5levelctl-windows-amd64.exe status --host 192.168.1.50
.\dist\s5levelctl-windows-amd64.exe status --host 192.168.1.50 --accept-new-host-key
```

Leveling deaktivieren und Griffin neu starten:

```powershell
.\dist\s5levelctl-windows-amd64.exe disable --host 192.168.1.50 --restart
```

Status prüfen oder die Originalfunktion wieder aktivieren:

```powershell
.\dist\s5levelctl-windows-amd64.exe status --host 192.168.1.50
.\dist\s5levelctl-windows-amd64.exe enable --host 192.168.1.50 --restart
```

Das Standardpasswort des Developer Mode ist `ultimaker`. Für ein abweichendes
Passwort wird die Umgebungsvariable verwendet, damit es nicht in der
Shell-Historie landet:

```powershell
$env:S5LEVELCTL_PASSWORD = 'mein-passwort'
```

Die Fingerprints liegen unter dem benutzerspezifischen Konfigurationsordner in
`s5levelctl/known_hosts.json`.

## Bauen

Windows PowerShell:

```powershell
.\build.ps1
```

Linux/macOS:

```sh
./build.sh
```

Erzeugt werden Builds für Windows amd64, Linux amd64 und Linux arm64 unter
`dist/`. Das Programm läuft auf dem PC und verbindet sich per SSH zum Drucker;
es muss nicht auf dem Drucker installiert werden.

## Firmware-Dummy und End-to-End-Test

Unter `test/firmware-dummy/` liegt ein Container, der die für S5LevelCtl
benötigten Teile der Firmware 9.0.5 nachbildet:

- SSH als `root` mit Developer-Mode-Passwort `ultimaker`
- `/etc/ultimaker_firmware`
- den originalen Griffin-Pfad der `alignZAxisProcedure.py`
- einen kontrollierten Dummy für `systemctl restart griffin.printer`
- persistente Dateiänderungen und einen `reset-dummy`-Befehl

Der vollständige Test baut und startet den langlebigen Container, setzt seinen
Zustand zurück und prüft Status, Deaktivierung, Sicherung, Dienstneustart und
erneute Aktivierung:

```powershell
.\test\e2e.ps1
```

Der Dummy bleibt danach absichtlich unter dem Namen
`s5levelctl-firmware-dummy` auf Port `2222` aktiv und wird bei weiteren Tests
wiederverwendet. Manuelle Aufrufe sind damit ebenfalls möglich:

```powershell
docker compose up -d --build
.\dist\s5levelctl-windows-amd64.exe status --host 127.0.0.1 --port 2222 --accept-new-host-key
docker exec s5levelctl-firmware-dummy reset-dummy
```

Beenden ohne Löschen: `docker stop s5levelctl-firmware-dummy`. Der Container
wird nicht automatisch entfernt.

## Rückkehr zum Originalzustand

Normalerweise genügt `enable --restart`. Zusätzlich legt das Programm vor
jeder Änderung eine Datei der Form
`alignZAxisProcedure.py.s5levelctl-backup-<Zeitstempel>` auf dem Drucker an.
Bereits vorhandene Sicherungen mit dem alten Präfix `.ultilevel-backup-` bleiben
unangetastet und können weiterhin manuell verwendet werden.
Die Sicherung wird absichtlich nicht automatisch zurückkopiert, weil nach einem
Firmware-Update eine alte vollständige Python-Datei inkompatibel sein könnte.
