package domain

import "fmt"

type Config struct {
	Server    ServerConfig      `json:"server"`
	Database  DatabaseConfig    `json:"database"`
	JWT       JWTConfig         `json:"jwt"`
	VPN       VPNConfig         `json:"vpn"`
	Ports     PortConfig        `json:"ports"`
	Google    GoogleOAuthConfig `json:"google"`
	Resend    ResendConfig      `json:"resend"`
	Turnstile TurnstileConfig   `json:"turnstile"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

func (c DatabaseConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" +
		fmt.Sprintf("%d", c.Port) + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

type JWTConfig struct {
	Secret      string `json:"secret"`
	ExpireHours int    `json:"expire_hours"`
}

type VPNConfig struct {
	ServerIP   string `json:"server_ip"`
	Subnet     string `json:"subnet"`      // 10.8.0.0
	SubnetMask string `json:"subnet_mask"` // 255.255.255.0
	CCDDir     string `json:"ccd_dir"`     // client-config-dir path
	ConfigDir  string `json:"config_dir"`
	CAPath     string `json:"ca_path"`
	ServerCert string `json:"server_cert"`
	ServerKey  string `json:"server_key"`
}

type PortConfig struct {
	BasePort       int `json:"base_port"`        // 30000
	MaxPort        int `json:"max_port"`         // 39999
	PortsPerDevice int `json:"ports_per_device"` // 4
}

type GoogleOAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

type ResendConfig struct {
	APIKey    string `json:"api_key"`
	FromEmail string `json:"from_email"`
	BaseURL   string `json:"base_url"` // dashboard base URL for links in emails
}

type TurnstileConfig struct {
	SiteKey   string `json:"site_key"`
	SecretKey string `json:"secret_key"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
		Database: DatabaseConfig{
			Host: "localhost", Port: 5432,
			User: "mobileproxy", Password: "mobileproxy",
			DBName: "mobileproxy", SSLMode: "disable",
		},
		JWT: JWTConfig{Secret: "change-me-in-production", ExpireHours: 24},
		VPN: VPNConfig{
			ServerIP: "0.0.0.0", Subnet: "10.8.0.0", SubnetMask: "255.255.255.0",
			CCDDir: "/etc/openvpn/ccd", ConfigDir: "/etc/openvpn",
		},
		Ports: PortConfig{BasePort: 30000, MaxPort: 39999, PortsPerDevice: 4},
		Google: GoogleOAuthConfig{
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "",
		},
		Resend: ResendConfig{
			APIKey:    "",
			FromEmail: "",
			BaseURL:   "",
		},
		Turnstile: TurnstileConfig{
			SiteKey:   "",
			SecretKey: "",
		},
	}
}
