# Phase 1: OpenVPN Throughput - Research

**Researched:** 2026-02-25
**Domain:** OpenVPN performance tuning, transparent TCP proxy, Linux iptables NAT, Go networking
**Confidence:** HIGH (codebase fully read; external claims verified against official OpenVPN docs and community sources)

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PROTO-01 | OpenVPN direct access delivers usable throughput (pages load fully, speed tests complete) | Buffer tuning (sndbuf/rcvbuf 0), mssfix, fast-io, peekTimeout reduction, MTU alignment, cipher selection all directly address throughput |
| PROTO-02 | HTTP and SOCKS5 proxies remain stable under production load | Connection pooling in transparent proxy, race condition fix in client-connect-ovpn.sh (iptables REDIRECT must exist before OpenVPN considers connection open), DNAT PREROUTING rules isolated from OpenVPN client path |
</phase_requirements>

---

## Summary

PocketProxy's OpenVPN throughput path has two distinct problem areas. The first is the customer-facing OpenVPN server (`openvpn-client` container, port 1195, tun1) and its interaction with the transparent proxy (`proxy.go`) that routes each customer's TCP traffic through the phone's HTTP proxy on the device VPN subnet. The second is PROTO-02 stability: HTTP/SOCKS5 proxies use a separate DNAT path that must survive concurrent OpenVPN config changes without disruption.

The transparent proxy already has correct structural design (SNI extraction, HTTP Host extraction, bidirectional copy with half-close). The bottleneck is almost certainly not the Go proxy code itself, but a combination of: (a) `peekTimeout` blocking the hot path for 2 seconds on every new connection, (b) the `sndbuf`/`rcvbuf` values being set to fixed 512 KB instead of letting the OS autosize, and (c) OpenVPN's `client-connect-ovpn.sh` notifying the tunnel API (and thus adding the iptables REDIRECT rule) *after* OpenVPN considers the client admitted, meaning the client's first packets arrive before the transparent proxy mapping exists and are silently dropped.

The diagnosis sequence is: instrument latency (first-byte time on tun1) vs. bandwidth separately, because the fix differs — peekTimeout kills first-byte latency, buffer sizing kills sustained throughput. The iptables race kills the first few seconds of every new connection regardless of throughput.

**Primary recommendation:** Fix the iptables/mapping race first (that one blocks all traffic briefly on every connect), then reduce peekTimeout to 200 ms, then switch sndbuf/rcvbuf to 0 on both server and client configs and push "sndbuf 0"/"rcvbuf 0" to clients. These three changes are low-risk and cover the most likely causes. Use `tcpdump -i tun1` + `iperf3` through the tunnel to measure before/after.

---

## Standard Stack

### Core (already in use — do not change)

| Component | Version | Purpose | Notes |
|-----------|---------|---------|-------|
| kylemanna/openvpn | 2.4 (image tag) | OpenVPN server containers | Unmaintained but functional; defer image upgrade per STATE.md decision |
| Go transparent proxy | `internal/transparentproxy/proxy.go` | Routes OpenVPN client TCP through phone HTTP proxy | Correct architecture, tuning needed |
| iptables REDIRECT + RETURN | Linux kernel | Intercepts OpenVPN client TCP for transparent proxy | Race condition: rules added after client admitted |
| Linux iptables DNAT/PREROUTING | Linux kernel | Routes HTTP/SOCKS5 ports to device VPN IPs | Used by HTTP/SOCKS5 path, independent of OpenVPN path |

### Supporting

