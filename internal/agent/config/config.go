// Package config lädt die Agent-Konfiguration und verwaltet den persistenten
// State (die einmalig vergebene Geräte-ID und das Agent-Token).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config steuert das Verhalten des Agents (aus YAML + ENV).
type Config struct {
	// ServerURL z.B. "https://inventory.example.com:8443".
	ServerURL string `yaml:"server_url"`
	// EnrollmentToken wird nur beim ersten Start (Enrollment) benötigt.
	EnrollmentToken string `yaml:"enrollment_token"`
	// CACertPath pinnt optional die CA des Servers (empfohlen bei eigener CA).
	CACertPath string `yaml:"ca_cert_path"`
	// InsecureSkipVerify schaltet die TLS-Prüfung ab – NUR für lokale Tests.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
	// Interval ist das Checkin-Intervall (Server kann ein anderes vorschlagen).
	Interval time.Duration `yaml:"interval"`
	// StatePath ist der Pfad zur State-Datei (Default: neben der Konfig).
	StatePath string `yaml:"state_path"`
	// DisableAutoUpdate schaltet das Selbst-Update ab (Default: aktiviert).
	DisableAutoUpdate bool `yaml:"disable_auto_update"`
	// UpdateCheckInterval steuert, wie oft nach OS-Updates gesucht wird (Default 6h).
	UpdateCheckInterval time.Duration `yaml:"update_check_interval"`
	// DisableUpdateCheck deaktiviert die Suche nach OS-Updates komplett.
	DisableUpdateCheck bool `yaml:"disable_update_check"`
	// DisablePublicIP unterbindet die Ermittlung der öffentlichen IP (externer Aufruf).
	DisablePublicIP bool `yaml:"disable_public_ip"`
	// DisableRemote schaltet die Echtzeit-Funktionen (Wake-Poll, Remote-Terminal) ab.
	DisableRemote bool `yaml:"disable_remote"`
}

// Load liest die Agent-Konfiguration. configPath ist erforderlich.
func Load(configPath string) (Config, error) {
	cfg := Config{Interval: 5 * time.Minute}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("agent-konfig lesen: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("agent-konfig parsen: %w", err)
	}
	if v := os.Getenv("ROSTER_SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("ROSTER_ENROLLMENT_TOKEN"); v != "" {
		cfg.EnrollmentToken = v
	}
	if cfg.ServerURL == "" {
		return cfg, fmt.Errorf("server_url fehlt in der konfiguration")
	}
	if cfg.StatePath == "" {
		cfg.StatePath = filepath.Join(filepath.Dir(configPath), "agent-state.json")
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Minute
	}
	return cfg, nil
}

// State hält die nach dem Enrollment dauerhaft benötigten Werte.
type State struct {
	AgentID    string `json:"agent_id"`
	AgentToken string `json:"agent_token"`
}

// LoadState liest den State; existiert er nicht, wird ein leerer State geliefert.
func LoadState(path string) (State, error) {
	var st State
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	return st, json.Unmarshal(data, &st)
}

// SaveState schreibt den State mit restriktiven Rechten (enthält das Agent-Token).
func SaveState(path string, st State) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Enrolled gibt an, ob der Agent bereits ein Token besitzt.
func (s State) Enrolled() bool { return s.AgentToken != "" }
