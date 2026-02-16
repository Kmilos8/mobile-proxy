package service

import (
	"fmt"
	"log"
	"os/exec"
)

// IPTablesService manages DNAT rules for port forwarding from relay server to VPN clients
type IPTablesService struct{}

func NewIPTablesService() *IPTablesService {
	return &IPTablesService{}
}

// AddDNAT creates a DNAT rule: external_port -> vpn_ip:device_port
func (s *IPTablesService) AddDNAT(externalPort int, vpnIP string, devicePort int) error {
	// iptables -t nat -A PREROUTING -p tcp --dport <ext_port> -j DNAT --to-destination <vpn_ip>:<dev_port>
	rule := fmt.Sprintf("-t nat -A PREROUTING -p tcp --dport %d -j DNAT --to-destination %s:%d",
		externalPort, vpnIP, devicePort)
	if err := s.run(rule); err != nil {
		return fmt.Errorf("add DNAT tcp: %w", err)
	}

	// Also for UDP (SOCKS5 UDP relay)
	ruleUDP := fmt.Sprintf("-t nat -A PREROUTING -p udp --dport %d -j DNAT --to-destination %s:%d",
		externalPort, vpnIP, devicePort)
	if err := s.run(ruleUDP); err != nil {
		// Not critical if UDP fails
		log.Printf("Warning: UDP DNAT failed for port %d: %v", externalPort, err)
	}

	return nil
}

// RemoveDNAT removes a DNAT rule
func (s *IPTablesService) RemoveDNAT(externalPort int, vpnIP string, devicePort int) error {
	rule := fmt.Sprintf("-t nat -D PREROUTING -p tcp --dport %d -j DNAT --to-destination %s:%d",
		externalPort, vpnIP, devicePort)
	if err := s.run(rule); err != nil {
		return fmt.Errorf("remove DNAT tcp: %w", err)
	}

	ruleUDP := fmt.Sprintf("-t nat -D PREROUTING -p udp --dport %d -j DNAT --to-destination %s:%d",
		externalPort, vpnIP, devicePort)
	_ = s.run(ruleUDP) // best effort

	return nil
}

// SetupForDevice creates all 4 DNAT rules for a device
func (s *IPTablesService) SetupForDevice(basePort int, vpnIP string) error {
	// HTTP proxy: basePort -> vpnIP:8080
	if err := s.AddDNAT(basePort, vpnIP, 8080); err != nil {
		return err
	}
	// SOCKS5: basePort+1 -> vpnIP:1080
	if err := s.AddDNAT(basePort+1, vpnIP, 1080); err != nil {
		return err
	}
	// UDP relay: basePort+2 -> vpnIP:1081
	if err := s.AddDNAT(basePort+2, vpnIP, 1081); err != nil {
		return err
	}
	// OpenVPN: basePort+3 -> vpnIP:1194
	if err := s.AddDNAT(basePort+3, vpnIP, 1194); err != nil {
		return err
	}
	return nil
}

// TeardownForDevice removes all 4 DNAT rules for a device
func (s *IPTablesService) TeardownForDevice(basePort int, vpnIP string) error {
	_ = s.RemoveDNAT(basePort, vpnIP, 8080)
	_ = s.RemoveDNAT(basePort+1, vpnIP, 1080)
	_ = s.RemoveDNAT(basePort+2, vpnIP, 1081)
	_ = s.RemoveDNAT(basePort+3, vpnIP, 1194)
	return nil
}

func (s *IPTablesService) run(args string) error {
	cmd := exec.Command("iptables", splitArgs(args)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iptables %s: %s: %w", args, string(output), err)
	}
	return nil
}

func splitArgs(s string) []string {
	var args []string
	current := ""
	for _, ch := range s {
		if ch == ' ' {
			if current != "" {
				args = append(args, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		args = append(args, current)
	}
	return args
}