| Tool | Purpose | When to Use |
|------|---------|-------------|
| `tcpdump -i tun1` | Capture raw packets on OpenVPN client tun interface | Diagnosis — see if packets arrive at all |
| `iperf3` | Bandwidth measurement through tunnel | Distinguish latency vs. throughput bottleneck |
| `ping` through tunnel | Latency measurement | Baseline RTT to isolate crypto overhead from bandwidth limits |
| `ss -s` / `netstat -s` | Socket buffer statistics | Verify sndbuf/rcvbuf changes took effect |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| iptables REDIRECT (current) | iptables TPROXY | TPROXY is more correct (preserves original dst without `getsockopt SO_ORIGINAL_DST`), but requires kernel `ip_vs` module and more complex routing; REDIRECT works for this use case |
| kylemanna/openvpn 2.4 | Build custom image with OpenVPN 2.6 | 2.6 adds `data-ciphers` negotiation, AEAD-only mode, smaller per-packet overhead; defer per STATE.md |
| Fixed sndbuf/rcvbuf values | `sndbuf 0` / `rcvbuf 0` | Setting to 0 lets Linux TCP autotuning adapt; fixed values cap throughput on high-BDP links |

---

## Architecture Patterns

### Traffic Flow: Customer OpenVPN Connection

```
Customer client (laptop/phone)
  ↓ UDP/1195
[openvpn-client container] (tun1, 10.9.0.0/24)
  ↓ decrypted TCP (10.9.0.x → internet dest)
  ↓ iptables PREROUTING REDIRECT → :12345
[tunnel container: transparent proxy :12345]
  ↓ SO_ORIGINAL_DST for real dest
  ↓ extract SNI or HTTP Host
  ↓ HTTP CONNECT to device HTTP proxy
[device HTTP proxy on 192.168.255.y:8080]
  ↓ cellular internet
```

The tunnel container runs both the device tunnel server (tun0, 192.168.255.0/24) and the transparent proxy. Both are in the same process (`cmd/tunnel/main.go`). The `network_mode: host` means iptables rules apply to the host network namespace.

### Pattern 1: Race Condition in client-connect-ovpn.sh

**What:** OpenVPN calls `client-connect-ovpn.sh` synchronously before the client can send IP traffic. The script sends a `wget` POST to the API, which then POSTs to the tunnel's `/openvpn-client-connect` endpoint. Only after that POST completes does the transparent proxy mapping and iptables REDIRECT rule exist. However, OpenVPN 2.4 uses a blocking client-connect: the server waits for the script to return before the client is admitted. If the API call in the script is fast enough, there is NO race — the REDIRECT rule exists before the first client IP packet.

**The actual risk:** `wget ... || true` means if the API call fails (network error, API down), the client is admitted anyway with NO iptables REDIRECT rule and NO mapping — traffic flows unproxied or is dropped by UFW.

**Verified from code:** `client-connect-ovpn.sh` line 16: `wget -q -O - ... || true` — failure is silently swallowed and exit 0 is always returned. OpenVPN then admits the client. Without the iptables REDIRECT rule, the client's TCP traffic hits the kernel's normal routing and either goes out raw (bypassing the phone proxy) or is dropped.

**How to fix:** Remove `|| true` — let the script exit 1 on API failure, which causes OpenVPN to reject the connection with an error. Alternatively, make the API call synchronous and check the HTTP response code before exiting. Add a retry with timeout.

**Second risk:** Ordering of iptables `-A` (append) rules. The current code appends RETURN rules before the REDIRECT rule, which is correct. But `-A` appends to the end of the PREROUTING chain — if UFW or other rules run before PREROUTING custom rules, order matters. The code uses `-A` not `-I`, so earlier rules could match first. This is lower risk than the `|| true` issue.

### Pattern 2: peekTimeout Blocking Every Connection

**What:** `proxy.go` line 141-143:
```go
conn.SetReadDeadline(time.Now().Add(peekTimeout))  // peekTimeout = 2 seconds
peeked, _ := clientReader.Peek(peekSize)
conn.SetReadDeadline(time.Time{})
```

**Effect:** For EVERY new connection, the transparent proxy blocks up to 2 seconds waiting for client data. TLS ClientHello typically arrives within milliseconds, but anything that triggers a full 2-second wait (e.g., a protocol that sends server-first, or a slow client) kills first-byte latency. For a speed test hitting many connections in parallel, this multiplies.

**Fix:** Reduce `peekTimeout` from 2s to 100–200ms. If no data arrives, fall back to raw IP+port routing (already done via `connectAddr = origDst`).

