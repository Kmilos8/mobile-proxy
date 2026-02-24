package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mobileproxy/server/internal/transparentproxy"
	"github.com/songgao/water"
)

// Packet type prefixes
const (
	TypeAuth     = 0x01
	TypeData     = 0x02
	TypePing     = 0x03
	TypeAuthOK   = 0x01
	TypeAuthFail = 0x03
	TypePong     = 0x04
	TypeCommand  = 0x05 // Server→device command push
)

const (
	defaultPort      = 443
	defaultAPIURL    = "http://127.0.0.1:8080"
	tunName          = "tun0"
	tunIP            = "192.168.255.1"
	tunSubnet        = "192.168.255.0/24"
	tunMTU           = 1400
	maxPacketSize    = tunMTU + 100 // headroom for encapsulation
	keepaliveTimeout = 60 * time.Second
	cleanupInterval  = 10 * time.Second
	ipPoolStart      = 2
	ipPoolEnd        = 254
	deviceIDLen      = 16
	udpRecvBufSize   = 4 * 1024 * 1024 // 4MB UDP socket buffer
	udpSendBufSize   = 4 * 1024 * 1024
)

type client struct {
	udpAddr  *net.UDPAddr
	deviceID string
	vpnIP    net.IP
	lastSeen atomic.Int64 // unix timestamp — lock-free updates
}

func (c *client) touch() {
	c.lastSeen.Store(time.Now().Unix())
}

func (c *client) idleSince(now time.Time) time.Duration {
	return now.Sub(time.Unix(c.lastSeen.Load(), 0))
}

type tunnelServer struct {
	udpConn  *net.UDPConn
	tunIface *water.Interface
	apiURL   string

	mu      sync.RWMutex
	clients map[string]*client // vpnIP string -> client
	addrMap map[string]string  // udpAddr string -> vpnIP string

	// Device ID → client lookup for command push
	deviceMap   map[string]*client // deviceID string -> client
	deviceMapMu sync.RWMutex

	ipPool   []bool // true = in use, index 0 = .2
	ipPoolMu sync.Mutex

	// Pre-allocated send buffer for tunToUdp (single goroutine, no lock needed)
	tunBuf []byte

	// Transparent proxy for OpenVPN client traffic
	transparentProxy *transparentproxy.Proxy
}

