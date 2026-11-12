package printer

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Runner interface {
	Run(command, stdin string) (string, error)
}

type Result struct {
	Firmware string `json:"firmware"`
	Path     string `json:"path"`
	Mode     string `json:"mode"`
	Backup   string `json:"backup,omitempty"`
	Changed  bool   `json:"changed"`
}

func Execute(r Runner, action string) (Result, error) {
	out, err := r.Run("python3 - "+action, remoteScript)
	if err != nil {
		return Result{}, err
	}
	var result Result
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		return Result{}, fmt.Errorf("unerwartete Antwort des Druckers: %w (%s)", err, out)
	}
	if result.Mode != "DETAILED" && result.Mode != "NEVER" {
		return Result{}, fmt.Errorf("nicht unterstützter Leveling-Modus %q; es wurde nichts verändert", result.Mode)
	}
	return result, nil
}

const remoteScript = `
import datetime
import json
import os
import re
import shutil
import sys
import tempfile

TARGET = "/usr/share/griffin/griffin/printer/procedures/pre_and_post_print/auto_bed_level_adjust/alignZAxisProcedure.py"
ACTION = sys.argv[1] if len(sys.argv) > 1 else "status"

def fail(message):
    print(message, file=sys.stderr)
    raise SystemExit(2)

if ACTION not in ("status", "diagnose", "disable", "enable", "backup"):
    fail("Unbekannte Aktion")
if not os.path.isfile(TARGET):
    fail("Erwartete Griffin-Datei fehlt: " + TARGET)

with open(TARGET, "r", encoding="utf-8") as handle:
    source = handle.read()

pattern = re.compile(r"^(\s*self\.__probing_mode\s*=\s*ProbeMode\.)(DETAILED|NEVER)(\s*(?:#.*)?)$", re.MULTILINE)
matches = list(pattern.finditer(source))
if len(matches) != 1:
    fail("Sicherheitsabbruch: genau eine unterstützte __probing_mode-Zuweisung erwartet, gefunden: %d" % len(matches))

old_mode = matches[0].group(2)
new_mode = old_mode
if ACTION == "disable":
    new_mode = "NEVER"
elif ACTION == "enable":
    new_mode = "DETAILED"

backup = ""
changed = False
if ACTION == "backup" or new_mode != old_mode:
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    backup = TARGET + ".s5levelctl-backup-" + stamp
    shutil.copy2(TARGET, backup)

if new_mode != old_mode:
    replacement = matches[0].group(1) + new_mode + matches[0].group(3)
    updated = source[:matches[0].start()] + replacement + source[matches[0].end():]
    directory = os.path.dirname(TARGET)
    fd, temporary = tempfile.mkstemp(prefix=".s5levelctl-", dir=directory)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write(updated)
            handle.flush()
            os.fsync(handle.fileno())
        stat = os.stat(TARGET)
        os.chmod(temporary, stat.st_mode)
        os.chown(temporary, stat.st_uid, stat.st_gid)
        os.replace(temporary, TARGET)
        changed = True
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)

firmware = ""
try:
    with open("/etc/ultimaker_firmware", "r", encoding="utf-8") as handle:
        firmware = handle.read().strip()
except OSError:
    firmware = "nicht ermittelbar"

print(json.dumps({
    "firmware": firmware,
    "path": TARGET,
    "mode": new_mode,
    "backup": backup,
    "changed": changed,
}, separators=(",", ":")))
`