### Pattern 3: sndbuf/rcvbuf Fixed Values vs. OS Autotuning

**What:** `client-server.conf` has:
```
sndbuf 524288
rcvbuf 524288
push "sndbuf 524288"
push "rcvbuf 524288"
```

**Effect:** These 512 KB fixed buffers override the OS socket autotuning. On high-BDP cellular links (e.g., 50 Mbps LTE with 80ms RTT, BDP = ~500 KB), 512 KB is adequate but still caps throughput. The real issue is this overrides Linux's TCP Autotune, which would grow buffers dynamically. Setting `sndbuf 0` / `rcvbuf 0` lets the OS tune buffers per-connection, which is consistently faster than any fixed value per community testing (5 Mbps → 60 Mbps documented in angristan/openvpn-install#352).

**Verified:** OpenVPN docs confirm: value 0 means "use OS default." The `configureTUN()` in `tunnel/main.go` already sets `net.core.rmem_max=16777216` and `net.core.wmem_max=16777216`, so OS autotune has headroom to grow.

### Pattern 4: fast-io Impact (Low Priority)

**What:** `fast-io` in OpenVPN avoids poll/epoll/select before each UDP write. Only works with `proto udp` (which this is) on non-Windows.

**Effect:** 5–10% CPU reduction. On cellular where CPU is rarely the bottleneck, this is low priority. Add it to `client-server.conf` after the higher-impact changes are made.

### Pattern 5: Connection Pooling in Transparent Proxy (PROTO-02)

**What:** The transparent proxy creates a new TCP connection to the phone's HTTP proxy for every tunneled connection (`dialHTTPConnect` is called fresh per `handleConn`). Each CONNECT handshake adds RTT overhead.

**Effect:** For HTTP/1.1 connections from the customer's browser, the browser holds connections open (keep-alive), so the proxy's per-connection setup cost is paid once per browser connection, not per request. This is acceptable. For HTTP/2 or TLS multiplexed connections, a single browser connection opens once and reuses the tunnel. **Connection pooling to the phone proxy would only help if the customer's software opens many short-lived connections.** This is secondary to fixing the race and timeout issues.

**Note:** PROTO-02 (HTTP and SOCKS5 proxies remain stable) refers primarily to the DNAT path (`PREROUTING DNAT` rules), not the transparent proxy. The DNAT path is entirely separate from the OpenVPN client REDIRECT path. Changes to OpenVPN client iptables rules (REDIRECT) do not touch DNAT rules. PROTO-02 is satisfied as long as: (a) the DNAT rules for the device are not accidentally deleted when adding OpenVPN REDIRECT rules, and (b) the `AddMapping`/`RemoveMapping` calls in the transparent proxy are thread-safe (they are — `sync.RWMutex`).

### Anti-Patterns to Avoid

- **Do not increase mssfix above 1400 without measurement:** Higher mssfix values can cause fragmentation if the actual link MTU is lower. Current `mssfix 1400` in `client-server.conf` is a safe default; leave it unless tcpdump shows fragmentation.
- **Do not change tun-mtu without changing mssfix together:** They are coupled. Current config has `tun-mtu 1500` (default) and `mssfix 1400`. This is correct. Changing one requires recalculating the other.
- **Do not upgrade kylemanna/openvpn image during Phase 1:** STATE.md explicitly defers this. The image is unmaintained but functional. OpenVPN 2.6 changes cipher negotiation (`ncp-ciphers` → `data-ciphers`) which would require config changes; do not introduce that risk while debugging throughput.
- **Do not add compression:** LZ4/LZO adds CPU overhead with negligible benefit on pre-encrypted HTTPS traffic (which is ~95% of web traffic). The current config has no compression — keep it that way.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SNI extraction | Custom parser | Already exists in `proxy.go` | The existing parser handles TLS 1.2/1.3 ClientHello correctly |
| HTTP Host extraction | Custom parser | Already exists in `proxy.go` | Handles CRLF and LF line endings |
| Bidirectional copy | Custom goroutine logic | `io.CopyBuffer` + `CloseWrite` pattern (already in code) | Half-close semantics are subtle; current code handles them correctly |
| iptables rule management | Custom netlink library | `exec.Command("iptables", ...)` (current) | iptables CLI is the right level of abstraction for this use case |

