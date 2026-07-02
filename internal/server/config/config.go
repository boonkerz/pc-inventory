// Package config lädt die Serverkonfiguration aus Defaults, einer optionalen
// YAML-Datei und Umgebungsvariablen (ENV gewinnt vor Datei vor Default).
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config bündelt alle Laufzeiteinstellungen des Servers.
type Config struct {
	// Addr ist die Listen-Adresse, z.B. ":8443".
	Addr string `yaml:"addr"`
	// DatabaseURL: "sqlite:///pfad/inventory.db" oder "postgres://user:pw@host/db".
	DatabaseURL string `yaml:"database_url"`
	// TLSCert/TLSKey: Pfade zum Serverzertifikat. Leer => HTTP (nur für lokale Tests).
	TLSCert string `yaml:"tls_cert"`
	TLSKey  string `yaml:"tls_key"`
	// TLSSelfSigned: fehlt das Zertifikat unter TLSCert/TLSKey, erzeugt der Server beim
	// Start ein selbstsigniertes (für Entwicklung/Docker – der Agent pinnt es als CA).
	TLSSelfSigned bool `yaml:"tls_self_signed"`
	// TLSHosts sind die SANs (Hostnamen/IPs) des selbstsignierten Zertifikats.
	TLSHosts string `yaml:"tls_hosts"`
	// CheckinInterval ist das empfohlene Agent-Polling-Intervall.
	CheckinInterval time.Duration `yaml:"checkin_interval"`
	// OfflineAfter: ab dieser Stille gilt ein Gerät als offline (Default 2.5x Intervall).
	OfflineAfter time.Duration `yaml:"offline_after"`
	// Seed-Admin beim ersten Start (nur wenn noch kein Benutzer existiert).
	SeedAdminUser     string `yaml:"seed_admin_user"`
	SeedAdminPassword string `yaml:"seed_admin_password"`
	// SeedEnrollToken legt beim Start ein festes Enrollment-Token an (für Test-/
	// Automatisierungsumgebungen, z.B. den Docker-Stack). Leer = deaktiviert.
	SeedEnrollToken string `yaml:"seed_enroll_token"`
	// SecureCookie: Cookie nur über HTTPS senden (in Produktion true).
	SecureCookie bool `yaml:"secure_cookie"`
	// Require2FA: alle Benutzer müssen Zwei-Faktor (TOTP) einrichten (Default true).
	Require2FA bool `yaml:"require_2fa"`
	// BehindProxy: der Server läuft hinter einem TLS-terminierenden Reverse-Proxy.
	// Dann bleibt SecureCookie aktiv, obwohl der Server selbst nur HTTP spricht.
	BehindProxy bool `yaml:"behind_proxy"`
	// ResultRetention: Aufbewahrungsdauer der Task-/Befehls-Historie; ältere Läufe
	// werden periodisch gelöscht (0 = nie löschen). Default 30 Tage.
	ResultRetention time.Duration `yaml:"result_retention"`
}

// Default liefert eine sinnvolle Basiskonfiguration.
func Default() Config {
	return Config{
		Addr:              ":8443",
		DatabaseURL:       "sqlite://./inventory.db",
		CheckinInterval:   5 * time.Minute,
		OfflineAfter:      0, // 0 => abgeleitet aus CheckinInterval
		SeedAdminUser:     "admin",
		SeedAdminPassword: "",
		SecureCookie:      true,
		Require2FA:        true,
		ResultRetention:   30 * 24 * time.Hour,
	}
}

// Load liest die Konfiguration. configPath darf leer sein.
func Load(configPath string) (Config, error) {
	cfg := Default()

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, fmt.Errorf("konfig lesen: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("konfig parsen: %w", err)
		}
	}

	applyEnv(&cfg)

	if cfg.OfflineAfter == 0 {
		cfg.OfflineAfter = cfg.CheckinInterval*2 + cfg.CheckinInterval/2
	}
	if cfg.TLSCert == "" && !cfg.BehindProxy && cfg.SecureCookie {
		// Ohne TLS (und ohne TLS-terminierenden Proxy) kann der Browser kein
		// Secure-Cookie senden – sonst käme man nicht mehr rein.
		cfg.SecureCookie = false
	}
	return cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("PCINV_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("PCINV_DB"); v != "" {
		cfg.DatabaseURL = v
	}
	if v := os.Getenv("PCINV_TLS_CERT"); v != "" {
		cfg.TLSCert = v
	}
	if v := os.Getenv("PCINV_TLS_KEY"); v != "" {
		cfg.TLSKey = v
	}
	if v := os.Getenv("PCINV_TLS_SELFSIGNED"); v != "" {
		cfg.TLSSelfSigned = v == "true" || v == "1"
	}
	if v := os.Getenv("PCINV_TLS_HOSTS"); v != "" {
		cfg.TLSHosts = v
	}
	if v := os.Getenv("PCINV_SEED_ADMIN_USER"); v != "" {
		cfg.SeedAdminUser = v
	}
	if v := os.Getenv("PCINV_SEED_ADMIN_PASSWORD"); v != "" {
		cfg.SeedAdminPassword = v
	}
	if v := os.Getenv("PCINV_SEED_ENROLL_TOKEN"); v != "" {
		cfg.SeedEnrollToken = v
	}
	if v := os.Getenv("PCINV_CHECKIN_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CheckinInterval = d
		}
	}
	if v := os.Getenv("PCINV_REQUIRE_2FA"); v != "" {
		cfg.Require2FA = v == "true" || v == "1"
	}
	if v := os.Getenv("PCINV_BEHIND_PROXY"); v != "" {
		cfg.BehindProxy = v == "true" || v == "1"
	}
	if v := os.Getenv("PCINV_SECURE_COOKIE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.SecureCookie = b
		}
	}
	if v := os.Getenv("PCINV_RESULT_RETENTION_DAYS"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d >= 0 {
			cfg.ResultRetention = time.Duration(d) * 24 * time.Hour
		}
	}
}
