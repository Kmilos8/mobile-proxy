package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

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
	socksForwardPort = 12345 // transparent TCP → SOCKS5 forwarder
	ovpnSubnet       = "10.9.0.0/24"
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

	// NAT routing: policy routing tables for OpenVPN client traffic
	routingMu          sync.Mutex
	deviceRouteTable   map[string]int              // device VPN IP (192.168.255.x) -> routing table number
	clientToDevice     map[string]string           // client VPN IP (10.9.0.x) -> device VPN IP (192.168.255.x)
	clientSocksAuth    map[string]socksAuth        // client VPN IP (10.9.0.x) -> SOCKS5 credentials
	clientBandwidthUsed  map[string]*atomic.Int64  // client VPN IP -> bytes forwarded
	clientBandwidthLimit map[string]int64          // client VPN IP -> limit (0=unlimited)

	// Per-port bandwidth tracking for HTTP/SOCKS5 DNAT connections
	portToUsername    map[int]string    // external port -> connection username
	portBandwidthAcc map[int]int64     // external port -> accumulated bytes (running total)
}

type socksAuth struct {
	user string
	pass string
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

	srv := &tunnelServer{
		udpConn:              conn,
		tunIface:             iface,
		apiURL:               apiURL,
		clients:              make(map[string]*client),
		addrMap:              make(map[string]string),
		deviceMap:            make(map[string]*client),
		ipPool:               make([]bool, ipPoolEnd-ipPoolStart+1),
		tunBuf:               make([]byte, maxPacketSize+1), // +1 for type prefix
		deviceRouteTable:     make(map[string]int),
		clientToDevice:       make(map[string]string),
		clientSocksAuth:      make(map[string]socksAuth),
		clientBandwidthUsed:  make(map[string]*atomic.Int64),
		clientBandwidthLimit: make(map[string]int64),
		portToUsername:       make(map[int]string),
		portBandwidthAcc:     make(map[int]int64),
	}

	// Start goroutines
	go srv.udpToTun()
	go srv.tunToUdp()
	go srv.cleanupLoop()
	go srv.tcpAuthListener(port)
	go srv.startPushAPI()
	go srv.startSocksForwarder()

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

	// Increase TUN queue length for high throughput
	if out, err := runCmd("ip", "link", "set", "dev", name, "txqueuelen", "5000"); err != nil {
		log.Printf("Warning: txqueuelen: %s: %v", string(out), err)
	}

	// Enable IP forwarding
	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// TCP/network performance tuning for mobile VPN throughput
	// Load BBR congestion control module (model-based, handles mobile packet loss gracefully)
	exec.Command("modprobe", "tcp_bbr").Run()

	sysctls := map[string]string{
		"net.ipv4.tcp_congestion_control": "bbr",
		"net.core.rmem_max":              "16777216",               // 16MB
		"net.core.wmem_max":              "16777216",               // 16MB
		"net.ipv4.tcp_rmem":              "4096 1048576 16777216",  // min 4K, default 1MB, max 16MB
		"net.ipv4.tcp_wmem":              "4096 1048576 16777216",  // min 4K, default 1MB (was 16KB!), max 16MB
		"net.core.netdev_max_backlog":    "5000",
		"net.ipv4.tcp_mtu_probing":       "1",
	}
	for k, v := range sysctls {
		if out, err := exec.Command("sysctl", "-w", k+"="+v).CombinedOutput(); err != nil {
			log.Printf("Warning: sysctl %s=%s failed: %s: %v", k, v, string(out), err)
		}
	}
	log.Printf("TCP tuning applied (BBR + buffer sizes)")

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

	// OpenVPN client subnet (10.9.0.0/24) — FORWARD rules only.
	// No MASQUERADE: packets keep their original source (10.9.0.x) so the server's
	// tunToUdp can match them to the correct device via clientToDevice map.
	// Response packets from the phone with dst=10.9.0.x route back via tun1 (OpenVPN).
	ovpnSubnet := "10.9.0.0/24"
	// Remove any leftover MASQUERADE from previous version
	for {
		if _, err := runCmd("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", ovpnSubnet, "-j", "MASQUERADE"); err != nil {
			break
		}
	}
	log.Printf("Removed MASQUERADE for %s (NAT routing preserves source IP)", ovpnSubnet)
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
	// Allow DNAT'd TCP from OpenVPN clients to reach the SOCKS5 forwarder.
	// The nft INPUT chain has DROP policy (UFW), so we need an explicit ACCEPT.
	// DNAT rewrites dst to 192.168.255.1:12345 (tun0 IP), so match on dst+dport.
	fwdPort := strconv.Itoa(socksForwardPort)
	if _, err := runCmd("iptables", "-C", "INPUT", "-d", tunIP, "-p", "tcp", "--dport", fwdPort, "-j", "ACCEPT"); err != nil {
		if out, err := runCmd("iptables", "-I", "INPUT", "1", "-d", tunIP, "-p", "tcp", "--dport", fwdPort, "-j", "ACCEPT"); err != nil {
			log.Printf("Warning: INPUT ACCEPT for SOCKS forwarder failed: %s: %v", string(out), err)
		} else {
			log.Printf("Added INPUT ACCEPT for dst %s tcp dport %s", tunIP, fwdPort)
		}
	}

	// Blackhole safety: unmapped OpenVPN clients can't leak through server's internet.
	// Table 99 has only "blackhole default" — any client without a specific routing rule
	// will hit this and get dropped.
	runCmd("ip", "route", "replace", "blackhole", "default", "table", "99")
	runCmd("ip", "rule", "del", "from", ovpnSubnet, "priority", "32000") // idempotent cleanup
	if out, err := runCmd("ip", "rule", "add", "from", ovpnSubnet, "lookup", "99", "priority", "32000"); err != nil {
		log.Printf("Warning: blackhole rule failed: %s: %v", string(out), err)
	} else {
		log.Printf("Blackhole safety net: unmapped %s clients -> table 99", ovpnSubnet)
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
			go s.setupDeviceRouting(ipStr)
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

	// Set up routing table for this device + notify API
	go s.setupDeviceRouting(ipStr)
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

		if ok {
			// Direct match: packet addressed to a registered device VPN IP
			buf[0] = TypeData
			s.udpConn.WriteToUDP(buf[:1+n], c.udpAddr)
			continue
		}

		// NAT-routed traffic: packet from OpenVPN client (10.9.0.x) routed through
		// a device's routing table. Source IP tells us which device should get it.
		srcIP := net.IPv4(buf[13], buf[14], buf[15], buf[16]).String()
		s.routingMu.Lock()
		deviceIP, mapped := s.clientToDevice[srcIP]
		ctr := s.clientBandwidthUsed[srcIP]
		limit := s.clientBandwidthLimit[srcIP]
		s.routingMu.Unlock()

		if mapped {
			// Bandwidth enforcement — atomic increment outside lock
			var used int64
			if ctr != nil {
				used = ctr.Add(int64(n))
			}
			// Hard cutoff: drop packet silently if limit exceeded
			if limit > 0 && used > limit {
				continue
			}

			s.mu.RLock()
			c, ok = s.clients[deviceIP]
			s.mu.RUnlock()
			if ok {
				buf[0] = TypeData
				s.udpConn.WriteToUDP(buf[:1+n], c.udpAddr)
			}
		}
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
				go s.teardownDeviceRouting(ipStr)
			}
			s.mu.Unlock()
		}
	}
}