**Key insight:** The existing code is structurally correct. The work is tuning constants and fixing the error handling in the shell script — not rewriting components.

---

## Common Pitfalls

### Pitfall 1: Diagnosing the Wrong Layer

**What goes wrong:** Measuring speed with a web browser and concluding "OpenVPN is slow" when the actual bottleneck is the 2-second peekTimeout on every new connection.
**Why it happens:** Browser speed tests open many parallel connections; each one pays the 2-second timeout.
**How to avoid:** Use `iperf3 -c <server> -P 4` through the VPN tunnel to measure raw bandwidth without application protocol overhead. Then use `curl --max-time 10 -o /dev/null -w "%{time_starttransfer}" https://example.com` to measure first-byte latency. If iperf3 bandwidth is good but first-byte is slow, the problem is peekTimeout.
**Warning signs:** Speed test shows < 1 Mbps but `ping` through tunnel shows reasonable RTT (< 200ms).

### Pitfall 2: Swallowed Script Errors

**What goes wrong:** `client-connect-ovpn.sh` exits 0 even when the API call fails, so customers get admitted with no iptables REDIRECT rule, and their traffic is either dropped or unproxied.
**Why it happens:** `|| true` at the end of the `wget` call.
**How to avoid:** Remove `|| true`. Let non-zero exit from wget propagate. OpenVPN will return a connection error to the client, which is the correct behavior when the mapping can't be set up.
**Warning signs:** Customer reports "VPN connects but no internet," especially on first connect or after server restart.

### Pitfall 3: iptables Rule Ordering with UFW

**What goes wrong:** The REDIRECT rules are added with `-A` (append), so they go after any UFW rules in PREROUTING. If UFW has a RETURN or DROP rule that matches before the REDIRECT, traffic bypasses the transparent proxy.
**Why it happens:** UFW adds its own PREROUTING rules. On Ubuntu, UFW inserts a `ufw-before-logging-forward` chain early.
**How to avoid:** Use `-I PREROUTING 1` (insert at position 1) for the REDIRECT rules, not `-A`. The existing RETURN rules for VPN subnets (10.9.0.0/24 and 192.168.255.0/24) should also be inserted before the REDIRECT.
**Warning signs:** iptables rules appear correct when listed, but traffic doesn't reach the transparent proxy.

### Pitfall 4: sndbuf/rcvbuf Mismatch Between Server and Client

**What goes wrong:** Server config sets sndbuf/rcvbuf to 0 but client config retains explicit values (or vice versa), causing a mismatch.
**Why it happens:** The server pushes `push "sndbuf X"` to override client settings. If you change server to 0 but forget to change the push directive, the client gets 0 from server push but the push only applies to the client's data channel socket.
**How to avoid:** Change both `sndbuf`/`rcvbuf` lines AND the corresponding `push "sndbuf ..."`/`push "rcvbuf ..."` lines in `client-server.conf` together. The generated `.ovpn` file in `openvpn_handler.go` (DownloadOVPN) also has hardcoded `sndbuf 524288` / `rcvbuf 524288` — update this too.
**Warning signs:** Speed improves on some clients but not others depending on which config they use.

### Pitfall 5: Transparent Proxy Not Receiving Connections After Restart

**What goes wrong:** After restarting the tunnel container, existing OpenVPN client sessions have stale iptables REDIRECT rules pointing to the transparent proxy, but the transparent proxy mappings (in-memory) are lost on restart.
**Why it happens:** The `Proxy.mappings` map is in-memory; restarting the tunnel process clears it. Reconnecting the OpenVPN client triggers client-connect-ovpn.sh again and re-adds the mapping, but only after the reconnect.
**How to avoid:** This is acceptable behavior — OpenVPN clients will reconnect after keepalive timeout. Document it. No code change needed.

---

## Code Examples

