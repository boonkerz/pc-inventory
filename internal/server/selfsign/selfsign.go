// Package selfsign erzeugt selbstsignierte Server-Zertifikate für Entwicklungs-/
// Docker-Umgebungen. Das Zertifikat ist seine eigene CA, sodass ein Agent es als
// vertrauenswürdige CA pinnen kann (ca_cert_path) – echter TLS-Pfad ohne
// insecure_skip_verify, nur eben mit eigenem Cert statt öffentlicher CA.
package selfsign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EnsureCert erzeugt ein selbstsigniertes Zertifikat unter certPath/keyPath, falls
// noch keines existiert. hosts sind die SANs (DNS-Namen und/oder IP-Adressen).
func EnsureCert(certPath, keyPath string, hosts []string) error {
	if fileExists(certPath) && fileExists(keyPath) {
		return nil
	}
	if len(hosts) == 0 {
		hosts = []string{"localhost", "127.0.0.1"}
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("schlüssel erzeugen: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: hosts[0], Organization: []string{"Roster Dev"}},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true, // selbstsigniert = eigene CA -> pinbar
	}
	for _, h := range hosts {
		if ip := net.ParseIP(strings.TrimSpace(h)); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else if h = strings.TrimSpace(h); h != "" {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("zertifikat erzeugen: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		return err
	}
	if err := writePEM(certPath, "CERTIFICATE", der, 0o644); err != nil {
		return err
	}
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	return writePEM(keyPath, "EC PRIVATE KEY", keyDER, 0o600)
}

func writePEM(path, blockType string, der []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: der})
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