// routingTableForDevice returns a routing table number based on the device's VPN IP.
// 192.168.255.2 -> table 102, 192.168.255.3 -> table 103, etc.
func routingTableForDevice(deviceVPNIP string) (int, error) {
	ip := net.ParseIP(deviceVPNIP)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP: %s", deviceVPNIP)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("not IPv4: %s", deviceVPNIP)
	}
	return int(ip4[3]) + 100, nil
}

// setupDeviceRouting creates a policy routing table for a device.
// After this, any client with an ip rule pointing to this table will have their
// traffic routed through the device's VPN IP on tun0.
func (s *tunnelServer) setupDeviceRouting(deviceVPNIP string) {
	tableNum, err := routingTableForDevice(deviceVPNIP)
	if err != nil {
		log.Printf("[routing] setupDeviceRouting failed: %v", err)
		return
	}

	// Create routing table: default route via device IP through tun0
	tableStr := strconv.Itoa(tableNum)
	if out, err := runCmd("ip", "route", "replace", "default", "via", deviceVPNIP, "dev", tunName, "table", tableStr); err != nil {
		log.Printf("[routing] ip route replace table %s failed: %s: %v", tableStr, string(out), err)
		return
	}

	s.routingMu.Lock()
	s.deviceRouteTable[deviceVPNIP] = tableNum
	s.routingMu.Unlock()

	log.Printf("[routing] setup: table %d -> default via %s dev %s", tableNum, deviceVPNIP, tunName)
}