### Diagnosis: Capture Traffic on tun1

```bash
# On the relay server — check if decrypted packets appear on tun1
# If you see TCP SYN packets here, OpenVPN is decrypting correctly
tcpdump -i tun1 -n -c 100

# Capture with timestamp to measure first-byte delay
tcpdump -i tun1 -n -tttt 'tcp[tcpflags] & tcp-syn != 0'

# Check if traffic reaches transparent proxy port 12345
tcpdump -i any port 12345 -n -c 50
```

### Diagnosis: Measure Throughput vs. Latency

```bash
# On relay server — iperf3 server
iperf3 -s

# On customer machine through VPN — measure bandwidth
iperf3 -c <relay_server_ip> -t 10 -P 4

# Measure first-byte latency
curl -o /dev/null -s -w "time_starttransfer: %{time_starttransfer}s\n" https://example.com
```

### Fix: peekTimeout Reduction in proxy.go

```go
// Current value (line 29 in proxy.go):
peekTimeout = 2 * time.Second

// Change to:
peekTimeout = 200 * time.Millisecond
```

### Fix: sndbuf/rcvbuf in client-server.conf

```
# Replace:
sndbuf 524288
rcvbuf 524288
push "sndbuf 524288"
push "rcvbuf 524288"

# With:
sndbuf 0
rcvbuf 0
push "sndbuf 0"
push "rcvbuf 0"
```

### Fix: sndbuf/rcvbuf in DownloadOVPN (openvpn_handler.go)

```go
// Current (lines 218-219 in openvpn_handler.go):
ovpn.WriteString("sndbuf 524288\n")
ovpn.WriteString("rcvbuf 524288\n")

// Change to:
ovpn.WriteString("sndbuf 0\n")
ovpn.WriteString("rcvbuf 0\n")
```

### Fix: client-connect-ovpn.sh — Remove Silent Failure

```sh
# Current (line 16):
wget -q -O - --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/connect" || true

# Fix — fail loudly so OpenVPN rejects the client on API failure:
if ! wget -q -O - --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/connect"; then
  echo "ERROR: Failed to notify API for $username — rejecting client"
  exit 1
fi
```

### Fix: fast-io addition to client-server.conf

```
# Add after the performance tuning block:
fast-io
```

### Add Connection Pooling to Transparent Proxy (PROTO-02, optional)

