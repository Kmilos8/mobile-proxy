# Pitfalls Research

**Domain:** Mobile proxy platform — Android devices as HTTP/SOCKS5/OpenVPN endpoints
**Researched:** 2026-02-25
**Confidence:** HIGH (critical pitfalls drawn from direct code inspection + MEDIUM for market/UX patterns from WebSearch)

---

## Critical Pitfalls

### Pitfall 1: The Transparent Proxy Double-Encapsulation Bottleneck

**What goes wrong:**
The current OpenVPN path is: client -> OpenVPN (UDP, AES encryption) -> server tun1 -> iptables REDIRECT -> transparent proxy (Go, port 12345) -> phone HTTP proxy (over tun0, second TLS/AES layer) -> cellular internet. Every byte passes through two full encryption layers, two kernel-to-userspace copies in Go, and two TCP connection setups. The result is latency multiplication and throughput division. This is the most likely cause of the current "pages barely load" symptom.

**Why it happens:**
The transparent proxy approach was chosen to map OpenVPN clients to device proxies without requiring client-side reconfiguration. It is architecturally correct but introduces a second full proxy hop (HTTP CONNECT over the tunnel to the phone) on top of the OpenVPN layer. Each new TCP connection from the client causes: OpenVPN handshake overhead + transparent proxy accept + HTTP CONNECT to phone + phone connection to internet. For a modern HTTPS page with 20-50 parallel connections, this serializes into a waterfall of setup overhead.

**How to avoid:**
- Measure first: capture tcpdump on tun1 during a page load to verify traffic is reaching the transparent proxy and the phone proxy is responding. Confirm the bottleneck is latency (many slow connections) vs. bandwidth (single slow transfer).
- If latency is the bottleneck: persistent connection pooling in the transparent proxy to the phone's HTTP proxy eliminates per-request CONNECT overhead. Keep a pool of pre-established CONNECT tunnels per client VPN IP.
- If bandwidth is the bottleneck: profile whether it is the OpenVPN encryption CPU, the TUN MTU mismatch causing fragmentation, or the phone's cellular upload capacity.
- The `peekTimeout` of 2 seconds in `transparentproxy/proxy.go` means the proxy waits up to 2 seconds for SNI/Host before forwarding. This alone adds 2 seconds to TLS handshake for connections where Peek() blocks.

**Warning signs:**
- tcpdump on tun1 shows many small TCP SYN packets but very few large data transfers
- OpenVPN logs show normal throughput but browser shows "waiting for server" on every resource
- Speed tests show very low speed but ping is normal
- Transparent proxy logs show high rate of new `[tproxy] added mapping` lines per page load
- Multiple CONNECT attempts per second per user

**Phase to address:** OpenVPN fix phase (immediate priority — current active bug)

---

### Pitfall 2: MTU Mismatch Causing Silent Packet Fragmentation

**What goes wrong:**
The tunnel server sets `tunMTU = 1400` on the device-side TUN (tun0). The OpenVPN client-server config sets `tun-mtu 1500` and `mssfix 1400` on tun1. Packets flowing from the OpenVPN client through tun1 arrive at the transparent proxy at up to 1500 bytes. The transparent proxy then sends these over tun0 (MTU 1400) to the phone. Oversized packets must be fragmented. If the phone's proxy or the cellular network drops fragmented UDP packets (common on mobile carriers), connections silently stall. This manifests as connections that start then stop, or slow downloads that plateau.

**Why it happens:**
MTU is set independently on each layer without accounting for the full encapsulation overhead chain: cellular network MTU (1500) -> OpenVPN header (69 bytes overhead) -> tun1 -> transparent proxy -> HTTP CONNECT overhead -> tun0 (1400) -> cellular MTU on phone side. Each hop shrinks the effective payload. When inner MTU exceeds outer MTU, TCP's Path MTU Discovery (PMTUD) relies on ICMP "fragmentation needed" messages which cellular carriers frequently block (ICMP black holes), so PMTUD silently fails.