// teardownDeviceRouting removes the routing table and all client ip rules for a device.
func (s *tunnelServer) teardownDeviceRouting(deviceVPNIP string) {
	s.routingMu.Lock()
	tableNum, ok := s.deviceRouteTable[deviceVPNIP]
	if !ok {
		s.routingMu.Unlock()
		return
	}
	delete(s.deviceRouteTable, deviceVPNIP)

	// Find and remove all client rules pointing to this device
	var clientsToRemove []string
	for clientIP, devIP := range s.clientToDevice {
		if devIP == deviceVPNIP {
			clientsToRemove = append(clientsToRemove, clientIP)
		}
	}
	for _, clientIP := range clientsToRemove {
		delete(s.clientToDevice, clientIP)
		delete(s.clientSocksAuth, clientIP)
		delete(s.clientBandwidthUsed, clientIP)
		delete(s.clientBandwidthLimit, clientIP)
	}
	s.routingMu.Unlock()

	// Remove client ip rules and iptables REDIRECT rules
	for _, clientIP := range clientsToRemove {
		for {
			if _, err := runCmd("ip", "rule", "del", "from", clientIP+"/32", "priority", "100"); err != nil {
				break
			}
		}
		dnatTarget := tunIP + ":" + strconv.Itoa(socksForwardPort)
		for {
			if _, err := runCmd("iptables", "-t", "nat", "-D", "PREROUTING",
				"-s", clientIP+"/32", "-p", "tcp", "-j", "DNAT",
				"--to-destination", dnatTarget); err != nil {
				break
			}
		}
		log.Printf("[routing] removed client rules: %s/32", clientIP)
	}

	// Remove routing table
	tableStr := strconv.Itoa(tableNum)
	runCmd("ip", "route", "flush", "table", tableStr)
	log.Printf("[routing] teardown: table %d for device %s", tableNum, deviceVPNIP)
}

type connInfo struct {
	Port      int    `json:"port"`
	ProxyType string `json:"proxy_type"`
	Username  string `json:"username"`
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
			// Track port→username mapping for bandwidth accounting
			if ci.Username != "" {
				s.routingMu.Lock()
				s.portToUsername[ci.Port] = ci.Username
				s.routingMu.Unlock()
			}
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
			// Clean up port→username tracking
			s.routingMu.Lock()
			delete(s.portToUsername, ci.Port)
			delete(s.portBandwidthAcc, ci.Port)
			s.routingMu.Unlock()
		}
		log.Printf("Notified API + DNAT teardown: device %s vpn_ip=%s base_port=%d connections=%v", deviceID, vpnIP, result.BasePort, result.Connections)
	} else {
		log.Printf("Notified API: device %s disconnected (status=%d)", deviceID, resp.StatusCode)
	}
}

