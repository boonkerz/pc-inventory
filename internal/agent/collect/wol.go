package collect

import (
	"encoding/hex"
	"net"
	"strings"
)

// SendWOL sendet ein Wake-on-LAN Magic Packet (Broadcast, UDP/9) an die MAC.
func SendWOL(mac string) (int, string) {
	clean := strings.NewReplacer(":", "", "-", "", ".", "").Replace(strings.TrimSpace(mac))
	hw, err := hex.DecodeString(clean)
	if err != nil || len(hw) != 6 {
		return 1, "ungültige MAC-Adresse: " + mac
	}
	packet := make([]byte, 6, 102)
	for i := range packet {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		packet = append(packet, hw...)
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return 1, err.Error()
	}
	defer conn.Close()
	if rc, e := conn.SyscallConn(); e == nil {
		_ = rc.Control(func(fd uintptr) { _ = setBroadcast(fd) })
	}
	if _, err := conn.WriteToUDP(packet, &net.UDPAddr{IP: net.IPv4bcast, Port: 9}); err != nil {
		return 1, err.Error()
	}
	return 0, "Magic Packet gesendet an " + mac
}
