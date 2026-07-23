// Package proxmox ist ein minimaler Client für die Proxmox-VE-API (nur die von
// Roster benötigten Aufrufe: Gäste auflisten, Gast rebooten). Authentifizierung per
// API-Token; kein Fremd-SDK.
package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/boonkerz/roster/internal/server/model"
)

// Client spricht mit genau einem Proxmox-Host.
type Client struct {
	baseURL string
	tokenID string
	secret  string
	http    *http.Client
}

// New baut einen Client aus einem gespeicherten Host.
func New(h *model.ProxmoxHost) *Client {
	base := strings.TrimRight(strings.TrimSpace(h.BaseURL), "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return &Client{
		baseURL: base,
		tokenID: strings.TrimSpace(h.TokenID),
		secret:  strings.TrimSpace(h.TokenSecret),
		http: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !h.VerifyTLS},
			},
		},
	}
}

// Guest ist ein Proxmox-Gast (Container oder VM).
type Guest struct {
	Node   string `json:"node"`
	VMID   int    `json:"vmid"`
	Type   string `json:"type"`   // lxc | qemu
	Name   string `json:"name"`   // Anzeigename
	Status string `json:"status"` // running | stopped …
}

func (c *Client) auth(req *http.Request) {
	// Proxmox-API-Token-Header: "PVEAPIToken=USER@REALM!TOKENID=SECRET".
	req.Header.Set("Authorization", "PVEAPIToken="+c.tokenID+"="+c.secret)
}

// ListGuests liefert alle Gäste (LXC + QEMU) clusterweit inkl. Node und Status.
func (c *Client) ListGuests(ctx context.Context) ([]Guest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api2/json/cluster/resources?type=vm", nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxmox %d: %s", resp.StatusCode, snippet(body))
	}
	var out struct {
		Data []struct {
			Node   string `json:"node"`
			VMID   int    `json:"vmid"`
			Type   string `json:"type"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("proxmox antwort ungültig: %w", err)
	}
	guests := make([]Guest, 0, len(out.Data))
	for _, d := range out.Data {
		if d.Type != "lxc" && d.Type != "qemu" {
			continue
		}
		guests = append(guests, Guest{Node: d.Node, VMID: d.VMID, Type: d.Type, Name: d.Name, Status: d.Status})
	}
	return guests, nil
}

// Reboot rebootet einen Gast (LXC/QEMU) auf dem angegebenen Node.
func (c *Client) Reboot(ctx context.Context, node, gtype string, vmid int) error {
	if gtype != "lxc" && gtype != "qemu" {
		return fmt.Errorf("ungültiger Gast-Typ %q", gtype)
	}
	if node == "" || vmid <= 0 {
		return fmt.Errorf("node und vmid erforderlich")
	}
	url := fmt.Sprintf("%s/api2/json/nodes/%s/%s/%d/status/reboot", c.baseURL, node, gtype, vmid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("proxmox reboot %d: %s", resp.StatusCode, snippet(body))
	}
	return nil
}

func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}
