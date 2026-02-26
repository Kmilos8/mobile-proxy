// Package transparentproxy provides a transparent TCP proxy that intercepts
// iptables-redirected connections and forwards them through an HTTP CONNECT proxy.
// This is used for OpenVPN client traffic: clients connect to the OpenVPN server,
// their TCP traffic gets redirected via iptables REDIRECT, and this proxy
// forwards it through the phone's HTTP proxy using the CONNECT method.

//go:build linux

package transparentproxy

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	connectTimeout   = 15 * time.Second   // dial + CONNECT handshake timeout
	peekTimeout      = 200 * time.Millisecond // time to wait for initial client data (TLS ClientHello arrives in <10ms; 200ms covers slow clients without blocking speed tests)
	idleTimeout      = 120 * time.Second  // 2 min idle before killing connection
	halfCloseTimeout = 5 * time.Second    // after one direction finishes, wait this long for the other
	queueTimeout     = 10 * time.Second   // max time to wait for a connection slot
	maxRetries       = 2
	copyBufSize      = 128 * 1024 // 128KB copy buffer for throughput
	peekSize         = 4096       // enough for TLS ClientHello or HTTP request line + headers
	maxActiveConns   = 8          // max TOTAL concurrent connections per device (dial + data transfer)
)

// proxyTarget holds the HTTP CONNECT proxy endpoint and credentials.
type proxyTarget struct {
	Endpoint string // host:port (e.g. 192.168.255.2:8080)
	Username string
	Password string
	connSem  chan struct{} // limits TOTAL concurrent connections (dial + data) to this device
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
		connSem:  make(chan struct{}, maxActiveConns),
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

	clientAddr := conn.RemoteAddr().(*net.TCPAddr)
	log.Printf("[tproxy] accept: %s -> %s", clientAddr, conn.LocalAddr())

	// Get the original destination via SO_ORIGINAL_DST
	origDst, err := getOriginalDst(conn)
	if err != nil {
		log.Printf("[tproxy] failed to get original dst for %s: %v", clientAddr, err)
		return
	}

	// Get the client's source IP to look up the proxy mapping
	clientIP := clientAddr.IP.String()

	target, ok := p.getMapping(clientIP)
	if !ok {
		log.Printf("[tproxy] no mapping for client %s", clientIP)
		return
	}

	// Wrap client conn in a buffered reader so we can peek at initial data
	// to extract the hostname (SNI for TLS, Host for HTTP).
	// The phone proxy requires domain names in CONNECT, not raw IPs.
	clientReader := bufio.NewReaderSize(conn, copyBufSize)

	// Peek at initial client data with a short deadline
	conn.SetReadDeadline(time.Now().Add(peekTimeout))
	peeked, _ := clientReader.Peek(peekSize)
	conn.SetReadDeadline(time.Time{})

	// Determine the CONNECT address: prefer hostname over raw IP
	connectAddr := origDst
	if len(peeked) > 0 {
		_, port, _ := net.SplitHostPort(origDst)
		if peeked[0] == 0x16 {
			// TLS handshake — extract SNI
			if host := extractSNI(peeked); host != "" {
				connectAddr = net.JoinHostPort(host, port)
			}
		} else {
			// Likely HTTP — extract Host header
			if host := extractHTTPHost(peeked); host != "" {
				connectAddr = net.JoinHostPort(host, port)
			}
		}
	}

	// Acquire connection semaphore — limits TOTAL concurrent connections (dial + data
	// transfer) per device. Without this, a browser opening 30+ connections saturates
	// the cellular UDP tunnel and the device proxy becomes unresponsive.
	select {
	case target.connSem <- struct{}{}:
		// acquired
	case <-time.After(queueTimeout):
		log.Printf("[tproxy] queue full, dropping %s -> %s", clientIP, connectAddr)
		return
	}
	defer func() { <-target.connSem }()

	// Retry CONNECT
	var result *connectResult
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, lastErr = p.dialHTTPConnect(target, connectAddr)
		if lastErr == nil {
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("[tproxy] CONNECT %s attempt %d failed: %v, retrying", connectAddr, attempt+1, lastErr)
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}

	if lastErr != nil {
		log.Printf("[tproxy] CONNECT %s via %s failed after %d attempts: %v", connectAddr, target.Endpoint, maxRetries, lastErr)
		return
	}
	defer result.conn.Close()
	log.Printf("[tproxy] CONNECT %s via %s OK, proxying", connectAddr, target.Endpoint)

	// Set TCP_NODELAY on both sides for lower latency
	setTCPNoDelay(conn)
	setTCPNoDelay(result.conn)

	// Bidirectional copy with proper half-close to prevent FIN_WAIT1 leak.
	done := make(chan struct{}, 2)

	// client → remote (phone proxy): read from clientReader (preserves peeked data)
	go func() {
		buf := make([]byte, copyBufSize)
		io.CopyBuffer(result.conn, clientReader, buf)
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

	// Wait for both directions but with an idle timeout.
	// After one direction finishes, use a short timeout for the other to drain.
	timer := time.NewTimer(idleTimeout)
	defer timer.Stop()
	for i := 0; i < 2; i++ {
		select {
		case <-done:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(halfCloseTimeout)
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

// extractSNI parses a TLS ClientHello and returns the SNI hostname, or "".
func extractSNI(data []byte) string {
	// Minimum TLS record: 5 byte header + 1 byte handshake type
	if len(data) < 6 || data[0] != 0x16 {
		return ""
	}
	// Record length
	recordLen := int(data[3])<<8 | int(data[4])
	if recordLen > len(data)-5 {
		recordLen = len(data) - 5
	}
	hs := data[5 : 5+recordLen]

	// Handshake type must be ClientHello (1)
	if len(hs) < 1 || hs[0] != 0x01 {
		return ""
	}
	// Handshake length (3 bytes)
	if len(hs) < 4 {
		return ""
	}
	hsLen := int(hs[1])<<16 | int(hs[2])<<8 | int(hs[3])
	if hsLen > len(hs)-4 {
		hsLen = len(hs) - 4
	}
	msg := hs[4 : 4+hsLen]

	// Skip client version (2) + random (32) = 34 bytes
	if len(msg) < 34 {
		return ""
	}
	pos := 34

	// Session ID
	if pos >= len(msg) {
		return ""
	}
	sidLen := int(msg[pos])
	pos += 1 + sidLen

	// Cipher suites
	if pos+2 > len(msg) {
		return ""
	}
	csLen := int(msg[pos])<<8 | int(msg[pos+1])
	pos += 2 + csLen

	// Compression methods
	if pos >= len(msg) {
		return ""
	}
	cmLen := int(msg[pos])
	pos += 1 + cmLen

	// Extensions
	if pos+2 > len(msg) {
		return ""
	}
	extLen := int(msg[pos])<<8 | int(msg[pos+1])
	pos += 2
	extEnd := pos + extLen
	if extEnd > len(msg) {
		extEnd = len(msg)
	}

	for pos+4 <= extEnd {
		extType := int(msg[pos])<<8 | int(msg[pos+1])
		extDataLen := int(msg[pos+2])<<8 | int(msg[pos+3])
		pos += 4
		if pos+extDataLen > extEnd {
			break
		}
		if extType == 0x0000 { // SNI extension
			extData := msg[pos : pos+extDataLen]
			return parseSNIExtension(extData)
		}
		pos += extDataLen
	}
	return ""
}

// parseSNIExtension parses the SNI extension data and returns the hostname.
func parseSNIExtension(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	listLen := int(data[0])<<8 | int(data[1])
	if listLen > len(data)-2 {
		listLen = len(data) - 2
	}
	list := data[2 : 2+listLen]

	pos := 0
	for pos+3 <= len(list) {
		nameType := list[pos]
		nameLen := int(list[pos+1])<<8 | int(list[pos+2])
		pos += 3
		if pos+nameLen > len(list) {
			break
		}
		if nameType == 0x00 { // DNS hostname
			return string(list[pos : pos+nameLen])
		}
		pos += nameLen
	}
	return ""
}

// extractHTTPHost extracts the Host header from an HTTP request.
func extractHTTPHost(data []byte) string {
	// Find Host header (case-insensitive)
	lower := bytes.ToLower(data)
	idx := bytes.Index(lower, []byte("\nhost:"))
	if idx < 0 {
		idx = bytes.Index(lower, []byte("\r\nhost:"))
		if idx >= 0 {
			idx += 1 // skip \r
		}
	}
	if idx < 0 {
		return ""
	}
	// Skip "\nhost:" prefix
	rest := data[idx+6:]
	// Find end of line
	eol := bytes.IndexAny(rest, "\r\n")
	if eol < 0 {
		eol = len(rest)
	}
	host := strings.TrimSpace(string(rest[:eol]))
	// Strip port if present (we'll add the right port from origDst)
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
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
