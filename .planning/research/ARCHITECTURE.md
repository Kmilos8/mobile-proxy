# Architecture Research

**Domain:** Mobile proxy platform (Android device fleet as proxy endpoints)
**Researched:** 2026-02-25
**Confidence:** HIGH (based on direct codebase analysis + verified external research)

## Standard Architecture

### System Overview

```
                        INTERNET
                           |
              +------------+------------+
              |         Nginx           |  (ports 80/443 → 8080 API, 3000 dash)
              +------------+------------+
                           |
              +------------+------------+
              |      Go API Server      |  (port 8080)
              |     (Gin framework)     |
              +----+--------+------+---+
                   |        |      |
          +--------+  +-----+  +--+--------+
          |  Postgres | | Worker|  |  Dash   |
          |  (state)  | | (bg)  |  | Next.js |
          +-----+-----+ +---+---+  +---------+
                |            |
        +-------+-------+    |
        |  Tunnel Server |   | (heartbeat/WS)
        |  (UDP/TUN)     |   |
        +-------+-------+   |
                |            |
       +--------+--------+   |
       |  iptables DNAT  |   |
       |  iptables REDIR |   |
       +--------+--------+   |
                |            |
     +----------+---------+  |
     | Android Device(s)  |--+
     |  HTTP :8080        |
     |  SOCKS5 :1080      |
     |  UDP relay :1081   |
     +--------------------+

    External Customer (HTTP/SOCKS5):
    Customer → Server IP:port → DNAT → Android Device VPN IP:port → Mobile internet

    External Customer (OpenVPN):
    Customer → OpenVPN server :1195 → iptables REDIRECT → Transparent proxy :12345
            → HTTP CONNECT → Android Device VPN IP:8080 → Mobile internet
```

### Two Parallel VPN Networks

The platform runs two distinct VPN layers simultaneously. Understanding the separation is critical for every feature decision:

**Network 1 — Device Tunnel (port 1194 / 192.168.255.0/24)**

Custom UDP-based tunnel. Android devices authenticate and connect to the Go tunnel server (not OpenVPN). The tunnel server creates a TUN interface (`tun0`), maintains a UDP state machine, and assigns each device an IP in `192.168.255.0/24`. All device-level features flow through this path.

**Network 2 — Customer OpenVPN (port 1195 / 10.9.0.0/24)**

Standard OpenVPN (kylemanna/openvpn image), client-server mode. End-user customers download a `.ovpn` config and connect. The server pushes `redirect-gateway def1`, meaning all customer traffic enters through this tunnel. Auth is via username/password mapped to `ProxyConnection` records. Authenticated with `auth-user-pass-verify` script calling the API.