// setupDNAT creates iptables DNAT rules to forward external ports to the device's VPN IP.
// Runs in the host network namespace (tunnel container uses network_mode: host).
func setupDNAT(basePort int, vpnIP string) {
	// Teardown any existing rules first to prevent duplicates
	teardownDNAT(basePort, vpnIP)

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
// Loops to remove ALL duplicate rules, not just the first match.
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
			// Loop to remove all duplicates
			for {
				if _, err := runCmd("iptables", splitArgs(args)...); err != nil {
					break // no more matching rules
				}
			}
		}
	}
	log.Printf("DNAT teardown: ports %d-%d -> %s", basePort, basePort+2, vpnIP)
}

// setupSingleDNAT creates a single-port DNAT rule based on proxy type.
func setupSingleDNAT(extPort int, vpnIP string, proxyType string) {
	// Teardown any existing rules first to prevent duplicates
	teardownSingleDNAT(extPort, vpnIP, proxyType)

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
// Loops to remove ALL duplicate rules, not just the first match.
func teardownSingleDNAT(extPort int, vpnIP string, proxyType string) {
	devPort := 8080
	if proxyType == "socks5" {
		devPort = 1080
	}
	for _, proto := range []string{"tcp", "udp"} {
		args := fmt.Sprintf("-t nat -D PREROUTING -p %s --dport %d -j DNAT --to-destination %s:%d",
			proto, extPort, vpnIP, devPort)
		// Loop to remove all duplicates
		for {
			if _, err := runCmd("iptables", splitArgs(args)...); err != nil {
				break
			}
		}
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

	go s.bandwidthFlushLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/push-command", s.handlePushCommand)
	mux.HandleFunc("/refresh-dnat", s.handleRefreshDNAT)
	mux.HandleFunc("/teardown-dnat", s.handleTeardownDNAT)
	mux.HandleFunc("/openvpn-client-connect", s.handleOpenVPNClientConnect)
	mux.HandleFunc("/openvpn-client-disconnect", s.handleOpenVPNClientDisconnect)
	mux.HandleFunc("/openvpn-client-reset-bandwidth", s.handleResetBandwidth)

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
		Username  string `json:"username"`
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
		// Track port→username mapping for bandwidth accounting
		if req.Username != "" {
			s.routingMu.Lock()
			s.portToUsername[req.BasePort] = req.Username
			s.routingMu.Unlock()
		}
		log.Printf("Refresh DNAT: device=%s base_port=%d vpn_ip=%s type=%s username=%s", req.DeviceID, req.BasePort, req.VpnIP, req.ProxyType, req.Username)
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
		// Clean up port→username tracking
		s.routingMu.Lock()
		delete(s.portToUsername, req.BasePort)
		delete(s.portBandwidthAcc, req.BasePort)
		s.routingMu.Unlock()
		log.Printf("Teardown DNAT: device=%s base_port=%d vpn_ip=%s type=%s", req.DeviceID, req.BasePort, req.VpnIP, req.ProxyType)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// handleOpenVPNClientConnect is called by the API when an OpenVPN client connects.
// Sets up:
//  1. ip rule for UDP routing through device tunnel (DNS, QUIC)
//  2. iptables REDIRECT for TCP → SOCKS5 forwarder (high-speed path)
//  3. SOCKS5 credentials for the forwarder to authenticate to the phone's proxy
func (s *tunnelServer) handleOpenVPNClientConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientVPNIP    string `json:"client_vpn_ip"`    // 10.9.0.x
		DeviceVPNIP    string `json:"device_vpn_ip"`    // 192.168.255.y
		SocksUser      string `json:"socks_user"`
		SocksPass      string `json:"socks_pass"`
		BandwidthLimit int64  `json:"bandwidth_limit"` // bytes, 0 = unlimited
		BandwidthUsed  int64  `json:"bandwidth_used"`  // current DB value — initial offset
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Look up routing table for this device
	s.routingMu.Lock()
	tableNum, ok := s.deviceRouteTable[req.DeviceVPNIP]
	if !ok {
		s.routingMu.Unlock()
		log.Printf("[routing] no routing table for device %s, client %s cannot connect", req.DeviceVPNIP, req.ClientVPNIP)
		http.Error(w, "device routing not ready", http.StatusServiceUnavailable)
		return
	}
	s.clientToDevice[req.ClientVPNIP] = req.DeviceVPNIP
	s.clientSocksAuth[req.ClientVPNIP] = socksAuth{user: req.SocksUser, pass: req.SocksPass}
	// Initialize bandwidth counter from DB value (survives tunnel restarts)
	ctr := &atomic.Int64{}
	ctr.Store(req.BandwidthUsed)
	s.clientBandwidthUsed[req.ClientVPNIP] = ctr
	s.clientBandwidthLimit[req.ClientVPNIP] = req.BandwidthLimit
	s.routingMu.Unlock()

	// Remove any stale ip rule for this client (idempotent)
	for {
		if _, err := runCmd("ip", "rule", "del", "from", req.ClientVPNIP+"/32", "priority", "100"); err != nil {
			break
		}
	}

	// Add ip rule: UDP traffic from this client uses the device's routing table
	tableStr := strconv.Itoa(tableNum)
	if out, err := runCmd("ip", "rule", "add", "from", req.ClientVPNIP+"/32", "lookup", tableStr, "priority", "100"); err != nil {
		log.Printf("[routing] ip rule add failed: %s: %v", string(out), err)
		http.Error(w, "routing rule failed", http.StatusInternalServerError)
		return
	}

	// Remove stale iptables DNAT rule for TCP (idempotent)
	dnatTarget := tunIP + ":" + strconv.Itoa(socksForwardPort) // 192.168.255.1:12345
	for {
		if _, err := runCmd("iptables", "-t", "nat", "-D", "PREROUTING",
			"-s", req.ClientVPNIP+"/32", "-p", "tcp", "-j", "DNAT",
			"--to-destination", dnatTarget); err != nil {
			break
		}
	}

	// Add iptables DNAT: TCP from this client → SOCKS5 forwarder on tun0 IP
	// (REDIRECT won't work: it rewrites dst to tun1 address, which causes
	// "cross-device link" routing failure with source-based ip rules)
	if out, err := runCmd("iptables", "-t", "nat", "-I", "PREROUTING",
		"-s", req.ClientVPNIP+"/32", "-p", "tcp", "-j", "DNAT",
		"--to-destination", dnatTarget); err != nil {
		log.Printf("[routing] iptables DNAT add failed: %s: %v", string(out), err)
	}

	log.Printf("[routing] client %s -> table %d (device %s) + TCP DNAT -> %s (socks_user=%s)",
		req.ClientVPNIP, tableNum, req.DeviceVPNIP, dnatTarget, req.SocksUser)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// handleOpenVPNClientDisconnect is called by the API when an OpenVPN client disconnects.
// Removes ip rule, iptables REDIRECT, and SOCKS5 credentials for the client.
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

	// Remove from client mappings
	s.routingMu.Lock()
	delete(s.clientToDevice, req.ClientVPNIP)
	delete(s.clientSocksAuth, req.ClientVPNIP)
	delete(s.clientBandwidthUsed, req.ClientVPNIP)
	delete(s.clientBandwidthLimit, req.ClientVPNIP)
	s.routingMu.Unlock()

	// Remove ip rules — loop to remove all duplicates
	for {
		if _, err := runCmd("ip", "rule", "del", "from", req.ClientVPNIP+"/32", "priority", "100"); err != nil {
			break
		}
	}

	// Remove iptables DNAT rules
	dnatTarget := tunIP + ":" + strconv.Itoa(socksForwardPort)
	for {
		if _, err := runCmd("iptables", "-t", "nat", "-D", "PREROUTING",
			"-s", req.ClientVPNIP+"/32", "-p", "tcp", "-j", "DNAT",
			"--to-destination", dnatTarget); err != nil {
			break
		}
	}

	log.Printf("[routing] client disconnect: removed rules for %s", req.ClientVPNIP)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// handleResetBandwidth resets the in-memory bandwidth counter for a client VPN IP.
// Accepts client_vpn_ip directly or username for reverse lookup.
func (s *tunnelServer) handleResetBandwidth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ClientVPNIP string `json:"client_vpn_ip"`
		Username    string `json:"username"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.routingMu.Lock()
	targetIP := req.ClientVPNIP
	if targetIP == "" && req.Username != "" {
		for ip, auth := range s.clientSocksAuth {
			if auth.user == req.Username {
				targetIP = ip
				break
			}
		}
	}
	if targetIP != "" {
		if ctr, ok := s.clientBandwidthUsed[targetIP]; ok {
			ctr.Store(0)
		}
	}
	// Also reset DNAT port bandwidth accumulator for this username
	if req.Username != "" {
		for port, user := range s.portToUsername {
			if user == req.Username {
				s.portBandwidthAcc[port] = 0
			}
		}
	}
	s.routingMu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// bandwidthFlushLoop periodically flushes bandwidth counters to the API server.
func (s *tunnelServer) bandwidthFlushLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.flushBandwidthToAPI()
	}
}

// readDNATBandwidth reads iptables byte counters for DNAT rules, zeros them,
// and returns accumulated bandwidth per username for HTTP/SOCKS5 proxy connections.
func (s *tunnelServer) readDNATBandwidth() map[string]int64 {
	// Read counters: -v for verbose (includes bytes), -n for numeric, -x for exact counts
	out, err := runCmd("iptables", "-t", "nat", "-L", "PREROUTING", "-v", "-n", "-x")
	if err != nil {
		log.Printf("[bandwidth] iptables read failed: %v", err)
		return nil
	}

	// Parse output lines. Format example:
	//   pkts bytes target prot opt in out source destination
	//   42  12345 DNAT   tcp  --  *  *   0.0.0.0/0  0.0.0.0/0  tcp dpt:30048 to:192.168.255.2:8080
	portBytes := make(map[int]int64)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "DNAT") || !strings.Contains(line, "tcp") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		byteCount, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || byteCount == 0 {
			continue
		}
		// Find dpt:NNNNN in the fields
		for _, f := range fields {
			if strings.HasPrefix(f, "dpt:") {
				portStr := strings.TrimPrefix(f, "dpt:")
				port, err := strconv.Atoi(portStr)
				if err == nil {
					portBytes[port] += byteCount
				}
			}
		}
	}

	// Zero the counters after reading
	if _, err := runCmd("iptables", "-t", "nat", "-Z", "PREROUTING"); err != nil {
		log.Printf("[bandwidth] iptables zero failed: %v", err)
	}

	// Map port bytes to usernames using portToUsername
	result := make(map[string]int64)
	s.routingMu.Lock()
	for port, bytes := range portBytes {
		if username, ok := s.portToUsername[port]; ok {
			s.portBandwidthAcc[port] += bytes
			result[username] += s.portBandwidthAcc[port]
		}
	}
	// Also include ports that had no new traffic but have accumulated bytes
	for port, acc := range s.portBandwidthAcc {
		if acc > 0 {
			if username, ok := s.portToUsername[port]; ok {
				if _, already := result[username]; !already {
					result[username] = acc
				}
			}
		}
	}
	s.routingMu.Unlock()

	return result
}

// flushBandwidthToAPI sends a snapshot of bandwidth usage (by username) to the API.
func (s *tunnelServer) flushBandwidthToAPI() {
	// 1. Snapshot OpenVPN client bandwidth (tracked via SOCKS5 forwarder)
	s.routingMu.Lock()
	snapshot := make(map[string]int64)
	for ip, ctr := range s.clientBandwidthUsed {
		if auth, ok := s.clientSocksAuth[ip]; ok {
			snapshot[auth.user] = ctr.Load()
		}
	}
	s.routingMu.Unlock()

	// 2. Read DNAT (HTTP/SOCKS5 proxy) bandwidth from iptables counters
	dnatBW := s.readDNATBandwidth()
	for username, bytes := range dnatBW {
		// A connection is either OpenVPN or HTTP/SOCKS5, so they won't overlap.
		// Use max in case both somehow report for the same username.
		if existing, ok := snapshot[username]; ok {
			if bytes > existing {
				snapshot[username] = bytes
			}
		} else {
			snapshot[username] = bytes
		}
	}

	if len(snapshot) == 0 {
		return
	}

	body, _ := json.Marshal(snapshot)
	apiURL := s.apiURL
	if apiURL == "" {
		apiURL = "http://127.0.0.1:8080"
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(apiURL+"/api/internal/bandwidth-flush", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[bandwidth] flush failed: %v", err)
		return
	}
	resp.Body.Close()
	log.Printf("[bandwidth] flushed %d connections to API (ovpn=%d, dnat=%d)", len(snapshot), len(snapshot)-len(dnatBW), len(dnatBW))
}

// ──────────────────────────────────────────────────────────────────────────────
// SOCKS5 transparent forwarder
//
// TCP from OpenVPN clients is iptables-REDIRECTed to this listener.
// For each connection:
//   1. SO_ORIGINAL_DST → real destination (e.g. google.com:443)
//   2. Source IP → look up device + SOCKS5 credentials
//   3. Connect to device's SOCKS5 proxy (192.168.255.x:1080) via tun0
//   4. SOCKS5 CONNECT to original destination
//   5. Bidirectional relay with io.Copy
//
// This achieves high throughput because the server↔device path uses kernel TCP
// (through tun0/UDP tunnel), and the device↔internet path uses kernel TCP over
// cellular. No userspace TCP packet construction anywhere.
// ──────────────────────────────────────────────────────────────────────────────

func (s *tunnelServer) startSocksForwarder() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", socksForwardPort))
	if err != nil {
		log.Fatalf("[socks-fwd] failed to listen on port %d: %v", socksForwardPort, err)
	}
	log.Printf("[socks-fwd] listening on port %d", socksForwardPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[socks-fwd] accept error: %v", err)
			continue
		}
		go s.handleForwardedConn(conn)
	}
}

func (s *tunnelServer) handleForwardedConn(conn net.Conn) {
	defer conn.Close()

	// 1. Get the real destination before iptables REDIRECT changed it
	origIP, origPort, err := getOriginalDst(conn)
	if err != nil {
		log.Printf("[socks-fwd] getOriginalDst failed: %v", err)
		return
	}

	// 2. Look up which device to route through (based on source IP)
	srcAddr := conn.RemoteAddr().(*net.TCPAddr)
	srcIP := srcAddr.IP.String()

	s.routingMu.Lock()
	deviceIP, ok := s.clientToDevice[srcIP]
	auth := s.clientSocksAuth[srcIP]
	s.routingMu.Unlock()

	if !ok {
		log.Printf("[socks-fwd] no device mapping for client %s", srcIP)
		return
	}

	// 3. Connect to device's SOCKS5 proxy via tun0
	socksAddr := fmt.Sprintf("%s:1080", deviceIP)
	socksConn, err := net.DialTimeout("tcp", socksAddr, 10*time.Second)
	if err != nil {
		log.Printf("[socks-fwd] dial %s failed: %v", socksAddr, err)
		return
	}
	defer socksConn.Close()

	// Disable Nagle — send SOCKS5 greeting immediately
	if tc, ok := socksConn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}

	// 4. SOCKS5 handshake (username/password auth + CONNECT)
	socksConn.SetDeadline(time.Now().Add(10 * time.Second))
	dstStr := origIP.String()
	if err := socks5Connect(socksConn, auth.user, auth.pass, dstStr, origPort); err != nil {
		log.Printf("[socks-fwd] SOCKS5 handshake to %s for %s:%d failed: %v",
			socksAddr, dstStr, origPort, err)
		return
	}
	socksConn.SetDeadline(time.Time{}) // clear deadline for relay

	// 5. Bidirectional relay
	done := make(chan struct{})
	go func() {
		io.Copy(socksConn, conn)
		if tc, ok := socksConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
		done <- struct{}{}
	}()
	io.Copy(conn, socksConn)
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.CloseWrite()
	}
	<-done
}

// getOriginalDst retrieves the original destination address of a connection
// that was redirected by iptables REDIRECT. Uses the SO_ORIGINAL_DST socket option.
func getOriginalDst(conn net.Conn) (net.IP, uint16, error) {
	const soOriginalDst = 80

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil, 0, fmt.Errorf("not a TCP connection")
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return nil, 0, err
	}

	var addr [16]byte // sockaddr_in: family(2) + port(2) + addr(4) + padding(8)
	var sockErr error

	err = rawConn.Control(func(fd uintptr) {
		addrLen := uint32(16)
		_, _, errno := syscall.Syscall6(
			syscall.SYS_GETSOCKOPT,
			fd,
			uintptr(syscall.SOL_IP),
			soOriginalDst,
			uintptr(unsafe.Pointer(&addr[0])),
			uintptr(unsafe.Pointer(&addrLen)),
			0,
		)
		if errno != 0 {
			sockErr = errno
		}
	})
	if err != nil {
		return nil, 0, err
	}
	if sockErr != nil {
		return nil, 0, sockErr
	}

	// struct sockaddr_in layout: family(2) + port(2 big-endian) + addr(4)
	port := binary.BigEndian.Uint16(addr[2:4])
	ip := net.IPv4(addr[4], addr[5], addr[6], addr[7])
	return ip, port, nil
}

// socks5Connect performs SOCKS5 handshake with username/password auth and CONNECT.
func socks5Connect(conn net.Conn, user, pass, dstIP string, dstPort uint16) error {
	// Auth negotiation: offer username/password method (0x02)
	if _, err := conn.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		return fmt.Errorf("auth negotiation write: %w", err)
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("auth negotiation read: %w", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x02 {
		return fmt.Errorf("auth method rejected: %x %x", resp[0], resp[1])
	}

	// Username/password auth (RFC 1929)
	authBuf := make([]byte, 3+len(user)+len(pass))
	authBuf[0] = 0x01 // subneg version
	authBuf[1] = byte(len(user))
	copy(authBuf[2:], user)
	authBuf[2+len(user)] = byte(len(pass))
	copy(authBuf[3+len(user):], pass)
	if _, err := conn.Write(authBuf); err != nil {
		return fmt.Errorf("auth write: %w", err)
	}

	if _, err := io.ReadFull(conn, resp); err != nil {
		return fmt.Errorf("auth response read: %w", err)
	}
	if resp[1] != 0x00 {
		return fmt.Errorf("auth failed: status %d", resp[1])
	}

	// CONNECT request (IPv4)
	ip := net.ParseIP(dstIP).To4()
	if ip == nil {
		return fmt.Errorf("invalid IPv4: %s", dstIP)
	}
	req := make([]byte, 10)
	req[0] = 0x05 // version
	req[1] = 0x01 // CONNECT
	req[2] = 0x00 // reserved
	req[3] = 0x01 // IPv4
	copy(req[4:8], ip)
	binary.BigEndian.PutUint16(req[8:10], dstPort)
	if _, err := conn.Write(req); err != nil {
		return fmt.Errorf("connect write: %w", err)
	}

	// Read reply (minimum 10 bytes for IPv4)
	reply := make([]byte, 10)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return fmt.Errorf("connect reply read: %w", err)
	}
	if reply[1] != 0x00 {
		return fmt.Errorf("connect failed: status %d", reply[1])
	}

	return nil
}
