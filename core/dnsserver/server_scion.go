package dnsserver

import (
	"context"
	"fmt"
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/transport"

	"github.com/miekg/dns"
	"github.com/netsys-lab/sqnet"
)

// ServerTLS represents an instance of a TLS-over-DNS-server.
type ServerSCION struct {
	*Server
	dnsserver *dns.Server
}

// NewServerSCION returns a new CoreDNS SCION server and compiles all plugin in to it.
func NewServerSCION(addr string, group []*Config) (*ServerSCION, error) {
	s, err := NewServer(addr, group)
	if err != nil {
		return nil, err
	}
	return &ServerSCION{s, nil}, nil
}

// Compile-time check to ensure Server implements the caddy.GracefulServer interface
var _ caddy.GracefulServer = &Server{}

// Serve implements caddy.TCPServer interface.
func (s *ServerSCION) Serve(l net.Listener) error {
	s.dnsserver = &dns.Server{
		Listener:  l,
		Net:       "tcp",
		ReusePort: false,
		//ReadTimeout: time.Minute,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			ctx := context.WithValue(context.Background(), Key{}, s)
			ctx = context.WithValue(ctx, LoopKey{}, 0)
			s.ServeDNS(ctx, w, r)
		})}

	return s.dnsserver.ActivateAndServe()
}

// ServePacket implements caddy.UDPServer interface.
func (s *ServerSCION) ServePacket(p net.PacketConn) error {
	log.Debug("Called ServePacket, which does nothing")
	return nil
}

// Listen implements caddy.TCPServer interface.
func (s *ServerSCION) Listen() (net.Listener, error) {
	l, err := sqnet.ListenString("0.0.0.0" + s.Server.Addr[len(transport.SCION+"://"):])
	if err != nil {
		log.Error(err)
	}
	return l, err

}

// ListenPacket implements caddy.UDPServer interface.
func (s *ServerSCION) ListenPacket() (net.PacketConn, error) {
	//log.Printf("listenaddr %s", s.Addr)
	/*addr, err := netaddr.ParseIPPort(s.Addr[len(transport.SCION+"://"):])
	if err != nil {
		addr = netaddr.IPPort{IP: netaddr.IPFrom4([4]byte{})}
		return nil, err
	}*/
	/*addr := netaddr.IPPortFrom(netaddr.IPFrom4([4]byte{127, 0, 0, 1}), 10001)
	p, err := pan.ListenUDP(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("returned: %v", p.LocalAddr())
	//buf := make([]byte, 1024)
	//x, y, z := p.ReadFrom(buf)
	//log.Printf("returned: %v, %v, %v, %v", p.LocalAddr(), x, y, z)

	return p, nil*/
	return nil, nil

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
