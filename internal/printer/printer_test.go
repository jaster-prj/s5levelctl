package printer

import (
	"errors"
	"strings"
	"testing"
)

type fakeRunner struct {
	out     string
	err     error
	command string
	stdin   string
}

func (f *fakeRunner) Run(command, stdin string) (string, error) {
	f.command, f.stdin = command, stdin
	return f.out, f.err
}

func TestExecuteStatus(t *testing.T) {
	f := &fakeRunner{out: `{"firmware":"9.0.5","path":"/x.py","mode":"DETAILED","changed":false}`}
	got, err := Execute(f, "status")
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != "DETAILED" || f.command != "python3 - status" {
		t.Fatalf("unerwartetes Ergebnis: %#v, command=%q", got, f.command)
	}
	if !strings.Contains(f.stdin, "Sicherheitsabbruch") {
		t.Fatal("Remote-Sicherheitsprüfung fehlt")
	}
}

func TestExecuteRejectsUnknownMode(t *testing.T) {
	f := &fakeRunner{out: `{"firmware":"9.0.5","path":"/x.py","mode":"BROKEN","changed":false}`}
	if _, err := Execute(f, "status"); err == nil {
		t.Fatal("unbekannter Modus wurde akzeptiert")
	}
}

func TestExecuteForwardsRemoteError(t *testing.T) {
	want := errors.New("remote error")
	f := &fakeRunner{err: want}
	if _, err := Execute(f, "disable"); !errors.Is(err, want) {
		t.Fatalf("Fehler = %v, erwartet %v", err, want)
	}
}
