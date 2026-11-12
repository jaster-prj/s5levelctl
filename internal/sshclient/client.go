package sshclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Options struct {
	Host             string
	Port             int
	User             string
	Password         string
	AcceptNewHostKey bool
}

type Client struct{ client *ssh.Client }

var keyFileMu sync.Mutex

func Dial(ctx context.Context, opts Options) (*Client, error) {
	address := net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))
	callback, err := hostKeyCallback(address, opts.AcceptNewHostKey)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            []ssh.AuthMethod{ssh.Password(opts.Password)},
		HostKeyCallback: callback,
		Timeout:         15 * time.Second,
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("SSH-Verbindung zu %s: %w", address, err)
	}
	sshConn, channels, requests, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH-Anmeldung bei %s: %w", address, err)
	}
	return &Client{client: ssh.NewClient(sshConn, channels, requests)}, nil
}

func (c *Client) Run(command, stdin string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("SSH-Sitzung: %w", err)
	}
	defer session.Close()
	if stdin != "" {
		session.Stdin = strings.NewReader(stdin)
	}
	out, err := session.CombinedOutput(command)
	if err != nil {
		return string(out), fmt.Errorf("Remote-Befehl fehlgeschlagen: %w: %s", err, string(out))
	}
	return string(out), nil
}

func (c *Client) Close() error { return c.client.Close() }

func hostKeyCallback(address string, accept bool) (ssh.HostKeyCallback, error) {
	path, err := knownKeysPath()
	if err != nil {
		return nil, err
	}
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		fingerprint := ssh.FingerprintSHA256(key)
		keyFileMu.Lock()
		defer keyFileMu.Unlock()
		keys, err := loadKeys(path)
		if err != nil {
			return err
		}
		if saved, ok := keys[address]; ok {
			if saved != fingerprint {
				return fmt.Errorf("SSH-Hostkey für %s hat sich geändert (gespeichert %s, empfangen %s)", address, saved, fingerprint)
			}
			return nil
		}
		if !accept {
			return fmt.Errorf("unbekannter SSH-Hostkey für %s: %s; nach manueller Prüfung erneut mit --accept-new-host-key aufrufen", address, fingerprint)
		}
		keys[address] = fingerprint
		return saveKeys(path, keys)
	}, nil
}

func knownKeysPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Konfigurationsverzeichnis: %w", err)
	}
	return filepath.Join(dir, "s5levelctl", "known_hosts.json"), nil
}

func loadKeys(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Hostkey-Datei lesen: %w", err)
	}
	keys := map[string]string{}
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("Hostkey-Datei ist ungültig: %w", err)
	}
	return keys, nil
}

func saveKeys(path string, keys map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("Hostkey-Verzeichnis erstellen: %w", err)
	}
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("Hostkey speichern: %w", err)
	}
	return nil
}
