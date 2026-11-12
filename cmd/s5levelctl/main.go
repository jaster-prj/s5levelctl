package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/local/s5levelctl/internal/printer"
	"github.com/local/s5levelctl/internal/sshclient"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage()
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Println("s5levelctl", version)
		return nil
	}

	action := args[0]
	switch action {
	case "status", "diagnose", "disable", "enable", "backup":
	default:
		return fmt.Errorf("unbekannter Befehl %q (erlaubt: status, diagnose, disable, enable, backup)", action)
	}

	fs := flag.NewFlagSet(action, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	host := fs.String("host", "", "IP-Adresse oder Hostname des Druckers")
	port := fs.Int("port", 22, "SSH-Port")
	user := fs.String("user", "root", "SSH-Benutzer")
	restart := fs.Bool("restart", false, "Griffin-Dienst nach Änderung neu starten")
	acceptKey := fs.Bool("accept-new-host-key", false, "neuen SSH-Hostkey lokal speichern")
	timeout := fs.Duration("timeout", 15*time.Second, "Verbindungs-Timeout")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *host == "" {
		return errors.New("--host fehlt")
	}
	if *port < 1 || *port > 65535 {
		return errors.New("--port muss zwischen 1 und 65535 liegen")
	}
	if *restart && action != "disable" && action != "enable" {
		return errors.New("--restart ist nur mit disable oder enable erlaubt")
	}
	password := os.Getenv("S5LEVELCTL_PASSWORD")
	if password == "" {
		password = "ultimaker"
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	client, err := sshclient.Dial(ctx, sshclient.Options{
		Host: *host, Port: *port, User: *user, Password: password, AcceptNewHostKey: *acceptKey,
	})
	if err != nil {
		return err
	}
	defer client.Close()

	result, err := printer.Execute(client, action)
	if err != nil {
		return err
	}
	printResult(result)

	if *restart && result.Changed {
		fmt.Println("Griffin-Dienst wird neu gestartet …")
		if _, err := client.Run("systemctl restart griffin.printer", ""); err != nil {
			return fmt.Errorf("Einstellung wurde geändert, Neustart schlug jedoch fehl: %w", err)
		}
		fmt.Println("Dienst neu gestartet. Bitte warten, bis der Drucker wieder bereit ist.")
	}
	return nil
}

func printResult(r printer.Result) {
	if strings.TrimSpace(r.Firmware) != "" {
		fmt.Println("Firmware:")
		for _, line := range strings.Split(strings.TrimSpace(r.Firmware), "\n") {
			fmt.Println("  " + line)
		}
	}
	fmt.Println("Konfigurationsdatei:", r.Path)
	switch r.Mode {
	case "DETAILED":
		fmt.Println("Automatisches Bed-Leveling: AKTIVIERT")
	case "NEVER":
		fmt.Println("Automatisches Bed-Leveling: DEAKTIVIERT")
	default:
		fmt.Println("Automatisches Bed-Leveling: UNBEKANNT")
	}
	if r.Backup != "" {
		fmt.Println("Sicherung auf dem Drucker:", r.Backup)
	}
	if r.Changed {
		fmt.Println("Änderung erfolgreich und atomar gespeichert.")
		if r.Mode == "NEVER" {
			fmt.Println("WICHTIG: Vor dem nächsten Druck das Druckbett manuell nivellieren.")
		}
	}
}

func usage() {
	fmt.Print(`S5LevelCtl – Active-Leveling-Schalter für UltiMaker S5

Verwendung:
  s5levelctl <Befehl> --host <IP> [Optionen]

Befehle:
  status      Zustand nur lesen
  diagnose    Firmware, Datei und Zustand prüfen (nur lesend)
  disable     automatisches Bed-Leveling deaktivieren
  enable      automatisches Bed-Leveling wieder aktivieren
  backup      Konfigurationsdatei nur sichern
  version     Programmversion anzeigen

Optionen:
  --accept-new-host-key   SSH-Fingerprint beim ersten Zugriff speichern
  --restart               Griffin-Dienst nach enable/disable neu starten
  --port 22               SSH-Port
  --user root             SSH-Benutzer
  --timeout 15s           Verbindungs-Timeout

Das Passwort wird aus S5LEVELCTL_PASSWORD gelesen; ohne Variable wird das
Developer-Mode-Standardpasswort "ultimaker" verwendet.

Beispiel:
  s5levelctl status --host 192.168.1.50 --accept-new-host-key
  s5levelctl disable --host 192.168.1.50 --restart
`)
}
