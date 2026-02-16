package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mobileproxy/server/internal/domain"
)

// VPNService manages OpenVPN configuration for devices
type VPNService struct {
	config    domain.VPNConfig
	iptables  *IPTablesService
}

func NewVPNService(config domain.VPNConfig, iptables *IPTablesService) *VPNService {
	return &VPNService{
		config:   config,
		iptables: iptables,
	}
}

// AssignVPNIP assigns a static VPN IP to a device via CCD file.
// VPN IPs are assigned based on device index: 10.8.0.2, 10.8.0.3, etc.
func (s *VPNService) AssignVPNIP(deviceName string, deviceIndex int) (string, error) {
	vpnIP := fmt.Sprintf("10.8.0.%d", deviceIndex+2) // +2 because .0=network, .1=server

	// Create CCD file
	ccdPath := filepath.Join(s.config.CCDDir, deviceName)
	content := fmt.Sprintf("ifconfig-push %s 255.255.255.0\n", vpnIP)

	if err := os.MkdirAll(s.config.CCDDir, 0755); err != nil {
		return "", fmt.Errorf("create ccd dir: %w", err)
	}

	if err := os.WriteFile(ccdPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write ccd file: %w", err)
	}

	log.Printf("Assigned VPN IP %s to device %s", vpnIP, deviceName)
	return vpnIP, nil
}

// GenerateClientConfig generates an OpenVPN client config for a device
func (s *VPNService) GenerateClientConfig(serverIP string, deviceName string) string {
	return fmt.Sprintf(`client
dev tun
proto udp
remote %s 1194
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
cipher AES-256-GCM
auth SHA256
verb 3

# IMPORTANT: Do NOT redirect all traffic through VPN
# Only VPN subnet traffic goes through tunnel
# Proxy traffic goes through cellular

<ca>
# CA certificate will be embedded here
</ca>

<cert>
# Client certificate for %s will be embedded here
</cert>

<key>
# Client key for %s will be embedded here
</key>
`, serverIP, deviceName, deviceName)
}

// OnDeviceConnected is called when a device connects to OpenVPN
// Sets up iptables DNAT rules
func (s *VPNService) OnDeviceConnected(basePort int, vpnIP string) error {
	log.Printf("Device connected: vpnIP=%s, basePort=%d", vpnIP, basePort)
	return s.iptables.SetupForDevice(basePort, vpnIP)
}

// OnDeviceDisconnected is called when a device disconnects from OpenVPN
// Removes iptables DNAT rules
func (s *VPNService) OnDeviceDisconnected(basePort int, vpnIP string) error {
	log.Printf("Device disconnected: vpnIP=%s, basePort=%d", vpnIP, basePort)
	return s.iptables.TeardownForDevice(basePort, vpnIP)
}
