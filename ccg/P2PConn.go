package ccg

import (
	"net"
)

type P2PConn struct {
	Username string
	conn *net.UDPConn
}

func NewP2PConn(username, addr string) *P2PConn {
	return nil
}
