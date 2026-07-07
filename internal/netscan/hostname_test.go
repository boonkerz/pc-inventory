package netscan

import "testing"

// nbName kodiert einen 15-Zeichen-NetBIOS-Namen (rechts mit Leerzeichen aufgefüllt).
func nbName(s string) []byte {
	b := make([]byte, 15)
	for i := range b {
		if i < len(s) {
			b[i] = s[i]
		} else {
			b[i] = ' '
		}
	}
	return b
}

func TestParseNetbiosNames(t *testing.T) {
	resp := make([]byte, 12+34) // Header + Antwort-Name
	// Header: an=1
	resp[7] = 1
	// Type/Class/TTL/RDLENGTH (10 Bytes)
	resp = append(resp, 0x00, 0x21, 0x00, 0x01, 0, 0, 0, 0, 0, 0)
	// RDATA
	resp = append(resp, 2) // NUM_NAMES = 2
	// Eintrag 1: Gruppenname (soll übersprungen werden), Suffix 0x00 aber Group-Flag
	e1 := nbName("WORKGROUP")
	e1 = append(e1, 0x00)       // Suffix
	e1 = append(e1, 0x80, 0x00) // Flags: Group-Bit gesetzt
	resp = append(resp, e1...)
	// Eintrag 2: eindeutiger Workstation-Name
	e2 := nbName("MYPC")
	e2 = append(e2, 0x00)       // Suffix Workstation
	e2 = append(e2, 0x04, 0x00) // Flags: unique
	resp = append(resp, e2...)

	if got := parseNetbiosNames(resp); got != "MYPC" {
		t.Fatalf("erwartet MYPC, bekam %q", got)
	}
}

func TestParseNetbiosNamesEmpty(t *testing.T) {
	if got := parseNetbiosNames([]byte{0, 0}); got != "" {
		t.Fatalf("erwartete leeren String, bekam %q", got)
	}
}

func TestParsePTRAnswer(t *testing.T) {
	// Header: qd=1, an=1
	msg := []byte{0, 0, 0, 0x84, 0, 1, 0, 1, 0, 0, 0, 0}
	// Frage: qname "1.0.in-addr.arpa" (Inhalt egal, nur Länge zählt)
	q := dnsEncodeName([]string{"1", "0", "in-addr", "arpa"})
	msg = append(msg, q...)
	msg = append(msg, 0x00, 0x0C, 0x00, 0x01) // Type PTR, Class IN
	answerStart := len(msg)
	_ = answerStart
	// Antwort: Name-Pointer auf die Frage (0xC00C), Type PTR, Class, TTL, RDLENGTH, RDATA
	rdata := dnsEncodeName([]string{"myhost", "local"})
	msg = append(msg, 0xC0, 0x0C)             // Kompressions-Pointer auf Offset 12
	msg = append(msg, 0x00, 0x0C, 0x00, 0x01) // Type PTR, Class IN
	msg = append(msg, 0, 0, 0, 60)            // TTL
	msg = append(msg, byte(len(rdata)>>8), byte(len(rdata)))
	msg = append(msg, rdata...)

	if got := parsePTRAnswer(msg); got != "myhost.local" {
		t.Fatalf("erwartet myhost.local, bekam %q", got)
	}
}

func TestDNSReadNameCompression(t *testing.T) {
	// "local" ab Offset 12, davor Füllbytes; danach "myhost" + Pointer auf 12.
	msg := make([]byte, 12)
	msg = append(msg, dnsEncodeName([]string{"local"})...) // Offset 12
	target := len(msg)
	msg = append(msg, 6, 'm', 'y', 'h', 'o', 's', 't', 0xC0, 0x0C)

	name, next, ok := dnsReadName(msg, target)
	if !ok || name != "myhost.local" {
		t.Fatalf("erwartet myhost.local, bekam %q (ok=%v)", name, ok)
	}
	if next != target+9 { // 1+6 Bytes Label + 2 Bytes Pointer
		t.Fatalf("erwartet next=%d, bekam %d", target+9, next)
	}
}