func main() {
	port := defaultPort
	if v := os.Getenv("TUNNEL_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}
	apiURL := defaultAPIURL
	if v := os.Getenv("API_URL"); v != "" {
		apiURL = v
	}

	// Create TUN interface
	config := water.Config{DeviceType: water.TUN}
	config.Name = tunName

	iface, err := water.New(config)
	if err != nil {
		log.Fatalf("Failed to create TUN interface: %v", err)
	}
	log.Printf("Created TUN interface: %s", iface.Name())

	// Configure TUN interface
	configureTUN(iface.Name())

	// Listen on UDP
	addr := &net.UDPAddr{Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port %d: %v", port, err)
	}

	// Increase UDP socket buffers for throughput
	conn.SetReadBuffer(udpRecvBufSize)
	conn.SetWriteBuffer(udpSendBufSize)

	log.Printf("Listening on UDP port %d (buffers: recv=%dKB, send=%dKB)",
		port, udpRecvBufSize/1024, udpSendBufSize/1024)

	// Start transparent proxy for OpenVPN client traffic
	tproxy := transparentproxy.New("0.0.0.0:12345")
	if err := tproxy.Start(); err != nil {
		log.Printf("Warning: transparent proxy failed to start: %v", err)
	}

	srv := &tunnelServer{
		udpConn:          conn,
		tunIface:         iface,
		apiURL:           apiURL,
		clients:          make(map[string]*client),
		addrMap:          make(map[string]string),
		deviceMap:        make(map[string]*client),
		ipPool:           make([]bool, ipPoolEnd-ipPoolStart+1),
		tunBuf:           make([]byte, maxPacketSize+1), // +1 for type prefix
		transparentProxy: tproxy,
	}

	// Start goroutines
	go srv.udpToTun()
	go srv.tunToUdp()
	go srv.cleanupLoop()
	go srv.tcpAuthListener(port)
	go srv.startPushAPI()

	// Block forever
	select {}
}

func runCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func configureTUN(name string) {
	// Assign IP address
	if out, err := runCmd("ip", "addr", "add", tunIP+"/24", "dev", name); err != nil {
		log.Printf("Warning: ip addr add: %s: %v", string(out), err)
	}

	// Set MTU and bring interface up
	if out, err := runCmd("ip", "link", "set", "dev", name, "mtu", strconv.Itoa(tunMTU), "up"); err != nil {
		log.Printf("Warning: ip link set: %s: %v", string(out), err)
	}

	// Increase TUN queue length for throughput
	if out, err := runCmd("ip", "link", "set", "dev", name, "txqueuelen", "1000"); err != nil {
		log.Printf("Warning: txqueuelen: %s: %v", string(out), err)
	}

	// Enable IP forwarding
	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// Add MASQUERADE rule for VPN subnet (check first, add if missing)
	if _, err := runCmd("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", tunSubnet, "-j", "MASQUERADE"); err != nil {
		if out, err := runCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", tunSubnet, "-j", "MASQUERADE"); err != nil {
			log.Printf("Warning: MASQUERADE rule failed: %s: %v", string(out), err)
		} else {
			log.Printf("Added MASQUERADE rule for %s", tunSubnet)
		}
	}

	// Add FORWARD rules for VPN subnet
	if _, err := runCmd("iptables", "-C", "FORWARD", "-s", tunSubnet, "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-A", "FORWARD", "-s", tunSubnet, "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: FORWARD rule failed: %s: %v", string(out), err)
		}
	}
	if _, err := runCmd("iptables", "-C", "FORWARD", "-d", tunSubnet, "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-A", "FORWARD", "-d", tunSubnet, "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: FORWARD rule failed: %s: %v", string(out), err)
		}
	}

	// OpenVPN client subnet (10.9.0.0/24) — FORWARD + MASQUERADE rules
	ovpnSubnet := "10.9.0.0/24"
	if _, err := runCmd("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", ovpnSubnet, "-j", "MASQUERADE"); err != nil {
		if out, err := runCmd("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ovpnSubnet, "-j", "MASQUERADE"); err != nil {
			log.Printf("Warning: MASQUERADE rule for %s failed: %s: %v", ovpnSubnet, string(out), err)
		} else {
			log.Printf("Added MASQUERADE rule for %s", ovpnSubnet)
		}
	}
	// Insert before UFW chains (which have policy DROP) so OpenVPN traffic is accepted
	if _, err := runCmd("iptables", "-C", "FORWARD", "-s", ovpnSubnet, "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-I", "FORWARD", "2", "-s", ovpnSubnet, "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: FORWARD rule for %s failed: %s: %v", ovpnSubnet, string(out), err)
		}
	}
	if _, err := runCmd("iptables", "-C", "FORWARD", "-d", ovpnSubnet, "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-I", "FORWARD", "3", "-d", ovpnSubnet, "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: FORWARD rule for %s failed: %s: %v", ovpnSubnet, string(out), err)
		}
	}
	// Allow redirected OpenVPN client TCP traffic to reach the transparent proxy
	if _, err := runCmd("iptables", "-C", "INPUT", "-s", ovpnSubnet, "-p", "tcp", "--dport", "12345", "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-I", "INPUT", "1", "-s", ovpnSubnet, "-p", "tcp", "--dport", "12345", "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: INPUT rule for tproxy failed: %s: %v", string(out), err)
		}
	}
	// Reject QUIC (UDP 443) from OpenVPN clients to force TCP/HTTPS through the transparent proxy.
	// REJECT (not DROP) sends ICMP unreachable so browsers fall back to TCP immediately.
	if _, err := runCmd("iptables", "-C", "FORWARD", "-s", ovpnSubnet, "-p", "udp", "--dport", "443", "-j", "REJECT", "--reject-with", "icmp-port-unreachable"); err != nil {
		// Remove any old DROP rule first
		runCmd("iptables", "-D", "FORWARD", "-s", ovpnSubnet, "-p", "udp", "--dport", "443", "-j", "DROP")
		if out, err := runCmd("iptables", "-I", "FORWARD", "1", "-s", ovpnSubnet, "-p", "udp", "--dport", "443", "-j", "REJECT", "--reject-with", "icmp-port-unreachable"); err != nil {
			log.Printf("Warning: QUIC reject rule failed: %s: %v", string(out), err)
		} else {
			log.Printf("QUIC rejected for OpenVPN clients (%s)", ovpnSubnet)
		}
	}

	log.Printf("TUN interface %s configured: %s/24, MTU %d", name, tunIP, tunMTU)
}

