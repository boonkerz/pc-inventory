package alert

import (
	"context"
	"sort"
)

// Notification ist eine zu versendende Benachrichtigung.
type Notification struct {
	Subject string
	Body    string
}

// Field beschreibt ein Konfigurationsfeld eines Providers, damit die UI das
// Formular generisch rendern kann.
type Field struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // text | password | number | checkbox
	Required bool   `json:"required,omitempty"`
	Help     string `json:"help,omitempty"`
}

// Provider ist ein Benachrichtigungs-Integrationstyp (E-Mail, Webhook, …).
type Provider interface {
	Type() string
	Label() string
	Fields() []Field
	Send(ctx context.Context, cfg map[string]string, n Notification) error
}

// ProviderInfo ist die UI-Beschreibung eines Providers (Katalog).
type ProviderInfo struct {
	Type   string  `json:"type"`
	Label  string  `json:"label"`
	Fields []Field `json:"fields"`
}

// registry hält alle eingebauten Provider, je Typ.
var registry = func() map[string]Provider {
	m := map[string]Provider{}
	for _, p := range []Provider{
		emailProvider{}, webhookProvider{}, pushoverProvider{}, telegramProvider{}, ntfyProvider{},
	} {
		m[p.Type()] = p
	}
	return m
}()

// ProviderByType liefert den Provider für einen Typ (nil, falls unbekannt).
func ProviderByType(t string) Provider { return registry[t] }

// Catalog liefert die Provider-Beschreibungen für die UI (stabil sortiert).
func Catalog() []ProviderInfo {
	out := make([]ProviderInfo, 0, len(registry))
	for _, p := range registry {
		out = append(out, ProviderInfo{Type: p.Type(), Label: p.Label(), Fields: p.Fields()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

// SecretKeys liefert die geheimen (zu maskierenden) Feld-Keys eines Provider-Typs.
func SecretKeys(t string) map[string]bool {
	p := registry[t]
	if p == nil {
		return nil
	}
	out := map[string]bool{}
	for _, f := range p.Fields() {
		if f.Type == "password" {
			out[f.Key] = true
		}
	}
	return out
}
