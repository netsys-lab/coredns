package dnsserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/transport"
	"github.com/lucas-clemente/quic-go"
	"github.com/miekg/dns"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsec-ethz/scion-apps/pkg/quicutil"
	"inet.af/netaddr"
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

	err := s.dnsserver.ActivateAndServe()
	if err != nil {
		panic(err)
	}
	return err
}

// ServePacket implements caddy.UDPServer interface.
func (s *ServerSCION) ServePacket(p net.PacketConn) error {
	/*s.dnsserver = &dns.Server{
		PacketConn: p,
		Net:        "udp",
		ReusePort:  false,
		//ReadTimeout: time.Minute,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			ctx := context.WithValue(context.Background(), Key{}, s)
			ctx = context.WithValue(ctx, LoopKey{}, 0)
			s.ServeDNS(ctx, w, r)
		})}

	err := s.dnsserver.ActivateAndServe()
	if err != nil {
		panic(err)
	}*/
	return nil
}

type quicConn struct {
	conn   quic.Session
	stream quic.Stream
}

func (q *quicConn) Read(b []byte) (int, error) {
	return q.stream.Read(b)
}

func (q *quicConn) Write(b []byte) (int, error) {
	return q.stream.Write(b)
}

func (q *quicConn) Close() error {
	return q.stream.Close()
}

func (q *quicConn) LocalAddr() net.Addr {
	return q.conn.LocalAddr()
}

func (q *quicConn) RemoteAddr() net.Addr {
	return q.conn.RemoteAddr()
}

func (q *quicConn) SetDeadline(t time.Time) error {
	return q.stream.SetDeadline(t)
}

func (q *quicConn) SetReadDeadline(t time.Time) error {
	return q.stream.SetReadDeadline(t)

}

func (q *quicConn) SetWriteDeadline(t time.Time) error {
	return q.stream.SetWriteDeadline(t)
}

type quicListener struct {
	quic.Listener
}

func (q *quicListener) Accept() (net.Conn, error) {
	ctx := context.Background()
	conn, err := q.Listener.Accept(ctx)
	if err != nil {
		return nil, err
	}
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	qconn := quicConn{conn: conn, stream: stream}

	return &qconn, nil
}

// Listen implements caddy.TCPServer interface.
func (s *ServerSCION) Listen() (net.Listener, error) {
	tlsCfg := &tls.Config{
		Certificates: quicutil.MustGenerateSelfSignedCert(),
		NextProtos:   []string{"hello-quic"},
	}
	listen := netaddr.IPPortFrom(netaddr.IPFrom4([4]byte{127, 0, 0, 1}), 10001)
	session, err := pan.ListenQUIC(context.Background(), listen, nil, tlsCfg, nil)
	if err != nil {
		return nil, err
	}
	return &quicListener{session}, nil
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
