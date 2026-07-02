package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const totpPeriod = 30 // Sekunden
const totpDigits = 6

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateTOTPSecret erzeugt ein zufälliges Base32-Secret (160 Bit) für TOTP.
func GenerateTOTPSecret() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return b32.EncodeToString(b)
}

// totpAt berechnet den TOTP-Code (RFC 6238, HMAC-SHA1, 6 Stellen) für einen Zeitschritt.
func totpAt(secret string, counter uint64) (string, bool) {
	key, err := b32.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", false
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	off := sum[len(sum)-1] & 0x0f
	code := (uint32(sum[off]&0x7f) << 24) | (uint32(sum[off+1]) << 16) | (uint32(sum[off+2]) << 8) | uint32(sum[off+3])
	code = code % 1_000_000
	return fmt.Sprintf("%06d", code), true
}

// VerifyTOTP prüft einen Code gegen das Secret mit ±1 Zeitfenster (Uhr-Drift).
func VerifyTOTP(secret, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	counter := uint64(time.Now().Unix() / totpPeriod)
	for _, c := range []uint64{counter - 1, counter, counter + 1} {
		if want, ok := totpAt(secret, c); ok && subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// OTPAuthURL liefert die otpauth://-URI für QR-Codes in Authenticator-Apps.
func OTPAuthURL(issuer, account, secret string) string {
	label := url.PathEscape(issuer + ":" + account)
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", fmt.Sprintf("%d", totpDigits))
	q.Set("period", fmt.Sprintf("%d", totpPeriod))
	return "otpauth://totp/" + label + "?" + q.Encode()
}

// GenerateRecoveryCodes erzeugt n Einmal-Backup-Codes (Format "xxxx-xxxx").
func GenerateRecoveryCodes(n int) []string {
	const alphabet = "abcdefghjkmnpqrstuvwxyz23456789" // ohne verwechselbare Zeichen
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		var sb strings.Builder
		for j, x := range b {
			if j == 4 {
				sb.WriteByte('-')
			}
			sb.WriteByte(alphabet[int(x)%len(alphabet)])
		}
		out = append(out, sb.String())
	}
	return out
}