### Component Responsibilities

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| **Nginx** | TLS termination, reverse proxy routing | API (8080), Dashboard (3000) |
| **Go API Server** | Business logic, REST API, WebSocket hub, JWT auth | Postgres, Tunnel server (8081), OpenVPN scripts |
| **Go Tunnel Server** | Device UDP tunnel, TUN interface, iptables management, transparent proxy | API (8080), Android devices (443 UDP+TCP) |
| **OpenVPN (client-server)** | End-user customer VPN access, auth delegation | API (/api/internal/openvpn/*) via scripts |
| **Transparent Proxy** | Intercept OpenVPN client TCP traffic, forward via HTTP CONNECT to device | iptables REDIRECT from 10.9.0.0/24 |
| **Go Worker** | Background jobs (auto-rotate, stale device cleanup, bandwidth rollups) | Postgres, API services |
| **Next.js Dashboard** | Admin UI — device monitoring, proxy management, pairing | API (REST + WebSocket) |
| **Android App** | Device-side HTTP/SOCKS5 proxy server, heartbeat, command execution | API (heartbeat, commands), Tunnel server (UDP) |
| **PostgreSQL** | Persistent state — devices, connections, customers, ports, commands | API, Worker |

## Recommended Project Structure

The structure already exists. Document it here for reference and build-order reasoning:

```
server/
├── cmd/
│   ├── api/          # API server entrypoint
│   ├── tunnel/       # Tunnel server entrypoint (main.go ~980 lines — key system)
│   └── worker/       # Background job runner
├── internal/
│   ├── api/
│   │   ├── handler/  # Gin route handlers per domain
│   │   └── middleware/
│   ├── domain/       # Shared types, models (single models.go — all domain structs)
│   ├── repository/   # Database access per entity
│   ├── service/      # Business logic per domain
│   │   ├── connection_service.go  # Port allocation, DNAT coordination
│   │   ├── device_service.go      # Heartbeat, commands, status logs, auto-rotate
│   │   ├── iptables_service.go    # DNAT/REDIRECT rule management (API side)
│   │   ├── vpn_service.go         # VPN config generation, IP assignment
│   │   └── ...
│   ├── transparentproxy/  # TCP interception engine (handles OpenVPN clients)
│   └── vpn/           # VPN-specific helpers
├── deployments/
│   ├── docker/        # Dockerfiles per service
│   ├── nginx/         # nginx.conf
│   └── openvpn/       # OpenVPN server configs + hook scripts
└── migrations/        # Sequential SQL migrations (009 migrations as of now)

dashboard/
└── src/
    └── app/           # Next.js App Router pages
        ├── devices/   # Device list + detail
        ├── connections/
        ├── customers/
        ├── overview/
        └── login/

android/
└── app/               # Android proxy app (no root required)
```

## Architectural Patterns

### Pattern 1: Two-Phase Connection Setup (Device Connect)

**What:** When an Android device connects to the tunnel server, a two-phase handshake fires automatically. Phase 1: device authenticates via UDP (or TCP fallback for Samsung devices with NetFilter blocking UDP auth). Phase 2: tunnel notifies API (`POST /api/internal/vpn/connected`), API responds with base port + connection list, tunnel sets up iptables DNAT rules.

**When to use:** This is the only path for device registration. Any feature requiring device state must operate within this flow.

**Trade-offs:** The dual-phase creates a coordination dependency between tunnel and API. If the API is slow or down, DNAT rules don't get installed and the device's proxy ports are unreachable. The current design handles this gracefully — devices can reconnect and re-trigger the notification.

**Example flow:**
```
Android UDP connect → Tunnel assigns 192.168.255.x
→ Tunnel POST /api/internal/vpn/connected {device_id, vpn_ip}
← API returns {base_port: 30000, connections: [{port:30000, type:"http"}, ...]}
→ Tunnel runs iptables DNAT: 30000→192.168.255.x:8080, 30001→:1080, 30002→:1081
→ Proxy ports are now reachable from outside
```

### Pattern 2: OpenVPN Client Traffic Interception (Transparent Proxy Chain)

**What:** End-user customers who connect via OpenVPN get assigned a `10.9.0.x` IP. Their traffic hits the OpenVPN server (`tun1`), routes into the server's kernel. iptables REDIRECT intercepts all TCP from `10.9.0.x` → port 12345 (the transparent proxy). The transparent proxy peeks at the first bytes: TLS? Extract SNI. HTTP? Extract Host header. Then issues HTTP CONNECT to the target device's `192.168.255.y:8080`. The device's HTTP proxy (running on the Android) forwards traffic through mobile internet.

**When to use:** This is the only way to chain OpenVPN client access to mobile device proxies without client-side configuration changes.

**Trade-offs:** Two extra hops add latency. SNI extraction adds ~2ms peek timeout. HTTP CONNECT overhead on first connection per destination is noticeable. The mobile device's HTTP proxy becomes a hard bottleneck — everything funnels through port 8080 on Android. This is the root cause of the current throughput issue.

**QUIC blocking:** iptables REJECTs UDP/443 from `10.9.0.0/24` (ICMP unreachable, not DROP) to force browsers to fall back to TCP immediately rather than waiting for QUIC timeout.

**Example data path:**
```
Customer browser (Windows) → OpenVPN tun1 → 10.9.0.5
→ iptables REDIRECT :12345 (transparent proxy, port 12345)
→ tproxy peeks bytes → TLS → extracts SNI "google.com"
→ HTTP CONNECT google.com:443 → 192.168.255.3:8080 (device proxy)
→ Android HTTP proxy → google.com via 4G (cellular IP appears to destination)
← response flows back same path in reverse
```

### Pattern 3: Command Delivery with Dual-Path Fallback

**What:** Dashboard sends command → API creates `DeviceCommand` record → immediately tries HTTP POST to tunnel push API (port 8081) → tunnel server sends UDP packet `[0x05][json]` to device. If device is offline or push fails, the command stays pending and is delivered on next heartbeat.

**When to use:** All control-plane operations (IP rotation, WiFi toggle, airplane mode, config update) use this. The dual-path ensures commands survive device reconnects.

**Trade-offs:** Heartbeat polling as fallback means offline devices get commands up to the heartbeat interval (currently ~30-60s) after reconnecting. Push API requires tunnel server to be reachable from API container — handled by `host.docker.internal`.

### Pattern 4: Port Allocation via Sequential Counter

**What:** Each device gets a base port at registration time (starting at 30000, +4 per device). Ports are `basePort` (HTTP), `basePort+1` (SOCKS5), `basePort+2` (UDP relay), `basePort+3` (OVPN). Per-connection proxy ports are also allocated from a shared pool. Port state lives in PostgreSQL.

**When to use:** Any feature touching proxy port assignment must go through `PortService.AllocatePort()`.

**Trade-offs:** Port fragmentation can occur if devices are deleted and re-registered. No port recycling currently — LOW priority problem for single-tenant deployments.

## Data Flow

### Request Flow — HTTP/SOCKS5 Customer Proxy

```
External client (curl/browser configured for proxy)
    → Server IP:30000 (TCP/UDP)
    → iptables DNAT: 30000 → 192.168.255.3:8080
    → Android HTTP proxy (no auth if device-level, or credentials if per-connection)
    → Mobile carrier NAT → destination website
    ← response path reverses
```

### Request Flow — OpenVPN Customer

```
Customer OpenVPN client (desktop/mobile)
    → Server UDP:1195 (OpenVPN client-server)
    → auth-user-pass-verify.sh → POST /api/internal/openvpn/auth
    ← API validates ProxyConnection username/password
    → OpenVPN assigns 10.9.0.x, runs client-connect-ovpn.sh
    → POST /api/internal/openvpn/connect {username, vpn_ip}
    → API looks up device VPN IP, POSTs /openvpn-client-connect to tunnel (port 8081)
    → Tunnel adds tproxy mapping: 10.9.0.x → 192.168.255.y:8080
    → Tunnel adds iptables: REDIRECT 10.9.0.x TCP → :12345
    → Customer traffic flows through VPN
    → iptables REDIRECT intercepts, transparent proxy handles
    → SNI/Host extraction → HTTP CONNECT → device → mobile internet
```

### Device Registration Flow

```
Android App first launch
    → POST /api/public/pair {code} (QR code scan from dashboard)
    ← API returns {device_id, auth_token, vpn_config, base_port, relay_server_ip}
    → App stores auth_token, connects UDP tunnel to relay_server_ip:443
    → Tunnel AUTH handshake → assigns 192.168.255.x
    → Tunnel notifies API → iptables DNAT setup
    → App starts HTTP proxy (:8080) and SOCKS5 proxy (:1080)
    → App begins 30s heartbeat loop → POST /api/devices/:id/heartbeat
    ← API delivers pending commands in heartbeat response
```

### Dashboard State Management

```
Next.js Dashboard
    → REST calls to GET /api/devices, /api/connections, /api/customers
    → WebSocket connection to /ws for real-time device status updates
    ← Go WSHub broadcasts device_connected, device_disconnected, heartbeat events
    → Dashboard renders device cards, proxy port tables, status indicators
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-50 devices | Current single-server Docker Compose — perfectly adequate |
| 50-500 devices | Multiple relay servers (RelayServer table already supports this). Each relay has its own tunnel server instance. API routes commands to correct relay via `relay_server_ip`. Already scaffolded. |
| 500-5000 devices | Horizontal API scaling behind Nginx load balancer (stateless API). Postgres read replicas for device list queries. Redis for WebSocket broadcast fanout. |
| 5000+ devices | Microservice split: separate device management, proxy management, billing. Kafka/NATS for event bus. Multiple regional relay clusters. |

### Scaling Priorities

1. **First bottleneck — Android device HTTP proxy throughput:** The `8080` port on Android is the hard limit for OpenVPN customer traffic. A single device's cellular connection (typically 20-100 Mbps) shared across all customers on that device. Mitigation: limit concurrent OpenVPN connections per device (not yet implemented).

2. **Second bottleneck — Tunnel server iptables rules:** Each connected device adds DNAT rules. Each OpenVPN customer connection adds REDIRECT rules. At 500+ concurrent connections, iptables rule traversal latency becomes measurable. Mitigation: `nftables` migration (better performance, atomic rule sets), or `ipset` for bulk IP matching.

3. **Third bottleneck — WebSocket broadcast:** Single Go WSHub broadcasts to all dashboard sessions. At 1000+ connected dashboard tabs, this is fine. Beyond that, use Redis pub/sub for cross-instance broadcast.

## Anti-Patterns

### Anti-Pattern 1: Routing Android Device Traffic Through the Server

**What people do:** Configure OpenVPN (device-side, port 1194) with `push "redirect-gateway"` to send device traffic through the server.

**Why it's wrong:** The entire point is that customer traffic exits via the phone's cellular IP. If device traffic is redirected through the server, all customer traffic would exit via the server's IP — defeating the product entirely. The device VPN (port 1194) must NOT push redirect-gateway.

**Do this instead:** The `server.conf` (port 1194) correctly does NOT push redirect-gateway. The `client-server.conf` (port 1195, for customers) DOES push redirect-gateway. These two OpenVPN instances must remain separate.

### Anti-Pattern 2: iptables Rules Without Idempotent Teardown

**What people do:** Append DNAT/REDIRECT rules on connect without removing existing rules first.

**Why it's wrong:** If a device disconnects and reconnects (or the tunnel server restarts), rules accumulate. iptables traverses ALL matching rules — duplicate rules cause the first match to win but the duplicates consume CPU. This was a real bug fixed in recent commits (teardown before setup, loop to remove all duplicates).

**Do this instead:** Always call teardown before setup. Always loop `iptables -D` until it fails to confirm all duplicates are gone. The current code in `setupDNAT()` and `teardownDNAT()` does this correctly.

### Anti-Pattern 3: Blocking iptables REDIRECT for the Transparent Proxy With DROP Instead of REJECT for QUIC

**What people do:** `iptables -A FORWARD -s 10.9.0.0/24 -p udp --dport 443 -j DROP` to block QUIC.

**Why it's wrong:** DROP causes the client to wait for a timeout before falling back to TCP. QUIC timeout is typically 3+ seconds, causing browsers to appear slow/frozen before TCP kicks in.

**Do this instead:** Use `REJECT --reject-with icmp-port-unreachable`. This sends an immediate ICMP unreachable, causing browsers to fall back to TCP in under 100ms. Current code does this correctly.

### Anti-Pattern 4: Sharing One OpenVPN Port for Both Devices and Customers

**What people do:** Use a single OpenVPN server for both Android devices (reverse tunnel) and end-user customers.

**Why it's wrong:** The two use-cases have fundamentally different requirements. Device tunnel must NOT redirect gateway (or devices lose internet). Customer tunnel MUST redirect gateway (so traffic exits via mobile). They need different subnets, different auth systems, different connect hooks.

**Do this instead:** Run two separate OpenVPN instances (current approach: port 1194 for devices via custom tunnel, port 1195 for customers via OpenVPN client-server with `verify-client-cert none`).

### Anti-Pattern 5: Proxying Customer HTTPS Traffic Via IP Address Instead of Domain

**What people do:** Issue HTTP CONNECT to the raw destination IP (from iptables SO_ORIGINAL_DST).

**Why it's wrong:** Many HTTP proxies on Android reject CONNECT requests with bare IP addresses. Mobile carrier proxies and Android's built-in proxy may require domain names for CONNECT. More importantly, many HTTPS servers enforce SNI-based virtual hosting — CONNECT to IP misses the SNI handshake context.

**Do this instead:** Peek at the first bytes, extract SNI (for TLS) or Host header (for HTTP), and issue CONNECT with the domain name. Fall back to IP if extraction fails. The current `transparentproxy/proxy.go` does this correctly.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| OpenVPN (kylemanna image) | Shell scripts (`client-connect.sh`, `auth-verify.sh`) call API via `wget` | Scripts run inside OpenVPN container, call `http://127.0.0.1:8080/api` |
| Android App | REST heartbeat + custom UDP tunnel protocol | Device token auth (currently open for MVP) |
| PostgreSQL | Direct SQL via `pgx` driver | 9 migration files, sequential numbered |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| API ↔ Tunnel Server | HTTP POST to port 8081 (push API) | Tunnel runs `network_mode: host`, API uses `host.docker.internal:8081` |
| API ↔ OpenVPN hooks | OpenVPN shell scripts → `wget` → API HTTP | No auth on internal routes (VPN container is trusted) |
| Tunnel Server ↔ Android | Custom binary UDP protocol (type byte + payload) | Auth uses device UUID bytes, data is raw IP packets |
| Transparent Proxy ↔ Device Proxy | HTTP CONNECT over TCP to `192.168.255.x:8080` | Device must be connected to tunnel for VPN IP to exist |
| Dashboard ↔ API | REST (JWT) + WebSocket (`/ws`) | WebSocket hub in API server, no separate WS server needed |

## Build Order Implications

The component dependency graph dictates what must be working before the next thing can be built or tested:

```
1. PostgreSQL (state store — everything else depends on it)
        ↓
2. Go API Server (business logic — tunnel + dashboard depend on it)
        ↓
3. Go Tunnel Server (device connectivity — proxy traffic depends on it)
        ↓
4. iptables DNAT rules (set up BY tunnel server on device connect — proxy ports depend on this)
        ↓
5. Android App connecting (device must be online for proxy ports to work)
        ↓
6. HTTP/SOCKS5 proxy traffic (works once DNAT is in place)
        ↓
7. OpenVPN client-server (depends on API for auth, tunnel for transparent proxy mapping)
        ↓
8. Transparent proxy (REDIRECT + tproxy mapping — depends on tunnel server running)
        ↓
9. OpenVPN customer traffic (full chain works once all above are operational)
        ↓
10. Next.js Dashboard (reads from API — works whenever API is up, no data plane dependency)
```

**Key insight:** The OpenVPN customer path (steps 7-9) is the most complex because it depends on ALL prior components AND adds two additional interception layers (iptables REDIRECT + transparent proxy). Any throughput problem in this path can originate at any of: OpenVPN encryption, iptables rule traversal, transparent proxy peeks, HTTP CONNECT handshake, Android HTTP proxy, or mobile network. Debugging requires isolating each layer.

**Dashboard build order:** The dashboard has no data-plane dependency — it can be built and refined at any time as long as the API is running. Dashboard work does NOT need to wait for proxy protocol stability.

## Sources

- Direct codebase analysis: `server/cmd/tunnel/main.go`, `server/internal/transparentproxy/proxy.go`, `server/internal/service/`, `server/deployments/openvpn/`
- [iProxy traffic path explanation](https://iproxy.online/blog/traffic-path-from-android-to-program-with-proxy) (MEDIUM confidence — competitor reference for pattern validation)
- [OpenVPN transparent proxy forum](https://forums.openvpn.net/viewtopic.php?t=32422) (MEDIUM confidence — community-verified iptables REDIRECT approach)
- [Go SNI proxy patterns](https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go) (HIGH confidence — official Go implementation reference)
- [OpenVPN throughput optimization](https://hamy.io/post/0003/optimizing-openvpn-throughput/) (MEDIUM confidence — practical guide, multiple sources agree)
- [BBR congestion control](https://lwn.net/Articles/776090/) (HIGH confidence — kernel documentation)
- [Next.js WebSocket architecture](https://www.wisp.blog/blog/choosing-the-right-architecture-for-socket-communication-in-nextjs-a-comprehensive-guide) (MEDIUM confidence — matches current implementation)

---
*Architecture research for: PocketProxy — mobile proxy platform*
*Researched: 2026-02-25*
