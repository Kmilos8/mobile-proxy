# MobileProxy

Turn Android phones with SIM cards into HTTP/SOCKS5 proxy servers with mobile IPs.

## How It Works

The core mechanism is **WiFi Split** - each Android device maintains two network connections simultaneously:

- **WiFi** carries the VPN tunnel to the relay server (fast, reliable, no cellular data cost)
- **Cellular** carries all proxy traffic (provides the mobile IP that end customers see)

```
Customer                        Relay Server (VPS)                  Android Device
                                ┌──────────────────────┐            ┌─────────────────┐
HTTP/SOCKS5 ──────────────────► │ Port 30000-39999     │            │                 │
proxy request                   │   iptables DNAT ──────── OpenVPN ─── Proxy Server   │
                                │                      │   tunnel   │   HTTP :8080    │
                                │ OpenVPN Server       │   (WiFi)   │   SOCKS5 :1080  │
                                │   10.8.0.0/24        │            │                 │
                                │                      │            │ Outbound via    │
                                │ Go API :8080         │            │   CELLULAR      │
                                │ Dashboard :3000      │            │   → Mobile IP   │
                                └──────────────────────┘            └─────────────────┘
```

Each device gets 4 ports on the relay server (base range 30000-39999):

| Offset | Protocol | Description |
|--------|----------|-------------|
| +0 | HTTP | HTTP CONNECT proxy |
| +1 | SOCKS5 | SOCKS5 proxy (TCP) |
| +2 | SOCKS5 UDP | SOCKS5 UDP relay |
| +3 | OpenVPN | Customer VPN access |

## Components

### Android App (Kotlin)

- **WiFi Split** via `ConnectivityManager.requestNetwork()` for both TRANSPORT_CELLULAR and TRANSPORT_WIFI
- **HTTP CONNECT proxy** with cellular-bound outbound sockets
- **SOCKS5 proxy** (RFC 1928) with IPv4, IPv6, and domain support
- **IP rotation** via cellular reconnect or Accessibility Service airplane mode toggle
- **Foreground service** with wake lock for persistent operation
- **Heartbeat reporting** (battery, signal, carrier, bandwidth) every 30s

### Go Backend (Gin)

- REST API for device management, proxy connections, customers
- JWT authentication
- Port allocation (4 ports per device, 30000-39999)
- iptables DNAT rule management for port forwarding through VPN
- OpenVPN CCD file management for static VPN IP assignment
- WebSocket for real-time dashboard updates
- Background worker for stale device detection and partition maintenance

### Next.js Dashboard

- Device list with real-time status (WebSocket + polling fallback)
- Proxy connection CRUD (create credentials, assign to devices)
- Customer management
- IP rotation controls per device
- Battery, signal strength, carrier, and bandwidth display

### Infrastructure

- **PostgreSQL 16** with INET types, partitioned bandwidth logs
- **OpenVPN** server (no `redirect-gateway` - only VPN subnet routed)
- **Nginx** reverse proxy with rate limiting and WebSocket support
- **Docker Compose** orchestration for all services

## Quick Start

### Prerequisites

- Docker and Docker Compose
- A VPS with a public IP (Ubuntu 22.04 recommended)
- Android devices with SIM cards

### 1. Clone and configure

```bash
git clone https://github.com/Kmilos8/mobile-proxy.git
cd mobile-proxy
cp .env.example .env
# Edit .env with your JWT secret and other settings
```

### 2. Initialize OpenVPN PKI

```bash
chmod +x server/deployments/openvpn/setup.sh
./server/deployments/openvpn/setup.sh YOUR_SERVER_PUBLIC_IP
```

### 3. Start services

```bash
docker-compose up -d
```

Services will be available at:
- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **OpenVPN**: UDP port 1194
- **Proxy ports**: 30000-30799

### 4. Login

Default credentials:
- Email: `admin@mobileproxy.local`
- Password: `admin123`

### 5. Generate device certificates

```bash
# For each Android device:
docker run -v openvpn_data:/etc/openvpn --rm -it kylemanna/openvpn \
  easyrsa build-client-full device-001 nopass

docker run -v openvpn_data:/etc/openvpn --rm kylemanna/openvpn \
  ovpn_getclient device-001 > device-001.ovpn
```

### 6. Install Android app

Build the app from `android/` in Android Studio and install on devices. Enter the relay server URL and start the service.

## IP Rotation

Two methods available (devices are not rooted):

1. **Cellular Reconnect** (primary) - Unregisters and re-requests the cellular network via `ConnectivityManager`, causing modem re-registration
2. **Airplane Mode Toggle** (fallback) - Accessibility Service programmatically navigates Quick Settings to toggle airplane mode

Trigger rotation from the dashboard or via API:

```bash
curl -X POST http://localhost:8080/api/devices/{id}/commands \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type": "rotate_ip"}'
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login, returns JWT token

### Devices
- `GET /api/devices` - List all devices
- `GET /api/devices/:id` - Get device details
- `POST /api/devices/register` - Register new device
- `POST /api/devices/:id/heartbeat` - Device heartbeat
- `POST /api/devices/:id/commands` - Send command to device
- `GET /api/devices/:id/ip-history` - Get IP change history

### Proxy Connections
- `GET /api/connections` - List connections
- `POST /api/connections` - Create connection with credentials
- `PATCH /api/connections/:id` - Enable/disable connection
- `DELETE /api/connections/:id` - Delete connection

### Customers
- `GET /api/customers` - List customers
- `POST /api/customers` - Create customer
- `PUT /api/customers/:id` - Update customer

### WebSocket
- `GET /ws` - Real-time device status updates

## Project Structure

```
mobile-proxy/
├── android/                    # Android app (Kotlin)
│   └── app/src/main/java/com/mobileproxy/
│       ├── core/
│       │   ├── network/        # WiFi Split: NetworkManager, CellularSocketFactory
│       │   ├── proxy/          # HTTP and SOCKS5 proxy servers
│       │   ├── rotation/       # IP rotation (cellular reconnect + accessibility)
│       │   ├── commands/       # Remote command execution
│       │   └── status/         # Heartbeat reporting
│       ├── service/            # Foreground service, VPN service, boot receiver
│       └── ui/                 # Setup/status UI
├── server/                     # Go backend
│       ├── cmd/api/            # API server entry point
│       ├── cmd/worker/         # Background worker
│       ├── internal/
│       │   ├── api/            # Handlers, middleware, router
│       │   ├── service/        # Business logic
│       │   ├── repository/     # Database access (pgx)
│       │   └── domain/         # Models and config
│       ├── migrations/         # PostgreSQL migrations
│       └── deployments/        # Docker, nginx, OpenVPN configs
└── dashboard/                  # Next.js 14 web dashboard
    └── src/
        ├── app/                # Pages (devices, connections, customers, login)
        ├── components/         # Sidebar, layout
        └── lib/                # API client, auth, WebSocket, utils
```

## Capacity

- 4 ports per device, range 30000-39999
- **2,500 devices** per relay server
- Current deployment: 200 devices (ports 30000-30799)

## License

Proprietary. All rights reserved.