func (s *tunnelServer) allocateIP() (net.IP, error) {
	s.ipPoolMu.Lock()
	defer s.ipPoolMu.Unlock()

	for i := 0; i < len(s.ipPool); i++ {
		if !s.ipPool[i] {
			s.ipPool[i] = true
			ip := net.IPv4(192, 168, 255, byte(i+ipPoolStart))
			return ip, nil
		}
	}
	return nil, fmt.Errorf("IP pool exhausted")
}

func (s *tunnelServer) releaseIP(ip net.IP) {
	s.ipPoolMu.Lock()
	defer s.ipPoolMu.Unlock()

	octet := ip.To4()
	if octet == nil {
		return
	}
	idx := int(octet[3]) - ipPoolStart
	if idx >= 0 && idx < len(s.ipPool) {
		s.ipPool[idx] = false
	}
}

func (s *tunnelServer) udpToTun() {
	buf := make([]byte, maxPacketSize+1)
	for {
		n, remoteAddr, err := s.udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP read error: %v", err)
			continue
		}
		if n < 1 {
			continue
		}

		pktType := buf[0]

		switch pktType {
		case TypeData:
			// Hot path — inline for performance
			if n < 21 { // 1 type + 20 min IP header
				continue
			}
			s.mu.RLock()
			ipStr, ok := s.addrMap[remoteAddr.String()]
			if ok {
				if c, exists := s.clients[ipStr]; exists {
					c.touch()
				}
			}
			s.mu.RUnlock()
			if ok {
				s.tunIface.Write(buf[1:n])
			}
		case TypeAuth:
			s.handleAuth(buf[1:n], remoteAddr)
		case TypePing:
			s.handlePing(remoteAddr)
		default:
			// Silently drop — don't log every scanner packet
		}
	}
}

func (s *tunnelServer) handleAuth(data []byte, addr *net.UDPAddr) {
	if len(data) < deviceIDLen {
		log.Printf("AUTH packet too short from %s", addr)
		s.sendAuthFail(addr)
		return
	}

	deviceID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		data[0:4], data[4:6], data[6:8], data[8:10], data[10:16])

	log.Printf("AUTH request from %s, device_id=%s", addr, deviceID)

	// Check if this device is already connected — reuse session silently
	s.mu.Lock()
	for ipStr, c := range s.clients {
		if c.deviceID == deviceID {
			log.Printf("Device %s reconnecting, updating session %s with new addr %s", deviceID, ipStr, addr)
			// Update UDP address for existing session (no disconnect/connect notify)
			delete(s.addrMap, c.udpAddr.String())
			c.udpAddr = addr
			c.touch()
			s.addrMap[addr.String()] = ipStr
			s.mu.Unlock()

			// Update device map
			s.deviceMapMu.Lock()
			s.deviceMap[deviceID] = c
			s.deviceMapMu.Unlock()

			// Send AUTH_OK with existing IP
			resp := make([]byte, 5)
			resp[0] = TypeAuthOK
			copy(resp[1:5], c.vpnIP)
			s.udpConn.WriteToUDP(resp, addr)
			log.Printf("AUTH_OK (reconnect): device=%s ip=%s", deviceID, ipStr)
			return
		}
	}
	s.mu.Unlock()

	// Allocate IP
	ip, err := s.allocateIP()
	if err != nil {
		log.Printf("IP allocation failed for %s: %v", deviceID, err)
		s.sendAuthFail(addr)
		return
	}

	ipStr := ip.To4().String()
	c := &client{
		udpAddr:  addr,
		deviceID: deviceID,
		vpnIP:    ip.To4(),
	}
	c.touch()

	s.mu.Lock()
	s.clients[ipStr] = c
	s.addrMap[addr.String()] = ipStr
	s.mu.Unlock()

	s.deviceMapMu.Lock()
	s.deviceMap[deviceID] = c
	s.deviceMapMu.Unlock()

	// Send AUTH_OK with assigned IP
	resp := make([]byte, 5)
	resp[0] = TypeAuthOK
	copy(resp[1:5], ip.To4())
	s.udpConn.WriteToUDP(resp, addr)

	log.Printf("AUTH_OK: device=%s assigned ip=%s", deviceID, ipStr)

	// Notify API
	go s.notifyConnected(deviceID, ipStr)
}