```go
// If connection pooling to phone proxy is needed:
// Use sync.Pool to reuse established CONNECT connections.
// WARNING: Phone proxy may close idle connections; verify keepalive behavior first.
// Current implementation (fresh dial per conn) is correct for correctness.
// Only add pooling if profiling shows dialHTTPConnect is a bottleneck.
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Fixed sndbuf/rcvbuf (64 KB) | OS autotuning via sndbuf 0 | OpenVPN 2.3+ | Dramatic improvement on high-BDP links (5 → 60 Mbps documented) |
| LZO compression | No compression or lz4-v2 | OpenVPN 2.4+ | Reduced CPU with no benefit on HTTPS traffic |
| cipher AES-256-CBC | AES-128-GCM (AEAD) | OpenVPN 2.4+ | ~2x faster encryption, 36→20 bytes overhead per packet |
| ncp-ciphers | data-ciphers | OpenVPN 2.6 | Config syntax change; not relevant (still on 2.4) |

**Deprecated/outdated:**
- `cipher` directive in OpenVPN 2.6: deprecated; use `data-ciphers`. Not relevant to current 2.4 image.
- `comp-lzo`: deprecated in 2.5. Not in current config; keep it out.

---

## Open Questions

1. **Is the openvpn-client container actually running on the relay server?**
   - What we know: `docker-compose.yml` has the service; `setup-client-ovpn.sh` exists
   - What's unclear: Whether PKI is initialized and the container is actually running in production
   - Recommendation: The diagnosis step (Plan 01-01) should start with `docker ps` and `docker logs openvpn-client` to confirm the container state

2. **What is the actual observed symptom?**
   - What we know: "OpenVPN throughput is effectively broken" (ROADMAP.md) and speed test timeout (success criteria)
   - What's unclear: Is the tunnel connecting at all but slow? Or failing to route traffic entirely?
   - Recommendation: Plan 01-01 diagnosis step should check `docker logs openvpn-client` for authentication errors before doing tcpdump

3. **Does the transparent proxy need connection pooling to the phone proxy?**
   - What we know: Each customer connection creates a fresh HTTP CONNECT to the phone
   - What's unclear: Whether the phone's HTTP proxy supports persistent connections and whether this matters in practice
   - Recommendation: Only add pooling if profiling shows it matters; the phone proxy likely handles keep-alive already

4. **Is the `client-server.conf` actually being used by the openvpn-client container?**
   - What we know: `setup-client-ovpn.sh` copies it to `/etc/openvpn/openvpn.conf` inside the Docker volume
   - What's unclear: Whether the setup script has been run since the performance tuning configs were added
   - Recommendation: Verify with `docker exec openvpn-client cat /etc/openvpn/openvpn.conf` as part of Plan 01-01 diagnosis

5. **Are mssfix 1400 and tun-mtu 1500 optimal for the cellular path?**
   - What we know: Default mssfix 1400 is a safe conservative value for UDP tunnels; tun-mtu 1500 is standard
   - What's unclear: Whether cellular carriers have lower effective MTU (some cap at 1480 or less due to their own tunneling)
   - Recommendation: Run `ping -M do -s 1400 <remote>` through the VPN to probe effective MTU; adjust mssfix if fragmentation is detected

---

## Validation Architecture

> Skipped — `workflow.nyquist_validation` is not present in `.planning/config.json` (treated as false).

---

## Sources

### Primary (HIGH confidence)
- Codebase: `server/internal/transparentproxy/proxy.go` — full transparent proxy implementation read directly
- Codebase: `server/deployments/openvpn/client-server.conf` — current OpenVPN server config
- Codebase: `server/deployments/openvpn/client-connect-ovpn.sh` — connect hook with silent failure bug
- Codebase: `server/cmd/tunnel/main.go` — iptables rule management, proxy startup, push API handlers
- Codebase: `server/internal/api/handler/openvpn_handler.go` — DownloadOVPN generates .ovpn with hardcoded buffer values

### Secondary (MEDIUM confidence)
- [angristan/openvpn-install issue #352](https://github.com/angristan/openvpn-install/issues/352) — sndbuf 0/rcvbuf 0 fix: 5 Mbps → 60 Mbps, confirmed by multiple commenters
- [linuxblog.io: Improving OpenVPN performance](https://linuxblog.io/improving-openvpn-performance-and-throughput/) — sndbuf 512000, fast-io 5–10% CPU improvement
- [hamy.io: Optimizing OpenVPN Throughput](https://hamy.io/post/0003/optimizing-openvpn-throughput/) — mssfix = link-mtu formula, sndbuf 0 rationale
- [OpenVPN community: client-connect timing](https://forums.openvpn.net/viewtopic.php?t=40796) — confirmed client-connect script blocks client admission in OpenVPN 2.4

### Tertiary (LOW confidence)
- OpenVPN CGNAT discussion threads — UDP unreliability on some cellular carriers; switching to TCP may help as last resort (not recommended for throughput)
- General iptables TPROXY documentation — alternative to REDIRECT; lower risk but higher complexity

---

## Metadata

**Confidence breakdown:**
- Core bottleneck identification (peekTimeout, sndbuf/rcvbuf, || true bug): HIGH — code read directly, issues are unambiguous
- Fix effectiveness (sndbuf 0 impact): MEDIUM — community-verified, multiple sources agree
- fast-io impact: MEDIUM — documented as 5–10% CPU; minor priority
- Connection pooling benefit: LOW — not measured; theoretical
- iptables ordering with UFW: MEDIUM — behavior is documented Linux/UFW interaction, not verified for this specific deployment

**Research date:** 2026-02-25
**Valid until:** 2026-04-25 (stable OpenVPN/Linux kernel domain; 60 days)