**How to avoid:**
- Set `tun-mtu 1420` and `mssfix 1360` on the client-server OpenVPN to give headroom for the second encapsulation hop.
- Verify with: `ping -M do -s 1300 8.8.8.8` through the VPN; if it fails, decrease until it works; that is your effective MTU.
- Enable `tcp-mtu-probing` sysctl (already done in tunnel server) to handle ICMP black holes.
- Push the verified `tun-mtu` and `mssfix` values to clients in the server conf rather than hardcoding on client .ovpn files.

**Warning signs:**
- Downloads start at good speed then drop to near zero after ~64KB (one TCP window's worth)
- `ping -M do -s 1400 <server>` works but browsing fails
- tcpdump shows many TCP retransmissions (RST after initial data)
- Connections work on WiFi but fail on 4G (WiFi has lower effective overhead)

**Phase to address:** OpenVPN fix phase

---

### Pitfall 3: OpenVPN over UDP Blocked by Mobile Carriers

**What goes wrong:**
The device-side tunnel uses UDP port 443 (fallback port 1194 in config). The OpenVPN client-server uses UDP port 1195. Some mobile carriers block non-HTTP UDP traffic, including standard OpenVPN ports. Others use CGNAT that closes UDP "connections" after 30-90 seconds of inactivity, causing silent tunnel drops with no indication in OpenVPN logs until the next keepalive cycle. The `keepalive 10 120` setting means the server waits 120 seconds before declaring a client dead — users experience 2 minutes of dead connection with no error.

**Why it happens:**
Mobile carrier networks (especially in some regions and MVNOs) prioritize TCP/HTTP traffic and aggressively time out UDP flows in their CGNAT tables. OpenVPN's keepalive is designed for stable networks; the 10/120 defaults are too generous for mobile carrier CGNAT timeouts which can be as low as 30 seconds.

**How to avoid:**
- Reduce keepalive to `keepalive 5 30` on client-server to detect drops within 30 seconds.
- Add a TCP fallback: run a second OpenVPN listener on TCP 443 (or use the same port with protocol auto-detection). TCP 443 is almost never blocked.
- For the device-side tunnel (port 443 UDP), this is already a good choice. Verify devices are not reconnecting frequently by monitoring reconnect rate in logs.
- Push `explicit-exit-notify 1` so clients signal clean disconnects rather than timing out.

**Warning signs:**
- Devices show "connected" in dashboard but traffic stops flowing
- Reconnection events spike in logs after ~30-60 second periods of no user activity
- Works on some carrier SIMs but not others
- Users report proxy works then goes dead after a few minutes of browsing

**Phase to address:** OpenVPN fix phase, then carrier compatibility testing during production hardening

---

### Pitfall 4: Plaintext Credentials in .ovpn Files and Database

**What goes wrong:**
The current codebase stores `conn.PasswordPlain` in the database and embeds raw credentials in generated .ovpn files in plaintext (`<auth-user-pass>` block). Any database breach or .ovpn file leak exposes all proxy credentials. The auth verification script passes credentials via environment variables (`via-env`), which are visible in `/proc/<pid>/environ` to any process with sufficient privilege on the host.

**Why it happens:**
Plaintext passwords are simpler to implement and compare. For a proxy platform, the password is typically machine-generated and users rarely see it directly — but the risk is identical to any credential exposure: database dump = all accounts compromised.

**How to avoid:**
- Hash passwords at rest with bcrypt (minimum cost 10). The auth comparison in `openvpn_handler.go` line 56 (`conn.PasswordPlain != req.Password`) becomes `bcrypt.CompareHashAndPassword()`.
- Generate .ovpn files on-demand only, never store them with embedded credentials. The current `DownloadOVPN` handler already generates on-demand — good — but the inline `<auth-user-pass>` block means any intercepted download exposes credentials. Consider a short-lived token approach instead.
- For the OpenVPN `auth-user-pass-verify via-env`, switch to `via-file` which writes to a temp file with restricted permissions rather than environment variables.

**Warning signs:**
- Database migration scripts that export tables include raw passwords in dumps
- .ovpn files downloadable via unauthenticated API endpoint
- `ps aux` shows auth credentials in process environment on host

**Phase to address:** Security hardening phase (before any production deployment)

---

### Pitfall 5: Race Between iptables REDIRECT and Transparent Proxy State

**What goes wrong:**
When an OpenVPN client connects, the flow is: OpenVPN fires `client-connect-ovpn.sh` -> shell script POSTs to API -> API POSTs to tunnel push API -> tunnel server adds mapping AND adds iptables REDIRECT rule. If the iptables rule is added before the mapping is added to `transparentproxy`, or if the mapping POST fails silently (HTTP error ignored with `|| true` in shell script), TCP connections from the OpenVPN client get redirected to port 12345 but `getMapping()` returns no target. The transparent proxy logs "no mapping for client X" and drops the connection. The user sees a connected VPN but zero traffic.

**Why it happens:**
The client-connect script uses `wget ... || true` (line 16 in `client-connect-ovpn.sh`) which masks API failures. The iptables REDIRECT rule and the Go in-memory mapping must be consistent. If the chain fails at any point, the REDIRECT rule is active but no mapping exists. The code in `handleOpenVPNClientConnect()` adds the iptables rule after the mapping, which is the correct order, but if the HTTP POST from OpenVPN script to the API times out, the entire setup may be skipped or partially completed.

**How to avoid:**
- Remove `|| true` from the client-connect script; let failures propagate so OpenVPN can retry or reject the connection cleanly.
- Add idempotent cleanup on connect: always call disconnect handler before connect handler for the same client IP.
- Add a mapping existence check in the transparent proxy with a retry/wait loop (100ms, 3 retries) to handle race conditions between rule activation and mapping registration.
- Log the specific failure reason when `getMapping()` returns false — the current log "no mapping for client X" does not distinguish between "never registered" and "already removed."

**Warning signs:**
- VPN connects (IP assigned, tunnel up) but HTTP requests time out immediately
- Transparent proxy logs show consistent "no mapping" for specific client IPs
- Issue resolves after manual disconnect/reconnect
- API logs show no `[openvpn-connect]` entry after a VPN client connects

**Phase to address:** OpenVPN fix phase

---

## Moderate Pitfalls

### Pitfall 6: Device Offline Shown as Online (Stale Keepalive State)

**What goes wrong:**
The tunnel server uses UDP keepalive with a 60-second timeout. An Android device that loses its cellular connection (switches towers, enters a no-signal area) will disappear from the tunnel without sending a disconnect. The server marks it offline after 60 seconds. During those 60 seconds, new OpenVPN client connections can be assigned to this device, sending traffic to a VPN IP with no route to a live device. HTTP/SOCKS5 ports remain open (DNAT rules persist) and connections time out silently.

**How to avoid:**
- Reduce the `keepaliveTimeout` from 60 seconds to 20-30 seconds in tunnel/main.go for faster detection.
- Surface the "last seen" timestamp in the dashboard with a visual age indicator (e.g., green = < 15s, yellow = 15-45s, red = > 45s).
- The `notifyDisconnected()` function already tears down DNAT rules — verify it also updates `device.VpnIP = ""` in the database so new connection assignments check device liveness.
- In `openvpn_handler.go` line 96, the check `device.VpnIP == ""` is the right gate — but it only works if the VPN IP is cleared promptly on device disconnect.

**Warning signs:**
- Dashboard shows device as online but HTTP proxy port returns connection refused
- Users report working proxy that goes dead without warning
- Device reconnect logs show frequent AUTH requests from the same device (signal bouncing)

**Phase to address:** Device management / dashboard phase

---

### Pitfall 7: iptables Rules Surviving Container Restarts

**What goes wrong:**
The tunnel container uses `network_mode: host` and `privileged: true`, which means its iptables rules persist in the host kernel across container restarts. If the tunnel container crashes and restarts, `configureTUN()` re-adds MASQUERADE and FORWARD rules. The existing teardown-before-add logic (check-then-add) prevents most duplicates, but if state gets inconsistent (e.g., half-teardown during OOM kill), orphan rules accumulate. DNAT rules for disconnected devices remain in PREROUTING until the loop cleans them, which never runs if the process was killed.

**How to avoid:**
- On startup, flush all PREROUTING rules in the nat table for the OpenVPN client subnet before setting up fresh rules: `iptables -t nat -F PREROUTING` is too aggressive but targeted flush of `10.9.0.0/24` source rules is safe.
- Track applied DNAT rules in a persistent file (e.g., `/tmp/tunnel-dnat-state.json`) so restart can clean up rules from the previous run.
- Add a health check endpoint to the tunnel push API that lists active mappings and iptables state — useful for debugging production issues.

**Warning signs:**
- `iptables -t nat -L PREROUTING -n` shows duplicate DNAT entries for the same port
- After container restart, some proxy ports work but others don't (orphan rules from previous session conflicting with new assignments)
- `iptables -t nat -L PREROUTING -n | wc -l` grows unboundedly over time

**Phase to address:** Stability / production hardening phase

---

### Pitfall 8: Single Go Goroutine for TUN-to-UDP (tunToUdp)

**What goes wrong:**
The `tunToUdp()` function in tunnel/main.go is a single goroutine with a pre-allocated buffer. It uses the same buffer for every packet with no locking (`// Pre-allocated send buffer for tunToUdp (single goroutine, no lock needed)` comment). This is correct for a single goroutine, but it means all downstream traffic to all connected devices serializes through one goroutine. With 10+ connected devices, this becomes a CPU bottleneck at high throughput. Any blocking `WriteToUDP()` stalls all other devices.

**How to avoid:**
- For the current scale (< 50 devices), this is likely not the bottleneck. Measure before optimizing.
- When it does become a bottleneck: use a per-client outbound channel with a worker goroutine per client. This allows parallel writes to different UDP addresses.
- A simpler first step: make `WriteToUDP()` non-blocking by using Go's `SetWriteDeadline()` with a short timeout and dropping packets if the UDP buffer is full (acceptable for real-time proxy traffic).

**Warning signs:**
- CPU usage on tunnel container spikes with multiple active devices
- Throughput per device decreases as more devices connect (sublinear scaling)
- `tunToUdp` goroutine shows high time in `WriteToUDP` in pprof profiles

**Phase to address:** Performance optimization phase (not immediate priority)

---

### Pitfall 9: Dashboard Shows No Connection State Context

**What goes wrong:**
Proxy management dashboards commonly display connection credentials (host, port, username, password) without showing whether the underlying device is currently online and serving traffic. A user creates a proxy connection, downloads the .ovpn file, but the device it is assigned to is offline. The user connects the VPN, gets an IP assignment (OpenVPN layer succeeds), but traffic fails. The user has no way to know if the device was offline before purchasing/using the connection.

**How to avoid:**
- Always show device online/offline status alongside every proxy connection card/row.
- Show the last-seen timestamp for devices, not just a binary online/offline indicator.
- Display a warning on any proxy connection whose assigned device is currently offline.
- When creating a new connection, validate that the selected device is online before allowing creation.

**Warning signs:**
- Support requests are "my proxy isn't working" when the root cause is device offline
- Dashboard shows 10 connections but only 3 associated devices are online — the other 7 silently fail
- No visual distinction between a working connection and one with an offline device

**Phase to address:** Dashboard redesign phase

---

### Pitfall 10: Aggressive IP Rotation Breaking Long-Running Proxy Sessions

**What goes wrong:**
If IP rotation is implemented (the `rotation_link` domain model exists), rotating the device's mobile IP while an OpenVPN session is active causes the downstream proxy to change its source IP mid-session. For use cases requiring sticky sessions (e-commerce checkout, authenticated sessions, form submissions), this silently breaks the session. The user has a VPN tunnel that stays up but suddenly their cookies/session become invalid from the target site's perspective.

**How to avoid:**
- Implement rotation at the connection level, not the device level. New connections get the new IP; existing connections hold their current IP until they disconnect.
- Surface the current IP and "connected since" timestamp in the dashboard per connection.
- Give operators a "rotate now" button that waits for current sessions to drain (configurable drain timeout) before rotating.
- Default rotation interval should be 15-30 minutes minimum, not seconds. Per the mobile proxy research, aggressive rotation is a known pitfall.

**Warning signs:**
- Users report being logged out of sites while VPN is connected
- Rotation interval set to < 5 minutes
- No session stickiness mechanism in the connection model

**Phase to address:** Dashboard / connection management phase

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Plaintext passwords in DB (`PasswordPlain` field) | Simple auth comparison | Total credential exposure on DB breach | Never in production — hash with bcrypt before launch |
| `|| true` in OpenVPN shell scripts | Prevents VPN rejection on API failures | Silently masks setup failures, leaving users with broken connections | Never — propagate errors properly |
| Single iptables-applying goroutine with no state tracking | Simple implementation | Orphan rules on crash, requires manual cleanup | Acceptable in dev, must solve before multi-tenant use |
| Hardcoded `TUNNEL_PUSH_URL` default (`http://127.0.0.1:8081`) | Zero config for single-server | Breaks silently in multi-relay setups without explicit env var | Acceptable if env var is always set in docker-compose |
| No connection pooling in transparent proxy | Simpler connection handler | Per-request CONNECT overhead to phone proxy, kills throughput | Never — add pooling as part of OpenVPN fix |
| In-memory IP pool with no persistence | Fast allocation | All device VPN IPs re-assigned on restart, breaks active sessions | Acceptable for now, becomes a problem at scale |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| OpenVPN `client-connect` script | Using `via-env` for credential auth — credentials visible in process list | Use `via-file` for temp file auth; credentials exist only for auth duration |
| kylemanna/openvpn Docker image | Assuming the container uses standard OpenVPN config paths — it has its own wrapper scripts | Always copy custom configs explicitly; verify with `docker exec` that the running process uses your config |
| Android device HTTP proxy | Sending CONNECT with raw IP address instead of hostname — some phones (MIUI) reject IP-based CONNECT | Parse SNI/Host header and use domain name in CONNECT (already implemented in `transparentproxy/proxy.go`) |
| iptables in `network_mode: host` | Assuming container-scoped iptables — all rules are in host kernel, persist after container stop | Always teardown rules on shutdown; track state in file for crash recovery |
| PostgreSQL connection from API container | Using `localhost` instead of service name `postgres` — fails in Docker networking | Use `DB_HOST: postgres` (service name) as set in docker-compose.yml |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| No connection pooling in transparent proxy | Each HTTP resource load causes new CONNECT to phone proxy (round-trip per resource) | Pool CONNECT connections per client VPN IP | At first production use — 20-50 resources per page = 20-50 sequential setups |
| `peekTimeout = 2 seconds` blocking on TLS peek | 2 second delay on every new TLS connection, compounding for parallel resources | Reduce to 200ms; most TLS ClientHello arrives in < 100ms | Immediately — every new TLS connection |
| conntrack table saturation | iptables REDIRECT stops working for new connections, existing connections also drop | Set `nf_conntrack_max` appropriately; monitor with `conntrack -L | wc -l` | At ~50K+ concurrent tracked flows (unlikely at current scale) |
| UDP socket buffer exhaustion | High packet loss on burst traffic, throughput collapses | Already set 4MB buffers in tunnel server — verify OS allows this with `sysctl net.core.rmem_max` | At 10+ concurrent active devices |
| OpenVPN single-threaded crypto | CPU maxes out on one core, throughput plateaus | OpenVPN 2.5+ has multi-threaded TLS; verify with `top` | At ~100 Mbps aggregate through the server |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| `PasswordPlain` stored in database | Full credential exposure on DB breach or log leak | Hash with bcrypt; never log passwords even partially |
| .ovpn files with embedded `<auth-user-pass>` block | Credential exposure if file is intercepted, shared, or stored insecurely by user | Generate .ovpn without embedded credentials; show credentials separately in dashboard |
| `script-security 3` in OpenVPN client-server config | Allows any script to run as root inside the OpenVPN container | Reduce to `script-security 2` (allows calling external scripts but not shell commands directly); audit all scripts |
| Push API (port 8081) accessible without authentication | Any process that can reach port 8081 can inject iptables rules or map proxy sessions | Add shared secret header validation on push API endpoints; restrict to localhost or internal network only |
| No rate limiting on auth endpoint | Brute-force of proxy credentials via OpenVPN auth endpoint | Add rate limiting in `openvpn_handler.go`'s Auth handler (e.g., 5 attempts per minute per username) |
| API container has `NET_ADMIN` capability | Compromise of API container allows host network manipulation | Remove `NET_ADMIN` from API container; iptables operations should only happen in tunnel container |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Connection card shows credentials but not device status | User cannot tell if proxy is working before trying it | Show device online/offline indicator prominently on every connection card |
| No "copy" button on proxy credentials | Users manually type long passwords, introduce errors | One-click copy for host, port, username, password fields separately |
| OpenVPN .ovpn download with no setup instructions | Non-technical users don't know how to use the file | Show inline instructions: "Import this file into OpenVPN Connect / Tunnelblick / etc." |
| Showing raw VPN IP (192.168.255.x) as the device identifier | Meaningless to operators who named their devices | Show device name/alias, show VPN IP only in expanded technical detail |
| No bandwidth or usage visibility | Operators cannot tell which devices are overloaded or inactive | Show bytes in/out per device per day, last traffic timestamp |
| Creating connections without validating device is online | User creates a connection to an offline device and cannot use it | Validate device online status during connection creation; show warning or block |
| Dashboard refreshes wipe form state | User filling in a new connection form loses data on auto-refresh | Separate live data refresh (per-device status cards) from form state; never refresh forms |

---

## "Looks Done But Isn't" Checklist

- [ ] **OpenVPN connects:** Tunnel establishes and shows correct mobile IP — but verify a full page load works (20+ parallel connections), not just `curl -I`. Connection = success at 1% of production load.
- [ ] **iptables REDIRECT rule exists:** Confirm traffic is actually reaching transparent proxy — run `conntrack -L | grep 12345` during a connection to verify.
- [ ] **Transparent proxy mapping exists:** After OpenVPN client connect, verify the Go in-memory mapping was registered — the iptables rule without the mapping is the "connected but no traffic" failure mode.
- [ ] **Device shows online in dashboard:** Confirm the VPN IP is live by actually making a request through the HTTP/SOCKS5 port, not just checking the WebSocket status message.
- [ ] **Credentials work:** Test the generated .ovpn file on a real client, not just by checking the auth API directly.
- [ ] **MTU is correct:** A successful `curl http://example.com` through the VPN does not validate MTU. Test with a 1MB download to catch fragmentation stalls.
- [ ] **Carrier compatibility:** Test on at least two different mobile carriers (not the same carrier on different devices) — CGNAT and UDP blocking behavior differs by carrier.
- [ ] **Dashboard device status is real-time:** Verify that marking a device offline (kill the Android app) updates the dashboard within 30 seconds without manual refresh.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| OpenVPN throughput bottleneck (active bug) | MEDIUM | 1) tcpdump diagnosis 2) add CONNECT pooling to transparent proxy 3) tune MTU/mssfix 4) validate with speed test |
| Orphan iptables rules after crash | LOW | `iptables -t nat -L PREROUTING -n` to inspect; `iptables -t nat -F PREROUTING` to flush (will break active sessions); restart tunnel container |
| Plaintext passwords exposed (DB breach) | HIGH | Rotate all generated passwords immediately; hash new passwords with bcrypt; notify affected customers |
| Device offline during active VPN sessions | LOW | Dashboard "disconnect all" button that sends explicit disconnect to OpenVPN management interface; iptables rules auto-clean on reconnect |
| Race condition leaves client connected but unmapped | LOW | Reconnect (disconnect + reconnect) from OpenVPN client forces new client-connect hook execution |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Transparent proxy double-encapsulation bottleneck | OpenVPN fix phase (immediate) | Speed test through VPN shows > 5 Mbps on idle cellular connection |
| MTU mismatch causing fragmentation | OpenVPN fix phase (immediate) | 1MB file download through VPN completes without stall |
| UDP blocked by cellular carriers | OpenVPN fix phase + carrier testing | Test on 3 different carrier SIMs; all show stable connection under load |
| Plaintext credentials | Security phase (before production) | Database dump contains only bcrypt hashes; no `PasswordPlain` column in production schema |
| iptables/mapping race condition | OpenVPN fix phase | 50 rapid connect/disconnect cycles produce zero "no mapping" log entries |
| Device offline shown as online | Dashboard redesign phase | Device disconnect visible in dashboard within 30 seconds; affected connections show warning |
| iptables rules surviving restart | Production hardening phase | Container restart leaves zero orphan rules (`iptables -t nat -L PREROUTING -n` count unchanged) |
| Single tunToUdp goroutine bottleneck | Performance phase (deferred) | Throughput scales linearly with device count up to 20 devices |
| Dashboard shows no device status on connections | Dashboard redesign phase | Every connection card shows device online/offline status and last-seen time |
| Aggressive rotation breaking sessions | Connection management phase | Rotation waits for active sessions to drain before switching IP |