func (s *tunnelServer) handlePing(addr *net.UDPAddr) {
	s.mu.RLock()
	ipStr, ok := s.addrMap[addr.String()]
	if ok {
		if c, exists := s.clients[ipStr]; exists {
			c.touch()
		}
	}
	s.mu.RUnlock()

	s.udpConn.WriteToUDP([]byte{TypePong}, addr)
}

func (s *tunnelServer) tunToUdp() {
	buf := s.tunBuf
	for {
		n, err := s.tunIface.Read(buf[1:]) // leave room for type prefix
		if err != nil {
			log.Printf("TUN read error: %v", err)
			continue
		}
		if n < 20 {
			continue
		}

		// Extract destination IP from IP header (bytes 16-19 of IP packet = buf[17:21])
		dstIP := net.IPv4(buf[17], buf[18], buf[19], buf[20]).String()

		s.mu.RLock()
		c, ok := s.clients[dstIP]
		s.mu.RUnlock()

		if !ok {
			continue
		}

		// Send DATA packet: [0x02][raw IP packet] — zero-copy from pre-allocated buffer
		buf[0] = TypeData
		s.udpConn.WriteToUDP(buf[:1+n], c.udpAddr)
	}
}

func (s *tunnelServer) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		var toRemove []string

		s.mu.RLock()
		for ipStr, c := range s.clients {
			if c.idleSince(now) > keepaliveTimeout {
				toRemove = append(toRemove, ipStr)
			}
		}
		s.mu.RUnlock()

		for _, ipStr := range toRemove {
			s.mu.Lock()
			c, ok := s.clients[ipStr]
			if ok {
				log.Printf("Client timeout: device=%s ip=%s (idle %v)",
					c.deviceID, ipStr, c.idleSince(now))
				delete(s.addrMap, c.udpAddr.String())
				delete(s.clients, ipStr)
				s.releaseIP(c.vpnIP)

				s.deviceMapMu.Lock()
				delete(s.deviceMap, c.deviceID)
				s.deviceMapMu.Unlock()

				go s.notifyDisconnected(c.deviceID, ipStr)
			}
			s.mu.Unlock()
		}
	}
}

type connInfo struct {
	Port      int    `json:"port"`
	ProxyType string `json:"proxy_type"`
}

func (s *tunnelServer) notifyConnected(deviceID, vpnIP string) {
	url := s.apiURL + "/api/internal/vpn/connected"
	body := fmt.Sprintf(`{"device_id":"%s","vpn_ip":"%s"}`, deviceID, vpnIP)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("Failed to notify connected for %s: %v", deviceID, err)
		return
	}
	defer resp.Body.Close()

	// Parse base_port and per-connection details from API response
	var result struct {
		BasePort    int        `json:"base_port"`
		Connections []connInfo `json:"connections"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.BasePort > 0 {
		setupDNAT(result.BasePort, vpnIP)
		for _, ci := range result.Connections {
			setupSingleDNAT(ci.Port, vpnIP, ci.ProxyType)
		}
		log.Printf("Notified API + DNAT setup: device %s vpn_ip=%s base_port=%d connections=%v", deviceID, vpnIP, result.BasePort, result.Connections)
	} else {
		log.Printf("Notified API: device %s connected (status=%d, no DNAT: base_port=%d)", deviceID, resp.StatusCode, result.BasePort)
	}
}

func (s *tunnelServer) notifyDisconnected(deviceID, vpnIP string) {
	url := s.apiURL + "/api/internal/vpn/disconnected"
	body := fmt.Sprintf(`{"device_id":"%s","vpn_ip":"%s"}`, deviceID, vpnIP)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("Failed to notify disconnected for %s: %v", deviceID, err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		BasePort    int        `json:"base_port"`
		Connections []connInfo `json:"connections"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.BasePort > 0 {
		teardownDNAT(result.BasePort, vpnIP)
		for _, ci := range result.Connections {
			teardownSingleDNAT(ci.Port, vpnIP, ci.ProxyType)
		}
		log.Printf("Notified API + DNAT teardown: device %s vpn_ip=%s base_port=%d connections=%v", deviceID, vpnIP, result.BasePort, result.Connections)
	} else {
		log.Printf("Notified API: device %s disconnected (status=%d)", deviceID, resp.StatusCode)
	}
}

