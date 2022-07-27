package dnsserver

import (
	"fmt"
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/netsys-lab/scion-apps/pkg/pan"
	"inet.af/netaddr"
)

// ServerTLS represents an instance of a TLS-over-DNS-server.
type ServerSCION struct {
	*Server
}

// NewServerSCION returns a new CoreDNS SCION server and compiles all plugin in to it.
func NewServerSCION(addr string, group []*Config) (*ServerSCION, error) {
	s, err := NewServer(addr, group)
	if err != nil {
		return nil, err
	}
	return &ServerSCION{s}, nil
}

// Compile-time check to ensure Server implements the caddy.GracefulServer interface
var _ caddy.GracefulServer = &Server{}

// Serve implements caddy.TCPServer interface.
func (s *ServerSCION) Serve(l net.Listener) error {
	return nil
}

// ServePacket implements caddy.UDPServer interface.
func (s *ServerSCION) ServePacket(p net.PacketConn) error {
	// testing
	return s.Server.ServePacket(p)
}

// Listen implements caddy.TCPServer interface.
func (s *ServerSCION) Listen() (net.Listener, error) {
	return nil, nil
}

// ListenPacket implements caddy.UDPServer interface.
func (s *ServerSCION) ListenPacket() (net.PacketConn, error) {
	addr, err := netaddr.ParseIPPort(s.Addr[len(transport.SCION+"://"):])
	if err != nil {
		return nil, err
	}
	p, err := pan.ListenUDP(nil, addr, nil)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information, assuming Quiet is false.
func (s *ServerSCION) OnStartupComplete() {
	if Quiet {
		return
	}

	out := startUpZones(transport.SCION+"://", s.Addr, s.zones)
	if out != "" {
		fmt.Print(out)
	}
}
