// Package transparentproxy provides a transparent TCP proxy that intercepts
// iptables-redirected connections and forwards them through an HTTP CONNECT proxy.
// This is used for OpenVPN client traffic: clients connect to the OpenVPN server,
// their TCP traffic gets redirected via iptables REDIRECT, and this proxy
// forwards it through the phone's HTTP proxy using the CONNECT method.

//go:build linux

package transparentproxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	connectTimeout = 10 * time.Second
	idleTimeout    = 300 * time.Second // 5 min idle before killing connection
	maxRetries     = 2
	copyBufSize    = 128 * 1024 // 128KB copy buffer for throughput
)

// proxyTarget holds the HTTP CONNECT proxy endpoint and credentials.
type proxyTarget struct {
	Endpoint string // host:port (e.g. 192.168.255.2:8080)
	Username string
	Password string
}

// connectResult holds the connection and any buffered reader from the CONNECT handshake.
type connectResult struct {
	conn   net.Conn
	reader *bufio.Reader // may contain buffered data after CONNECT response
}

// Proxy is a transparent HTTP CONNECT forwarder.
type Proxy struct {
	listenAddr string

	mu       sync.RWMutex
	mappings map[string]*proxyTarget // client VPN IP (10.9.0.x) -> proxy target
}

// New creates a new transparent proxy listening on the given address.
func New(listenAddr string) *Proxy {
	return &Proxy{
		listenAddr: listenAddr,
		mappings:   make(map[string]*proxyTarget),
	}
}

// AddMapping registers a client VPN IP to a device proxy endpoint with credentials.
func (p *Proxy) AddMapping(clientIP, proxyEndpoint, username, password string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mappings[clientIP] = &proxyTarget{
		Endpoint: proxyEndpoint,
		Username: username,
		Password: password,
	}
	log.Printf("[tproxy] added mapping: %s -> %s (user=%s)", clientIP, proxyEndpoint, username)
}

// RemoveMapping removes a client VPN IP mapping.
func (p *Proxy) RemoveMapping(clientIP string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.mappings, clientIP)
	log.Printf("[tproxy] removed mapping: %s", clientIP)
}

// getMapping returns the proxy target for a client VPN IP.
func (p *Proxy) getMapping(clientIP string) (*proxyTarget, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	t, ok := p.mappings[clientIP]
	return t, ok
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

	// Get the client's source IP to look up the proxy mapping
	clientAddr := conn.RemoteAddr().(*net.TCPAddr)
	clientIP := clientAddr.IP.String()

	target, ok := p.getMapping(clientIP)
	if !ok {
		log.Printf("[tproxy] no mapping for client %s", clientIP)
		return
	}

	// Retry CONNECT
	var result *connectResult
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, lastErr = p.dialHTTPConnect(target, origDst)
		if lastErr == nil {
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("[tproxy] CONNECT %s attempt %d failed: %v, retrying", origDst, attempt+1, lastErr)
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	if lastErr != nil {
		log.Printf("[tproxy] CONNECT %s via %s failed after %d attempts: %v", origDst, target.Endpoint, maxRetries, lastErr)
		return
	}
	defer result.conn.Close()

	// Set TCP_NODELAY on both sides for lower latency
	setTCPNoDelay(conn)
	setTCPNoDelay(result.conn)

	// Bidirectional copy with proper half-close to prevent FIN_WAIT1 leak.
	done := make(chan struct{}, 2)

	// client → remote (phone proxy): read from client, write to phone
	go func() {
		buf := make([]byte, copyBufSize)
		io.CopyBuffer(result.conn, conn, buf)
		if tc, ok := result.conn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		done <- struct{}{}
	}()

	// remote (phone proxy) → client: read from phone (via bufio.Reader to
	// capture any data buffered during CONNECT handshake), write to client
	go func() {
		buf := make([]byte, copyBufSize)
		io.CopyBuffer(conn, result.reader, buf)
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		done <- struct{}{}
	}()

	// Wait for both directions but with an idle timeout
	timer := time.NewTimer(idleTimeout)
	defer timer.Stop()
	for i := 0; i < 2; i++ {
		select {
		case <-done:
			// Reset timer when one direction finishes — give the other side time to drain
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(30 * time.Second)
		case <-timer.C:
			return
		}
	}
}

func (p *Proxy) dialHTTPConnect(target *proxyTarget, addr string) (*connectResult, error) {
	conn, err := net.DialTimeout("tcp", target.Endpoint, connectTimeout)
	if err != nil {
		return nil, fmt.Errorf("dial proxy: %w", err)
	}

	// Set TCP_NODELAY for faster handshake
	setTCPNoDelay(conn)

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
	if target.Username != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(target.Username + ":" + target.Password))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", creds)
	}
	connectReq += "\r\n"

	conn.SetDeadline(time.Now().Add(connectTimeout))
	_, err = conn.Write([]byte(connectReq))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("write CONNECT: %w", err)
	}

	// Read response using a bufio.Reader. CRITICAL: we must keep this reader
	// because it may have buffered data past the HTTP response headers.
	// Passing the raw conn to io.Copy would lose those bytes.
	br := bufio.NewReaderSize(conn, copyBufSize)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read CONNECT response: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		conn.Close()
		return nil, fmt.Errorf("CONNECT returned %d", resp.StatusCode)
	}

	// Clear deadline for data transfer
	conn.SetDeadline(time.Time{})
	return &connectResult{conn: conn, reader: br}, nil
}

// setTCPNoDelay sets TCP_NODELAY on a connection if it's a TCP connection.
func setTCPNoDelay(conn net.Conn) {
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}
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
