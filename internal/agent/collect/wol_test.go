package collect

import "testing"

func TestSendWOL(t *testing.T) {
	if code, msg := SendWOL("zz:zz:zz:zz:zz:zz"); code == 0 {
		t.Fatalf("ungültige MAC sollte fehlschlagen: %s", msg)
	}
	// Gültige MAC: sollte das Paket per Broadcast senden (Exit 0).
	code, msg := SendWOL("AA:BB:CC:DD:EE:FF")
	if code != 0 {
		t.Fatalf("gültige MAC sollte senden, bekam: %s", msg)
	}
	t.Log(msg)
}
