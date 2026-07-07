package collect

import (
	"context"
	"net"
	"sort"
	"strings"

	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"github.com/thomaspeterson/pc-inventory/internal/shared"
)

// ListenPorts liefert die lauschenden Sockets des Geräts (die lokale Angriffsfläche).
// TCP-Sockets im Zustand LISTEN sowie gebundene UDP-Sockets werden erfasst; Public=true
// markiert Bindungen, die nicht nur auf Loopback lauschen (grundsätzlich vom Netz
// erreichbar – ob wirklich „von außen", entscheidet zusätzlich NAT/Firewall).
func ListenPorts(ctx context.Context) []shared.ListenPort {
	conns, err := psnet.ConnectionsWithContext(ctx, "inet")
	if err != nil || len(conns) == 0 {
		return nil
	}
	names := map[int32]string{} // PID -> Prozessname (einmal auflösen)
	seen := map[string]bool{}   // Dedup proto|addr|port
	var out []shared.ListenPort
	for _, c := range conns {
		if c.Laddr.Port == 0 {
			continue
		}
		// Multicast-/Link-Local-Bindungen sind keine erreichbaren Dienste.
		if isMulticastOrLinkLocal(c.Laddr.IP) {
			continue
		}
		proto := ""
		switch c.Type {
		case 1: // SOCK_STREAM
			if !strings.EqualFold(c.Status, "LISTEN") {
				continue // nur lauschende TCP-Sockets
			}
			proto = "tcp"
		case 2: // SOCK_DGRAM
			// UDP kennt kein LISTEN – ephemere Client-Sockets (Browser/QUIC etc.)
			// sehen wie Dienste aus. Nur Wildcard-Bindungen oder privilegierte Ports
			// (<1024) gelten als Dienst, sonst zu viel Rauschen.
			if !isWildcard(c.Laddr.IP) && c.Laddr.Port >= 1024 {
				continue
			}
			proto = "udp"
		default:
			continue
		}
		if isIPv6Family(c.Family) {
			proto += "6"
		}
		key := proto + "|" + c.Laddr.IP + "|" + itoa(int(c.Laddr.Port))
		if seen[key] {
			continue
		}
		seen[key] = true

		name := ""
		if c.Pid > 0 {
			if n, ok := names[c.Pid]; ok {
				name = n
			} else if p, err := process.NewProcessWithContext(ctx, c.Pid); err == nil {
				name, _ = p.NameWithContext(ctx)
				names[c.Pid] = name
			}
		}
		out = append(out, shared.ListenPort{
			Proto:   proto,
			Address: c.Laddr.IP,
			Port:    int(c.Laddr.Port),
			Process: name,
			PID:     int(c.Pid),
			Public:  !isLoopback(c.Laddr.IP),
		})
	}
	// Stabil sortieren: öffentliche zuerst, dann nach Port.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Public != out[j].Public {
			return out[i].Public
		}
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func isIPv6Family(fam uint32) bool {
	// AF_INET6: 10 (Linux), 23 (Windows), 30 (macOS).
	return fam == 10 || fam == 23 || fam == 30
}

// isLoopback erkennt reine Loopback-Bindungen (nicht vom Netz erreichbar).
func isLoopback(ip string) bool {
	if isWildcard(ip) {
		return false // Wildcard = alle Adressen -> öffentlich
	}
	if p := net.ParseIP(strings.TrimSpace(ip)); p != nil {
		return p.IsLoopback()
	}
	return false
}

// isWildcard ist true für Bindungen an alle Adressen (0.0.0.0, ::, leer, *).
func isWildcard(ip string) bool {
	ip = strings.TrimSpace(ip)
	return ip == "" || ip == "*" || ip == "0.0.0.0" || ip == "::"
}

// isMulticastOrLinkLocal filtert nicht-Dienst-Bindungen (Multicast-Gruppen,
// Link-Local-Adressen wie fe80::).
func isMulticastOrLinkLocal(ip string) bool {
	p := net.ParseIP(strings.TrimSpace(ip))
	if p == nil {
		return false
	}
	return p.IsMulticast() || p.IsLinkLocalUnicast() || p.IsLinkLocalMulticast()
}