---

## Sources

- Direct code inspection: `server/cmd/tunnel/main.go`, `server/internal/transparentproxy/proxy.go`, `server/internal/api/handler/openvpn_handler.go`, `server/deployments/openvpn/client-server.conf` (HIGH confidence)
- [Optimizing OpenVPN Throughput — hamy.io](https://hamy.io/post/0003/optimizing-openvpn-throughput/) (MEDIUM confidence)
- [Improving OpenVPN Performance and Throughput — linuxblog.io](https://linuxblog.io/improving-openvpn-performance-and-throughput/) (MEDIUM confidence)
- [OpenVPN MTU Finding the Correct Settings — The Geek Pub](https://www.thegeekpub.com/271035/openvpn-mtu-finding-the-correct-settings/) (MEDIUM confidence)
- [Solving OpenVPN MTU Issues — blog.hambier.lu](https://blog.hambier.lu/post/solving-openvpn-mtu-issues) (MEDIUM confidence)
- [iptables REDIRECT vs DNAT vs TPROXY — gsoc-blog.ecklm.com](http://gsoc-blog.ecklm.com/iptables-redirect-vs.-dnat-vs.-tproxy/) (MEDIUM confidence)
- [When Linux Conntrack is No Longer Your Friend — Tigera](https://www.tigera.io/blog/when-linux-conntrack-is-no-longer-your-friend/) (MEDIUM confidence)
- [How to Build Your Own Mobile Proxy Farm — mobileproxy.space](https://mobileproxy.space/en/pages/how-to-build-your-own-mobile-proxy-farm-a-step-by-step-beginners-guide.html) (MEDIUM confidence)
- [OpenVPN forum: Over WiFi works, Over mobile data only partially — forums.openvpn.net](https://forums.openvpn.net/viewtopic.php?t=33369) (MEDIUM confidence, WebSearch)
- [Dealing with Cellular Broadband CGNAT — damow.net](https://damow.net/dealing-with-cellular-broadband-cgnat/) (MEDIUM confidence)
- [WebSocket Reconnection on Android — ably.com](https://ably.com/topic/websockets-android) (LOW confidence — general, not proxy-specific)

---

*Pitfalls research for: PocketProxy — mobile proxy platform (Android devices, HTTP/SOCKS5/OpenVPN)*
*Researched: 2026-02-25*
