//go:build !windows && !linux

package collect

import "context"

// Auf macOS und anderen Systemen sind diese Collectoren aktuell nicht implementiert.

func AVStatusJSON(ctx context.Context) string {
	return jsonOrError(map[string]string{"info": "Virenschutz-Abfrage auf diesem System nicht unterstützt"}, nil)
}

func BitLockerJSON(ctx context.Context) string {
	return jsonOrError(map[string]any{"volumes": []BitLockerVolume{}, "info": "BitLocker nur unter Windows"}, nil)
}

func SmartJSON(ctx context.Context) string {
	return jsonOrError(map[string]any{"disks": []SmartDisk{}, "info": "SMART-Abfrage auf diesem System nicht unterstützt"}, nil)
}

func EventLogJSON(ctx context.Context, logName string, count int) string {
	return jsonOrError(map[string]any{"events": []EventEntry{}, "info": "Log-Abfrage auf diesem System nicht unterstützt"}, nil)
}
