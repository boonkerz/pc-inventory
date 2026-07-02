//go:build !linux && !windows

package collect

import (
	"context"
	"errors"
)

// ListServices wird auf dieser Plattform nicht unterstützt.
func ListServices(ctx context.Context) string {
	return servicesJSON(nil, errors.New("Dienst-Verwaltung auf diesem System nicht unterstützt"))
}

// ControlService wird auf dieser Plattform nicht unterstützt.
func ControlService(ctx context.Context, name, action string) (int, string) {
	return 1, "nicht unterstützt auf diesem System"
}
