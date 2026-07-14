package api

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"
)

// sendWOL verschickt ein Wake-on-LAN-Magic-Packet als Broadcast vom Server aus.
// Best effort: erreicht das/die lokale(n) Netzsegment(e) des Servers – ob das Ziel
// aufwacht, hängt vom Netz (gleiche Broadcast-Domäne) und der Ziel-Hardware ab.
func sendWOL(mac string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) != 6 {
		return fmt.Errorf("ungültige MAC-Adresse %q", mac)
	}
	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xff
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], hw)
	}
	lc := net.ListenConfig{Control: func(_, _ string, c syscall.RawConn) error {
		var serr error
		if cerr := c.Control(func(fd uintptr) { serr = setBroadcast(fd) }); cerr != nil {
			return cerr
		}
		return serr
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pc, err := lc.ListenPacket(ctx, "udp4", ":0")
	if err != nil {
		return err
	}
	defer pc.Close()
	// Limitierte Broadcast-Adresse, Port 9 ("discard") – der klassische WoL-Port.
	_, err = pc.WriteTo(packet, &net.UDPAddr{IP: net.IPv4bcast, Port: 9})
	return err
}
