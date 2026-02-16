package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

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
)

const (
	defaultPort      = 1194
	defaultAPIURL    = "http://127.0.0.1:8080"
	tunName          = "tun0"
	tunIP            = "192.168.255.1"
	tunSubnet        = "192.168.255.0/24"
	tunMTU           = 1400
	maxPacketSize    = 1500
	keepaliveTimeout = 60 * time.Second
	cleanupInterval  = 10 * time.Second
	ipPoolStart      = 2
	ipPoolEnd        = 254
	deviceIDLen      = 16
)

type client struct {
	udpAddr  *net.UDPAddr
	deviceID string
	vpnIP    net.IP
	lastSeen time.Time
}

type tunnelServer struct {
	udpConn  *net.UDPConn
	tunIface *water.Interface
	apiURL   string

	mu      sync.RWMutex
	clients map[string]*client // vpnIP string -> client
	addrMap map[string]string  // udpAddr string -> vpnIP string

	ipPool   []bool // true = in use, index 0 = .2
	ipPoolMu sync.Mutex
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
	log.Printf("Listening on UDP port %d", port)

	srv := &tunnelServer{
		udpConn:  conn,
		tunIface: iface,
		apiURL:   apiURL,
		clients:  make(map[string]*client),
		addrMap:  make(map[string]string),
		ipPool:   make([]bool, ipPoolEnd-ipPoolStart+1),
	}

	// Start goroutines
	go srv.udpToTun()
	go srv.tunToUdp()
	go srv.cleanupLoop()

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
	buf := make([]byte, maxPacketSize)
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
		case TypeAuth:
			s.handleAuth(buf[1:n], remoteAddr)
		case TypeData:
			s.handleData(buf[1:n], remoteAddr)
		case TypePing:
			s.handlePing(remoteAddr)
		default:
			log.Printf("Unknown packet type 0x%02x from %s", pktType, remoteAddr)
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

	// Check if this device is already connected — disconnect old session
	s.mu.Lock()
	for ipStr, c := range s.clients {
		if c.deviceID == deviceID {
			log.Printf("Device %s reconnecting, removing old session %s", deviceID, ipStr)
			delete(s.addrMap, c.udpAddr.String())
			delete(s.clients, ipStr)
			s.releaseIP(c.vpnIP)
			go s.notifyDisconnected(c.deviceID, ipStr)
			break
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
		lastSeen: time.Now(),
	}

	s.mu.Lock()
	s.clients[ipStr] = c
	s.addrMap[addr.String()] = ipStr
	s.mu.Unlock()

	// Send AUTH_OK with assigned IP
	resp := make([]byte, 5)
	resp[0] = TypeAuthOK
	copy(resp[1:5], ip.To4())
	s.udpConn.WriteToUDP(resp, addr)

	log.Printf("AUTH_OK: device=%s assigned ip=%s", deviceID, ipStr)

	// Notify API
	go s.notifyConnected(deviceID, ipStr)
}

func (s *tunnelServer) handleData(data []byte, addr *net.UDPAddr) {
	if len(data) < 20 { // minimum IP header
		return
	}

	s.mu.RLock()
	ipStr, ok := s.addrMap[addr.String()]
	if ok {
		if c, exists := s.clients[ipStr]; exists {
			c.lastSeen = time.Now()
		}
	}
	s.mu.RUnlock()

	if !ok {
		return
	}

	// Write raw IP packet to TUN
	s.tunIface.Write(data)
}

func (s *tunnelServer) handlePing(addr *net.UDPAddr) {
	s.mu.RLock()
	ipStr, ok := s.addrMap[addr.String()]
	if ok {
		if c, exists := s.clients[ipStr]; exists {
			c.lastSeen = time.Now()
		}
	}
	s.mu.RUnlock()

	// Send PONG
	s.udpConn.WriteToUDP([]byte{TypePong}, addr)
}

func (s *tunnelServer) tunToUdp() {
	buf := make([]byte, maxPacketSize)
	for {
		n, err := s.tunIface.Read(buf)
		if err != nil {
			log.Printf("TUN read error: %v", err)
			continue
		}
		if n < 20 {
			continue
		}

		// Extract destination IP from IP header (bytes 16-19)
		dstIP := net.IPv4(buf[16], buf[17], buf[18], buf[19]).String()

		s.mu.RLock()
		c, ok := s.clients[dstIP]
		s.mu.RUnlock()

		if !ok {
			continue
		}

		// Send DATA packet: [0x02][raw IP packet]
		pkt := make([]byte, 1+n)
		pkt[0] = TypeData
		copy(pkt[1:], buf[:n])
		s.udpConn.WriteToUDP(pkt, c.udpAddr)
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
			if now.Sub(c.lastSeen) > keepaliveTimeout {
				toRemove = append(toRemove, ipStr)
			}
		}
		s.mu.RUnlock()

		for _, ipStr := range toRemove {
			s.mu.Lock()
			c, ok := s.clients[ipStr]
			if ok {
				log.Printf("Client timeout: device=%s ip=%s (idle %v)",
					c.deviceID, ipStr, now.Sub(c.lastSeen))
				delete(s.addrMap, c.udpAddr.String())
				delete(s.clients, ipStr)
				s.releaseIP(c.vpnIP)
				go s.notifyDisconnected(c.deviceID, ipStr)
			}
			s.mu.Unlock()
		}
	}
}

func (s *tunnelServer) notifyConnected(deviceID, vpnIP string) {
	url := s.apiURL + "/api/internal/vpn/connected"
	body := fmt.Sprintf(`{"device_id":"%s","vpn_ip":"%s"}`, deviceID, vpnIP)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("Failed to notify connected for %s: %v", deviceID, err)
		return
	}
	resp.Body.Close()
	log.Printf("Notified API: device %s connected with VPN IP %s (status=%d)", deviceID, vpnIP, resp.StatusCode)
}

func (s *tunnelServer) notifyDisconnected(deviceID, vpnIP string) {
	url := s.apiURL + "/api/internal/vpn/disconnected"
	body := fmt.Sprintf(`{"device_id":"%s","vpn_ip":"%s"}`, deviceID, vpnIP)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Printf("Failed to notify disconnected for %s: %v", deviceID, err)
		return
	}
	resp.Body.Close()
	log.Printf("Notified API: device %s disconnected (status=%d)", deviceID, resp.StatusCode)
}

func (s *tunnelServer) sendAuthFail(addr *net.UDPAddr) {
	s.udpConn.WriteToUDP([]byte{TypeAuthFail}, addr)
}
