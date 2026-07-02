//go:build !linux && !windows && !darwin

package collect

import "context"

// hardwareIdentity ist der Fallback für nicht gesondert unterstützte Betriebssysteme.
func hardwareIdentity(_ context.Context) (vendor, model, serial string) {
	return "", "", ""
}