// setupDNAT creates iptables DNAT rules to forward external ports to the device's VPN IP.
// Runs in the host network namespace (tunnel container uses network_mode: host).
func setupDNAT(basePort int, vpnIP string) {
	rules := []struct {
		extPort int
		devPort int
	}{
		{basePort, 8080},     // HTTP proxy
		{basePort + 1, 1080}, // SOCKS5
		{basePort + 2, 1081}, // UDP relay
	}
	for _, r := range rules {
		for _, proto := range []string{"tcp", "udp"} {
			args := fmt.Sprintf("-t nat -A PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
				proto, r.extPort, vpnIP, r.devPort)
			if out, err := runCmd("iptables", splitArgs(args)...); err != nil {
				log.Printf("DNAT add %s:%d->%s:%d (%s) failed: %s: %v", "ext", r.extPort, vpnIP, r.devPort, proto, string(out), err)
			}
		}
	}
	log.Printf("DNAT setup: ports %d-%d -> %s", basePort, basePort+2, vpnIP)
}

// teardownDNAT removes iptables DNAT rules for a device.
func teardownDNAT(basePort int, vpnIP string) {
	rules := []struct {
		extPort int
		devPort int
	}{
		{basePort, 8080},
		{basePort + 1, 1080},
		{basePort + 2, 1081},
	}
	for _, r := range rules {
		for _, proto := range []string{"tcp", "udp"} {
			args := fmt.Sprintf("-t nat -D PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
				proto, r.extPort, vpnIP, r.devPort)
			runCmd("iptables", splitArgs(args)...) // best effort
		}
	}
	log.Printf("DNAT teardown: ports %d-%d -> %s", basePort, basePort+2, vpnIP)
}

// setupSingleDNAT creates a single-port DNAT rule based on proxy type.
func setupSingleDNAT(extPort int, vpnIP string, proxyType string) {
	devPort := 8080 // HTTP proxy on device
	if proxyType == "socks5" {
		devPort = 1080
	}
	for _, proto := range []string{"tcp", "udp"} {
		args := fmt.Sprintf("-t nat -A PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
			proto, extPort, vpnIP, devPort)
		if out, err := runCmd("iptables", splitArgs(args)...); err != nil {
			log.Printf("DNAT add %d->%s:%d (%s) failed: %s: %v", extPort, vpnIP, devPort, proto, string(out), err)
		}
	}
	log.Printf("DNAT setup: port %d -> %s:%d (type=%s)", extPort, vpnIP, devPort, proxyType)
}

// teardownSingleDNAT removes a single-port DNAT rule based on proxy type.
func teardownSingleDNAT(extPort int, vpnIP string, proxyType string) {
	devPort := 8080
	if proxyType == "socks5" {
		devPort = 1080
	}
	for _, proto := range []string{"tcp", "udp"} {
		args := fmt.Sprintf("-t nat -D PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
			proto, extPort, vpnIP, devPort)
		runCmd("iptables", splitArgs(args)...) // best effort
	}
	log.Printf("DNAT teardown: port %d -> %s:%d (type=%s)", extPort, vpnIP, devPort, proxyType)
}

func splitArgs(s string) []string {
	var args []string
	for _, part := range strings.Fields(s) {
		args = append(args, part)
	}
	return args
}

func (s *tunnelServer) sendAuthFail(addr *net.UDPAddr) {
	s.udpConn.WriteToUDP([]byte{TypeAuthFail}, addr)
}

// tcpAuthListener handles TCP-based authentication.
// Clients that can't receive UDP (Samsung netfilter) use TCP for auth,
// then switch to UDP for the tunnel data relay.
// Protocol: client sends [0x01][16-byte device_id][4-byte UDP port big-endian]
// Server responds: [0x01][4-byte IPv4] on success, [0x03] on failure.
func (s *tunnelServer) tcpAuthListener(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on TCP port %d: %v", port, err)
	}
	log.Printf("TCP auth listener on port %d", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("TCP accept error: %v", err)
			continue
		}
		go s.handleTCPAuth(conn)
	}
}

