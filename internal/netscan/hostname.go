package netscan

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"time"
)

// resolveHostname bestimmt den Anzeigenamen eines Hosts mit mehreren Quellen, weil
// Reverse-DNS (PTR) in typischen LANs meist nicht konfiguriert ist:
//  1. Reverse-DNS (PTR) über den System-Resolver,
//  2. NetBIOS-Namensabfrage (UDP 137) – liefert Windows-Rechnernamen,
//  3. mDNS/Bonjour (UDP 5353) – liefert Apple-/Drucker-/Avahi-Namen (name.local).
//
// Die erste nicht-leere Antwort gewinnt. Alle Schritte haben ein kurzes Timeout.
func resolveHostname(ip string) string {
	if h := reverseDNS(ip); h != "" {
		return h
	}
	if h := netbiosName(ip); h != "" {
		return h
	}
	if h := mdnsName(ip); h != "" {
		return h
	}
	return ""
}

// --- NetBIOS Node-Status-Abfrage (RFC 1002, NBSTAT) ---

// netbiosName fragt die NetBIOS-Namenstabelle des Hosts ab (UDP 137) und liefert
// den eindeutigen Workstation-Namen (Suffix 0x00, kein Gruppenname).
func netbiosName(ip string) string {
	conn, err := net.DialTimeout("udp", net.JoinHostPort(ip, "137"), 300*time.Millisecond)
	if err != nil {
		return ""
	}
	defer conn.Close()

	// Node-Status-Request auf den Platzhalternamen "*".
	req := []byte{
		0x00, 0x00, // Transaction-ID
		0x00, 0x00, // Flags
		0x00, 0x01, // Questions = 1
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x20, // Namenslänge = 32 (First-Level-Encoding)
		// "*" (0x2A) + 15×0x00, jeweils Nibble+'A' kodiert:
		'C', 'K', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A',
		'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A',
		0x00,       // Namensende
		0x00, 0x21, // Type = NBSTAT
		0x00, 0x01, // Class = IN
	}
	_ = conn.SetDeadline(time.Now().Add(300 * time.Millisecond))
	if _, err := conn.Write(req); err != nil {
		return ""
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n < 57 {
		return ""
	}
	return parseNetbiosNames(buf[:n])
}

// parseNetbiosNames liest die Namenstabelle aus einer NBSTAT-Antwort und liefert
// den eindeutigen Workstation-Namen (Suffix 0x00, Bit „Group" nicht gesetzt).
func parseNetbiosNames(resp []byte) string {
	// Header (12) + Antwort-Name (34: 0x20 + 32 + 0x00) + Type/Class/TTL/RDLENGTH (10)
	// → RDATA beginnt mit NUM_NAMES.
	const off = 12 + 34 + 10
	if len(resp) <= off {
		return ""
	}
	num := int(resp[off])
	p := off + 1
	for i := 0; i < num; i++ {
		if p+18 > len(resp) {
			break
		}
		name := strings.TrimRight(string(resp[p:p+15]), " \x00")
		suffix := resp[p+15]
		flags := binary.BigEndian.Uint16(resp[p+16 : p+18])
		p += 18
		// Suffix 0x00 = Workstation-Dienst; oberstes Flag-Bit = Gruppenname (überspringen).
		if suffix == 0x00 && flags&0x8000 == 0 && name != "" {
			return name
		}
	}
	return ""
}

// --- mDNS Reverse-Lookup (RFC 6762) ---

// mdnsName fragt den Host per Unicast-mDNS (UDP 5300→5353) nach dem PTR-Record
// seiner Reverse-Adresse; Avahi/Apple antworten mit dem .local-Namen.
func mdnsName(ip string) string {
	v4 := net.ParseIP(ip).To4()
	if v4 == nil {
		return "" // IPv6 hier nicht unterstützt
	}
	conn, err := net.DialTimeout("udp", net.JoinHostPort(ip, "5353"), 300*time.Millisecond)
	if err != nil {
		return ""
	}
	defer conn.Close()

	qname := dnsEncodeName([]string{
		strconv.Itoa(int(v4[3])), strconv.Itoa(int(v4[2])),
		strconv.Itoa(int(v4[1])), strconv.Itoa(int(v4[0])),
		"in-addr", "arpa",
	})
	req := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	req = append(req, qname...)
	// Type = PTR (0x000C); Class = IN mit Unicast-Response-Bit (0x8001).
	req = append(req, 0x00, 0x0C, 0x80, 0x01)

	_ = conn.SetDeadline(time.Now().Add(300 * time.Millisecond))
	if _, err := conn.Write(req); err != nil {
		return ""
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}
	name := parsePTRAnswer(buf[:n])
	name = strings.TrimSuffix(name, ".")
	name = strings.TrimSuffix(name, ".local")
	return name
}

// parsePTRAnswer sucht in einer DNS-Antwort das erste PTR-Record und dekodiert
// dessen Ziel-Namen (inkl. Kompressions-Pointer).
func parsePTRAnswer(msg []byte) string {
	if len(msg) < 12 {
		return ""
	}
	qd := int(binary.BigEndian.Uint16(msg[4:6]))
	an := int(binary.BigEndian.Uint16(msg[6:8]))
	p := 12
	for i := 0; i < qd; i++ { // Fragen überspringen
		_, np, ok := dnsReadName(msg, p)
		if !ok {
			return ""
		}
		p = np + 4 // Type + Class
	}
	for i := 0; i < an; i++ {
		_, np, ok := dnsReadName(msg, p)
		if !ok || np+10 > len(msg) {
			return ""
		}
		typ := binary.BigEndian.Uint16(msg[np : np+2])
		rdlen := int(binary.BigEndian.Uint16(msg[np+8 : np+10]))
		rd := np + 10
		if typ == 0x000C { // PTR
			name, _, ok := dnsReadName(msg, rd)
			if ok {
				return name
			}
		}
		p = rd + rdlen
	}
	return ""
}

// dnsEncodeName kodiert Labels als längenpräfixierte DNS-Namen (mit 0x00-Ende).
func dnsEncodeName(labels []string) []byte {
	var b []byte
	for _, l := range labels {
		b = append(b, byte(len(l)))
		b = append(b, l...)
	}
	return append(b, 0x00)
}

// dnsReadName dekodiert einen DNS-Namen ab Offset off (folgt Kompressions-Pointern)
// und liefert Name, Offset direkt hinter dem Namensfeld und ok.
func dnsReadName(msg []byte, off int) (string, int, bool) {
	var labels []string
	pos := off
	end := -1 // Offset hinter dem ersten Pointer (Rückgabewert)
	for hops := 0; hops < 20; hops++ {
		if pos >= len(msg) {
			return "", 0, false
		}
		b := msg[pos]
		if b == 0x00 {
			pos++
			if end < 0 {
				end = pos
			}
			return strings.Join(labels, "."), end, true
		}
		if b&0xC0 == 0xC0 { // Kompressions-Pointer
			if pos+1 >= len(msg) {
				return "", 0, false
			}
			if end < 0 {
				end = pos + 2
			}
			pos = int(binary.BigEndian.Uint16(msg[pos:pos+2]) & 0x3FFF)
			continue
		}
		l := int(b)
		if pos+1+l > len(msg) {
			return "", 0, false
		}
		labels = append(labels, string(msg[pos+1:pos+1+l]))
		pos += 1 + l
	}
	return "", 0, false
}
