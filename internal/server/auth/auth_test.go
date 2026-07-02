package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("korrekt-pferd-batterie")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	ok, err := VerifyPassword("korrekt-pferd-batterie", hash)
	if err != nil || !ok {
		t.Errorf("korrektes passwort sollte verifizieren (ok=%v err=%v)", ok, err)
	}
	ok, _ = VerifyPassword("falsch", hash)
	if ok {
		t.Error("falsches passwort sollte nicht verifizieren")
	}
}

func TestHashToken(t *testing.T) {
	a := HashToken("abc")
	if a != HashToken("abc") {
		t.Error("hash sollte deterministisch sein")
	}
	if a == HashToken("abd") {
		t.Error("unterschiedliche tokens sollten unterschiedliche hashes haben")
	}
	if len(a) != 64 {
		t.Errorf("sha-256 hex sollte 64 zeichen haben, bekam %d", len(a))
	}
}

func TestTOTPVectorsRFC6238(t *testing.T) {
	// RFC 6238 Testvektor: Secret-Bytes "12345678901234567890" (SHA1).
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	cases := map[uint64]string{
		1:        "287082", // T=59
		37037036: "081804", // T=1111111109
	}
	for counter, want := range cases {
		got, ok := totpAt(secret, counter)
		if !ok || got != want {
			t.Errorf("totpAt(counter=%d) = %q (ok=%v), erwartet %q", counter, got, ok, want)
		}
	}
}

func TestVerifyTOTPAndRecovery(t *testing.T) {
	secret := GenerateTOTPSecret()
	now := uint64(time.Now().Unix() / 30)
	code, _ := totpAt(secret, now)
	if !VerifyTOTP(secret, code) {
		t.Error("aktueller Code sollte gültig sein")
	}
	if VerifyTOTP(secret, "000000") && code != "000000" {
		t.Error("falscher Code sollte abgelehnt werden")
	}
	codes := GenerateRecoveryCodes(10)
	if len(codes) != 10 {
		t.Fatalf("erwartete 10 Backup-Codes, %d", len(codes))
	}
	for _, c := range codes {
		if len(c) != 9 || c[4] != '-' { // xxxx-xxxx
			t.Errorf("unerwartetes Code-Format: %q", c)
		}
	}
}