func (s *tunnelServer) handleTCPAuth(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Read: [0x01][16-byte device_id][4-byte UDP port big-endian]
	buf := make([]byte, 21)
	n, err := conn.Read(buf)
	if err != nil || n < 21 || buf[0] != TypeAuth {
		log.Printf("TCP auth: invalid request from %s (n=%d, err=%v)", conn.RemoteAddr(), n, err)
		conn.Write([]byte{TypeAuthFail})
		return
	}

	deviceID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[1:5], buf[5:7], buf[7:9], buf[9:11], buf[11:17])
	udpPort := int(buf[17])<<24 | int(buf[18])<<16 | int(buf[19])<<8 | int(buf[20])

	// Get the client's IP from the TCP connection
	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	clientIP := tcpAddr.IP
	udpAddr := &net.UDPAddr{IP: clientIP, Port: udpPort}

	log.Printf("TCP AUTH from %s, device_id=%s, udp_port=%d", conn.RemoteAddr(), deviceID, udpPort)

	// Check if device already connected — update session
	s.mu.Lock()
	for ipStr, c := range s.clients {
		if c.deviceID == deviceID {
			log.Printf("Device %s reconnecting via TCP, updating session %s", deviceID, ipStr)
			delete(s.addrMap, c.udpAddr.String())
			c.udpAddr = udpAddr
			c.touch()
			s.addrMap[udpAddr.String()] = ipStr
			s.mu.Unlock()

			s.deviceMapMu.Lock()
			s.deviceMap[deviceID] = c
			s.deviceMapMu.Unlock()

			resp := make([]byte, 5)
			resp[0] = TypeAuthOK
			copy(resp[1:5], c.vpnIP)
			conn.Write(resp)
			log.Printf("TCP AUTH_OK (reconnect): device=%s ip=%s", deviceID, ipStr)
			return
		}
	}
	s.mu.Unlock()

	// Allocate IP
	ip, err := s.allocateIP()
	if err != nil {
		log.Printf("TCP auth: IP allocation failed for %s: %v", deviceID, err)
		conn.Write([]byte{TypeAuthFail})
		return
	}

	ipStr := ip.To4().String()
	c := &client{
		udpAddr:  udpAddr,
		deviceID: deviceID,
		vpnIP:    ip.To4(),
	}
	c.touch()

	s.mu.Lock()
	s.clients[ipStr] = c
	s.addrMap[udpAddr.String()] = ipStr
	s.mu.Unlock()

	s.deviceMapMu.Lock()
	s.deviceMap[deviceID] = c
	s.deviceMapMu.Unlock()

	// Send AUTH_OK
	resp := make([]byte, 5)
	resp[0] = TypeAuthOK
	copy(resp[1:5], ip.To4())
	conn.Write(resp)

	log.Printf("TCP AUTH_OK: device=%s assigned ip=%s udp=%s", deviceID, ipStr, udpAddr)
	go s.notifyConnected(deviceID, ipStr)
}

