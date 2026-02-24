// Package transparentproxy provides a transparent TCP proxy that intercepts
// iptables-redirected connections and forwards them through a SOCKS5 proxy.
// This is used for OpenVPN client traffic: clients connect to the OpenVPN server,
// their TCP traffic gets redirected via iptables REDIRECT, and this proxy
// forwards it through the phone's SOCKS5 proxy.

//go:build linux

package transparentproxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/net/proxy"
)

// Proxy is a transparent SOCKS5 forwarder.
// It accepts TCP connections redirected by iptables and forwards them
// through the appropriate device's SOCKS5 proxy based on the client's VPN IP.
type Proxy struct {
	listenAddr string

	mu       sync.RWMutex
	mappings map[string]string // client VPN IP (10.9.0.x) -> device SOCKS5 endpoint (192.168.255.y:1080)
}

// New creates a new transparent proxy listening on the given address.
func New(listenAddr string) *Proxy {
	return &Proxy{
		listenAddr: listenAddr,
		mappings:   make(map[string]string),
	}
}

// AddMapping registers a client VPN IP to a device SOCKS5 endpoint.
func (p *Proxy) AddMapping(clientIP, socksEndpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mappings[clientIP] = socksEndpoint
	log.Printf("[tproxy] added mapping: %s -> %s", clientIP, socksEndpoint)
}

// RemoveMapping removes a client VPN IP mapping.
func (p *Proxy) RemoveMapping(clientIP string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.mappings, clientIP)
	log.Printf("[tproxy] removed mapping: %s", clientIP)
}

// getMapping returns the SOCKS5 endpoint for a client VPN IP.
func (p *Proxy) getMapping(clientIP string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ep, ok := p.mappings[clientIP]
	return ep, ok
}

// Start begins listening for redirected TCP connections.
func (p *Proxy) Start() error {
	ln, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("transparent proxy listen: %w", err)
	}
	log.Printf("[tproxy] listening on %s", p.listenAddr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("[tproxy] accept error: %v", err)
				continue
			}
			go p.handleConn(conn)
		}
	}()

	return nil
}

func (p *Proxy) handleConn(conn net.Conn) {
	defer conn.Close()

	// Get the original destination via SO_ORIGINAL_DST
	origDst, err := getOriginalDst(conn)
	if err != nil {
		log.Printf("[tproxy] failed to get original dst: %v", err)
		return
	}

	// Get the client's source IP to look up the SOCKS mapping
	clientAddr := conn.RemoteAddr().(*net.TCPAddr)
	clientIP := clientAddr.IP.String()

	socksEndpoint, ok := p.getMapping(clientIP)
	if !ok {
		log.Printf("[tproxy] no mapping for client %s", clientIP)
		return
	}

	// Connect through the device's SOCKS5 proxy
	dialer, err := proxy.SOCKS5("tcp", socksEndpoint, nil, proxy.Direct)
	if err != nil {
		log.Printf("[tproxy] failed to create SOCKS5 dialer for %s: %v", socksEndpoint, err)
		return
	}

	remote, err := dialer.Dial("tcp", origDst)
	if err != nil {
		log.Printf("[tproxy] SOCKS5 dial %s via %s failed: %v", origDst, socksEndpoint, err)
		return
	}
	defer remote.Close()

	// Bidirectional copy
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(remote, conn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(conn, remote)
		done <- struct{}{}
	}()
	<-done
}

// getOriginalDst retrieves the original destination address from a connection
// that was redirected via iptables REDIRECT target.
func getOriginalDst(conn net.Conn) (string, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return "", fmt.Errorf("not a TCP connection")
	}

	raw, err := tcpConn.SyscallConn()
	if err != nil {
		return "", fmt.Errorf("syscall conn: %w", err)
	}

	var addr syscall.RawSockaddrInet4
	var getErr error

	err = raw.Control(func(fd uintptr) {
		const SO_ORIGINAL_DST = 80
		addrLen := uint32(unsafe.Sizeof(addr))
		_, _, errno := syscall.Syscall6(
			syscall.SYS_GETSOCKOPT,
			fd,
			syscall.SOL_IP,
			SO_ORIGINAL_DST,
			uintptr(unsafe.Pointer(&addr)),
			uintptr(unsafe.Pointer(&addrLen)),
			0,
		)
		if errno != 0 {
			getErr = fmt.Errorf("getsockopt SO_ORIGINAL_DST: %v", errno)
		}
	})
	if err != nil {
		return "", err
	}
	if getErr != nil {
		return "", getErr
	}

	ip := net.IPv4(addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3])
	port := int(addr.Port>>8) | int(addr.Port&0xff)<<8 // network byte order
	return fmt.Sprintf("%s:%d", ip, port), nil
}
