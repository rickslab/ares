package turn

import (
	"log"
	"net"

	"github.com/pion/stun"
)

type StunLogger struct {
	net.PacketConn
}

func (s *StunLogger) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = s.PacketConn.WriteTo(p, addr)
	if err != nil {
		return
	}

	if stun.IsMessage(p) {
		msg := &stun.Message{Raw: p}
		if err = msg.Decode(); err != nil {
			return
		}

		log.Printf("Outbound STUN: %s\n", msg.String())
	}
	return
}

func (s *StunLogger) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = s.PacketConn.ReadFrom(p)
	if err != nil {
		return
	}

	if stun.IsMessage(p) {
		msg := &stun.Message{Raw: p}
		if err = msg.Decode(); err != nil {
			return
		}

		log.Printf("Inbound STUN: %s\n", msg.String())
	}
	return
}