// startPushAPI starts an HTTP server for receiving command push requests from the API server.
// When a command is created, the API POSTs here and we relay it instantly to the device via UDP.
func (s *tunnelServer) startPushAPI() {
	pushPort := 8081
	if v := os.Getenv("PUSH_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			pushPort = p
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/push-command", s.handlePushCommand)
	mux.HandleFunc("/refresh-dnat", s.handleRefreshDNAT)
	mux.HandleFunc("/teardown-dnat", s.handleTeardownDNAT)
	mux.HandleFunc("/openvpn-client-connect", s.handleOpenVPNClientConnect)
	mux.HandleFunc("/openvpn-client-disconnect", s.handleOpenVPNClientDisconnect)

	log.Printf("Push API listening on port %d", pushPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", pushPort), mux); err != nil {
		log.Printf("Push API server failed: %v", err)
	}
}

func (s *tunnelServer) handlePushCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID string `json:"device_id"`
		ID       string `json:"id"`
		Type     string `json:"type"`
		Payload  string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Look up the device's UDP address
	s.deviceMapMu.RLock()
	c, ok := s.deviceMap[req.DeviceID]
	s.deviceMapMu.RUnlock()

	if !ok {
		http.Error(w, "device not connected", http.StatusNotFound)
		return
	}

	// Build command JSON to send to device
	cmdJSON, _ := json.Marshal(map[string]string{
		"id":      req.ID,
		"type":    req.Type,
		"payload": req.Payload,
	})

	// Send [0x05][json] via UDP
	pkt := make([]byte, 1+len(cmdJSON))
	pkt[0] = TypeCommand
	copy(pkt[1:], cmdJSON)

	if _, err := s.udpConn.WriteToUDP(pkt, c.udpAddr); err != nil {
		log.Printf("Push command to device %s failed: %v", req.DeviceID, err)
		http.Error(w, "send failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Pushed command %s (%s) to device %s", req.ID, req.Type, req.DeviceID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (s *tunnelServer) handleRefreshDNAT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID  string `json:"device_id"`
		BasePort  int    `json:"base_port"`
		VpnIP     string `json:"vpn_ip"`
		ProxyType string `json:"proxy_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.BasePort > 0 && req.VpnIP != "" {
		if req.ProxyType != "" {
			setupSingleDNAT(req.BasePort, req.VpnIP, req.ProxyType)
		} else {
			setupDNAT(req.BasePort, req.VpnIP)
		}
		log.Printf("Refresh DNAT: device=%s base_port=%d vpn_ip=%s type=%s", req.DeviceID, req.BasePort, req.VpnIP, req.ProxyType)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (s *tunnelServer) handleTeardownDNAT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DeviceID  string `json:"device_id"`
		BasePort  int    `json:"base_port"`
		VpnIP     string `json:"vpn_ip"`
		ProxyType string `json:"proxy_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.BasePort > 0 && req.VpnIP != "" {
		if req.ProxyType != "" {
			teardownSingleDNAT(req.BasePort, req.VpnIP, req.ProxyType)
		} else {
			teardownDNAT(req.BasePort, req.VpnIP)
		}
		log.Printf("Teardown DNAT: device=%s base_port=%d vpn_ip=%s type=%s", req.DeviceID, req.BasePort, req.VpnIP, req.ProxyType)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// handleOpenVPNClientConnect is called by the API when an OpenVPN client connects.
// It adds a transparent proxy mapping and iptables REDIRECT rule for the client's VPN IP.
func (s *tunnelServer) handleOpenVPNClientConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientVPNIP string `json:"client_vpn_ip"`  // 10.9.0.x
		DeviceVPNIP string `json:"device_vpn_ip"`  // 192.168.255.y
		ProxyPort   int    `json:"proxy_port"`      // HTTP proxy port (default 8080)
		SOCKSPort   int    `json:"socks_port"`      // legacy, unused
		ProxyUser   string `json:"socks_user"`      // proxy username
		ProxyPass   string `json:"socks_pass"`      // proxy password
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.ProxyPort == 0 {
		req.ProxyPort = 8080
	}

	proxyEndpoint := fmt.Sprintf("%s:%d", req.DeviceVPNIP, req.ProxyPort)

	// Add transparent proxy mapping with HTTP CONNECT proxy
	if s.transparentProxy != nil {
		s.transparentProxy.AddMapping(req.ClientVPNIP, proxyEndpoint, req.ProxyUser, req.ProxyPass)
	}

	// Add iptables REDIRECT rule: TCP traffic from this client -> transparent proxy port 12345
	args := fmt.Sprintf("-t nat -A PREROUTING -s %s -p tcp -j REDIRECT --to-port 12345", req.ClientVPNIP)
	if out, err := runCmd("iptables", splitArgs(args)...); err != nil {
		log.Printf("OpenVPN REDIRECT rule for %s failed: %s: %v", req.ClientVPNIP, string(out), err)
	} else {
		log.Printf("OpenVPN REDIRECT rule added for %s -> tproxy (proxy=%s)", req.ClientVPNIP, proxyEndpoint)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// handleOpenVPNClientDisconnect is called by the API when an OpenVPN client disconnects.
// It removes the transparent proxy mapping and iptables REDIRECT rule.
func (s *tunnelServer) handleOpenVPNClientDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientVPNIP string `json:"client_vpn_ip"` // 10.9.0.x
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Remove transparent proxy mapping
	if s.transparentProxy != nil {
		s.transparentProxy.RemoveMapping(req.ClientVPNIP)
	}

	// Remove iptables REDIRECT rule
	args := fmt.Sprintf("-t nat -D PREROUTING -s %s -p tcp -j REDIRECT --to-port 12345", req.ClientVPNIP)
	runCmd("iptables", splitArgs(args)...) // best effort
	log.Printf("OpenVPN client disconnect: removed mapping and REDIRECT for %s", req.ClientVPNIP)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
