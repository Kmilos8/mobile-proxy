# Project Research Summary

**Project:** PocketProxy — Mobile Proxy Platform
**Domain:** Android device fleet as HTTP/SOCKS5/OpenVPN proxy endpoints with operator dashboard
**Researched:** 2026-02-25
**Confidence:** MEDIUM (OpenVPN tuning: MEDIUM; architecture: HIGH from codebase; features: MEDIUM from competitor analysis; pitfalls: HIGH from code inspection)

## Executive Summary

PocketProxy is a mobile proxy platform where Android devices act as proxy endpoints, routing customer traffic through real cellular connections. The platform is already in partial production: the Go API, tunnel server, PostgreSQL, Docker Compose infrastructure, and Android app all work. Two problems are blocking the current milestone — OpenVPN customer throughput is effectively broken ("pages barely load"), and the Next.js dashboard is a scaffold with no functional UI. Both are solvable with targeted fixes; neither requires architectural replacement.

The recommended approach is sequential: fix the OpenVPN throughput problem first (it is blocking any production use of the customer-facing VPN feature), then build out the dashboard UI against the existing API. The throughput fix is primarily a configuration problem — setting `sndbuf 0` and `rcvbuf 0` instead of hardcoded 512KB values, adding `mssfix 1300` to client `.ovpn` files, adding `fast-io` to the server config, and adding connection pooling in the transparent proxy to eliminate per-request HTTP CONNECT overhead. No new infrastructure is needed. Dashboard work uses shadcn/ui + TanStack Query layered on top of the existing Next.js 14 + Tailwind setup.

The key risks are: (1) the transparent proxy double-encapsulation bottleneck — each TCP connection through the OpenVPN path requires two full proxy hops with no connection reuse, compounding for modern pages with 20-50 parallel connections; (2) plaintext credentials stored in the database (`PasswordPlain` field) which must be hashed with bcrypt before any production deployment; and (3) the `peekTimeout = 2 seconds` in the transparent proxy which adds 2 seconds latency to every new TLS connection. Address these in the OpenVPN fix phase before any customer-facing launch.

---

## Key Findings

### Recommended Stack

The existing stack is correct and must not be replaced. The only additions are: (1) OpenVPN config parameter changes — no new software, just tuning `client-server.conf` and regenerated `.ovpn` files; and (2) dashboard UI libraries layered onto the existing Next.js 14 + Tailwind scaffold.

For the dashboard, shadcn/ui is the right choice — it is the official recommendation for Next.js + Tailwind, uses a copy-paste model with no version lock-in, and provides every component needed (Card, Table, Badge, Dialog, Select). TanStack Query v5 replaces manual `fetch` in `useEffect` and provides polling with `refetchInterval` for live device status. React Hook Form + Zod handles proxy creation forms. All three integrate cleanly with the existing codebase. The kylemanna/openvpn image is unmaintained but must not be replaced during the throughput debugging phase — replace only after throughput is confirmed fixed.

**Core technologies (existing — keep):**
- Go + Gin: API server, tunnel server, worker — production-quality, no changes
- PostgreSQL 16: all persistent state — working, 9 migration files
- Custom UDP tunnel: device-to-server connectivity — working, Android authenticates and maintains tunnel
- Docker Compose: deployment — correct for current scale (0-50 devices)
- Next.js 14.1 + TypeScript + Tailwind: dashboard — scaffold only, needs UI libraries
- kylemanna/openvpn 2.4: both VPN networks — working but unmaintained; defer image replacement

**New libraries to add (dashboard only):**
- shadcn/ui (CLI: `shadcn@latest`): component library — the official Next.js + Tailwind choice
- @tanstack/react-query v5: data fetching and polling — replaces manual fetch/useEffect
- react-hook-form v7 + zod v3 + @hookform/resolvers v3: form management and validation

**OpenVPN config changes (no new software):**
- Set `sndbuf 0` / `rcvbuf 0` (OS-managed buffers; fixes 5→60 Mbps issues in documented cases)
- Add `fast-io` to `client-server.conf` (non-blocking UDP writes; 5-10% CPU gain)
- Set `mssfix 1300` in both server config and all client `.ovpn` files
- Add connection pooling in `transparentproxy/proxy.go` (eliminates per-request CONNECT overhead)
- Reduce `peekTimeout` from 2000ms to 200ms in transparent proxy

### Expected Features

The product is behind both primary competitors (iProxy.online, Proxidize) on almost every feature. The dashboard is a scaffold. The priority is getting the core proxy workflow functional for one operator first, then expanding.

**Must have for launch (v1):**
- Device list with online/offline status, current IP, device name — operators fly blind without this
- Proxy credentials display per port (HTTP and SOCKS5 host:port:user:pass) — customers cannot connect without this
- Multiple proxy ports per device (target 10) — single port per device is not economically viable
- Manual IP rotation from dashboard — every proxy product has this; no exceptions
- OpenVPN `.ovpn` file generation and download per port — third protocol, core to the product
- Proxy port create and delete — basic CRUD for port management
- QR code device onboarding — eliminates manual token entry; `qrcode.react` already installed
- Basic traffic counter per device — operators need to see usage
- Online/offline notification via email — operators must be alerted when devices drop
- Current IP display per port — updated after each rotation

**Should have after validation (v1.x):**
- Per-device carrier name, signal strength, battery level display
- Automatic IP rotation with configurable interval
- IP whitelist authentication (per port)
- Rotation via unique URL for external automation
- Device grouping with bulk actions (fleet size threshold: 20+ devices)
- REST API for programmatic proxy management
- IP history (last 5-10 IPs per port)

**Defer to v2+:**
- Multi-tenant team access with permission levels (requires user account overhaul)
- Proxy blacklist per port (no validated customer demand)
- Passive OS fingerprint spoofing (high complexity, limited audience)
- SMS forwarding (legal risk, requires deliberate TOS design)
- Per-port traffic logs with payload capture (storage costs, privacy concerns)

### Architecture Approach

The platform runs two entirely separate VPN networks simultaneously. Network 1 (port 1194 / `192.168.255.0/24`): the custom Go UDP tunnel that Android devices connect to — this must NOT redirect gateway or device traffic routes through the server instead of cellular. Network 2 (port 1195 / `10.9.0.0/24`): standard OpenVPN (kylemanna image) for customer access — this DOES redirect gateway so all customer traffic exits via the phone's cellular IP. Keeping these two networks strictly separate is the core architectural constraint; every feature decision must respect this boundary.

The transparent proxy is the critical path for OpenVPN customer traffic: OpenVPN client → tun1 → iptables REDIRECT → transparent proxy (port 12345) → HTTP CONNECT → Android device proxy (192.168.255.y:8080) → cellular internet. This two-hop chain is architecturally correct but currently unoptimized — no connection pooling means every TCP connection causes a fresh HTTP CONNECT, and the 2-second peek timeout adds latency to every TLS handshake.

**Major components:**
1. Go Tunnel Server (`cmd/tunnel/main.go`, ~980 lines) — device UDP tunnel, TUN interface management, iptables DNAT/REDIRECT, transparent proxy, push API on port 8081
2. Go API Server — business logic, REST API, WebSocket hub, JWT auth; all dashboard and tunnel server communication flows through it
3. OpenVPN (kylemanna, port 1195) — customer VPN access; auth delegates to API via shell scripts; transparent proxy handles traffic interception
4. Go Worker — background jobs: auto-rotate scheduler, stale device cleanup, bandwidth rollups
5. Next.js Dashboard — admin UI; reads from API via REST (JWT) and WebSocket (`/ws`) for real-time device status
6. Android App — device-side HTTP proxy (:8080), SOCKS5 proxy (:1080), UDP tunnel client, 30-second heartbeat loop

**Key patterns:**
- Two-phase device connect: UDP auth → tunnel assigns 192.168.255.x → API notifies → iptables DNAT installed
- Command delivery with dual-path fallback: push API (immediate) → heartbeat delivery (fallback for offline devices)
- Port allocation via sequential counter starting at 30000, +4 per device (HTTP, SOCKS5, UDP relay, OVPN)
- Idempotent iptables teardown: always teardown before setup; loop `iptables -D` until failure to clear all duplicates

### Critical Pitfalls

1. **Transparent proxy double-encapsulation bottleneck** — No connection pooling to the phone's HTTP proxy means every TCP connection (20-50 per modern HTTPS page) triggers a full HTTP CONNECT round-trip. Fix: add a persistent connection pool per client VPN IP in `transparentproxy/proxy.go`. Measure first with tcpdump on tun1 to confirm this is the bottleneck vs. MTU fragmentation.

2. **MTU mismatch causing silent packet fragmentation** — tun1 MTU 1500 + OpenVPN 69-byte overhead + tun0 MTU 1400 = fragmentation that mobile carriers' ICMP black holes prevent PMTUD from correcting. Fix: set `tun-mtu 1420` and `mssfix 1360` on client-server OpenVPN; verify with `ping -M do -s 1300 8.8.8.8` through the VPN.

3. **Plaintext credentials in database (`PasswordPlain` field)** — A database breach exposes all proxy credentials. Fix: hash with bcrypt (cost 10) before any production deployment; update `openvpn_handler.go` auth comparison to use `bcrypt.CompareHashAndPassword()`.

4. **iptables/mapping race condition** — The `|| true` in `client-connect-ovpn.sh` masks API failures, leaving the iptables REDIRECT rule active but no transparent proxy mapping registered. Symptom: VPN connects but all traffic times out immediately. Fix: remove `|| true`; add idempotent cleanup on connect; add retry loop in transparent proxy when `getMapping()` returns false.

5. **2-second peek timeout multiplying latency** — `peekTimeout = 2 seconds` in `transparentproxy/proxy.go` adds 2 seconds to every new TLS connection because the proxy waits for SNI before forwarding. Fix: reduce to 200ms; TLS ClientHello arrives in under 100ms in practice.

---

## Implications for Roadmap

Based on the combined research, the dependency graph is clear: OpenVPN throughput must be fixed before any customer-facing launch; dashboard work is independent of the data plane and can proceed in parallel; security hardening must happen before production exposure.

### Phase 1: OpenVPN Throughput Fix

**Rationale:** This is the active blocking bug. Every customer-facing capability depends on the OpenVPN path working reliably. Architecture research confirms the bottleneck is in the transparent proxy + config tuning, not the infrastructure. This is config changes + one targeted Go code change, not a rewrite. Must be resolved before any dashboard work is useful in production.

**Delivers:** A working OpenVPN customer path where customers can connect via `.ovpn` file and browse the web through the device's cellular connection at acceptable speed (target: >5 Mbps on an idle cellular connection).

**Addresses (from FEATURES.md):** OpenVPN `.ovpn` generation and download (table stakes, currently partially working), proxy port management basics.

**Avoids (from PITFALLS.md):**
- Transparent proxy double-encapsulation bottleneck (add connection pooling)
- MTU mismatch fragmentation (tune `tun-mtu` and `mssfix`)
- Peek timeout latency (reduce from 2000ms to 200ms)
- iptables/mapping race condition (fix `|| true` in shell script)
- CGNAT UDP timeout (reduce keepalive from 10/120 to 5/30)

**Does NOT need `/gsd:research-phase`:** Pitfalls and fixes are already fully documented in PITFALLS.md and STACK.md with specific line references and config values.

---

### Phase 2: Dashboard UI Redesign

**Rationale:** The dashboard is a scaffold with no functional UI. Operators currently have no visibility into devices, proxy ports, or credentials. Dashboard work has no data-plane dependency — it reads from an API that is already running. Can be worked in parallel with Phase 1, or immediately after.

**Delivers:** A functional operator dashboard where a fleet operator can: see all devices with online/offline status and current IP, view proxy credentials per port, create and delete proxy ports, trigger manual IP rotation, and onboard new Android devices via QR code.

**Uses (from STACK.md):**
- shadcn/ui initialized with Card, Table, Badge, Dialog, Button, Select, Input, Chart components
- TanStack Query v5 for polling device status (refetchInterval: 5000ms) and REST mutations
- react-hook-form + zod for proxy creation dialog forms
- Native WebSocket hook connecting to existing `/ws` endpoint for real-time status

**Implements (from ARCHITECTURE.md):**
- Dashboard State Management pattern: REST calls for list data + WebSocket for live device events
- Device status cards (shadcn/ui Card grid with TanStack Query polling)
- Connection table (shadcn/ui Table with sortable rows and one-click credential copy)
- Port creation dialog (shadcn/ui Dialog + react-hook-form + zod validation)

**Addresses (from FEATURES.md):** Device online/offline status, current IP display, proxy credentials display, proxy port create/delete, QR code onboarding, basic traffic counter, device name/label.

**Avoids (from PITFALLS.md):**
- Connection cards showing credentials without device status (always show device online/offline indicator on every connection card)
- Dashboard refresh wiping form state (separate live data polling from form state)
- Showing raw VPN IP as device identifier (show device name/alias; raw IP in expanded detail only)
- No copy button on credentials (one-click copy for each credential field)

**Needs `/gsd:research-phase`:** No. Stack is well-documented with official shadcn/ui docs at HIGH confidence.

---

### Phase 3: Security Hardening

**Rationale:** The `PasswordPlain` field in the database is a critical pre-production blocker. So is the `script-security 3` in the OpenVPN config and the unauthenticated push API on port 8081. These must be fixed before any external customer gets access to the system.

**Delivers:** Credentials hashed at rest, push API protected, OpenVPN shell script security reduced, rate limiting on auth endpoint.

**Addresses (from PITFALLS.md):**
- `PasswordPlain` stored in DB → bcrypt with cost 10; update auth comparison in `openvpn_handler.go` line 56
- `via-env` credential passing → switch to `via-file` in auth-user-pass-verify
- `script-security 3` → reduce to `script-security 2`
- Push API port 8081 unauthenticated → add shared secret header validation
- No rate limiting on auth endpoint → add 5 attempts/minute per username in openvpn_handler.go Auth handler

**Does NOT need `/gsd:research-phase`:** Standard Go bcrypt patterns are well-documented; specific file references are in PITFALLS.md.

---

### Phase 4: Connection Management and Monitoring Features

**Rationale:** Once the core proxy workflow is functional and secure, add the operational features that make a device fleet manageable: IP rotation, carrier/battery telemetry, offline notifications, and IP history. These build on the existing device heartbeat and command delivery infrastructure.

**Delivers:** Manual IP rotation from dashboard, per-device carrier/signal/battery display, online/offline email notifications, IP history (last 5-10 IPs), proxy port on/off toggle.

**Addresses (from FEATURES.md):** Manual IP rotation (P1), per-device carrier/signal/battery display (P2), offline notifications via email (P1), IP history (P2), proxy on/off toggle (P1).

**Avoids (from PITFALLS.md):**
- Device offline shown as online (reduce `keepaliveTimeout` from 60s to 20-30s; surface last-seen timestamp)
- Aggressive rotation breaking long-running sessions (implement rotation at connection level, not device level; drain existing sessions before rotating)

**Needs `/gsd:research-phase`:** Possibly for email notification delivery (SMTP vs. transactional email service choice). Otherwise standard patterns.

---

### Phase 5: Production Hardening and Stability

**Rationale:** iptables rules surviving container restarts and the single `tunToUdp` goroutine bottleneck are deferred stability issues. Address these once the product is in active production use and scale pressure reveals the limits.

**Delivers:** Tunnel container restart leaves zero orphan rules, iptables state persisted to file for crash recovery, health check endpoint on push API, keepalive timeout tuned for carrier CGNAT compatibility.

**Addresses (from PITFALLS.md):**
- iptables rules surviving container restarts (flush `10.9.0.0/24` source rules on startup; persist DNAT state to file)
- Stale device keepalive state (reduce tunnel `keepaliveTimeout` to 20s; surface last-seen in dashboard)
- Single `tunToUdp` goroutine (per-client outbound channel with worker goroutine — measure first)

**Needs `/gsd:research-phase`:** No. Specific remediation steps are documented in PITFALLS.md.

---

### Phase 6: v1.x Feature Expansion

**Rationale:** After validated production usage, add the competitive features that differentiate from competitors: auto-rotation scheduler, IP whitelist auth, rotation URL, device grouping, and REST API for programmatic access.

**Delivers:** Automatic IP rotation with configurable interval, IP whitelist authentication per port, rotation URL for external automation, device grouping with bulk actions, REST API.

**Addresses (from FEATURES.md):** Automatic rotation (P2), IP whitelist auth (P2), rotation URL (P2), device grouping + bulk actions (P2), REST API (P2).

**Avoids (from PITFALLS.md):**
- API requires stable data model — do not expose API before port/device model is confirmed stable

**Needs `/gsd:research-phase`:** REST API design (endpoint structure, auth model, versioning) may benefit from a research pass. Auto-rotation on non-rooted devices has architectural constraints worth researching before committing to implementation.

---

### Phase Ordering Rationale

- **Phase 1 before everything:** OpenVPN throughput is the active blocker. No customer-facing product until this works. Dashboard work can be started in parallel but is meaningless in production while the core data plane is broken.
- **Phase 2 (dashboard) after Phase 1 or concurrent:** Dashboard has zero data-plane dependency — it reads from an API that works. Build order from ARCHITECTURE.md confirms dashboard can be built any time API is up.
- **Phase 3 (security) before external customer access:** `PasswordPlain` is a critical pre-production blocker. Must be resolved before the system is exposed to untrusted users.
- **Phase 4 (monitoring) after core proxy works:** Rotation, notifications, and telemetry are only useful once devices are online and proxying traffic.
- **Phase 5 (hardening) as production reveals limits:** Stability issues compound at scale; address when scale pressure makes them visible.
- **Phase 6 (expansion) after validated production usage:** API design should reflect a stable internal model; device grouping requires meaningful fleet size.

### Research Flags

Phases needing deeper research during planning:
- **Phase 1 (OpenVPN fix):** No additional research needed — PITFALLS.md and STACK.md provide specific config values and code line references. Diagnose with tcpdump first to confirm which bottleneck (latency vs. bandwidth) dominates.
- **Phase 6 (REST API):** API versioning strategy and auth model (API key vs. JWT) need a brief research pass before committing to endpoint design. Auto-rotation on non-rooted Android devices needs architecture research (airplane mode toggle vs. carrier API vs. root-only options).

Phases with standard patterns (skip research-phase):
- **Phase 2 (dashboard):** shadcn/ui + TanStack Query patterns are at HIGH confidence from official docs. Use STACK.md installation instructions directly.
- **Phase 3 (security):** Go bcrypt is standard; specific code locations documented in PITFALLS.md. No research needed.
- **Phase 4 (monitoring):** Builds on existing device heartbeat and command delivery infrastructure. Standard patterns.
- **Phase 5 (hardening):** iptables cleanup and Go goroutine patterns are well-understood. PITFALLS.md has specific remediation steps.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM | OpenVPN tuning: MEDIUM (community sources, multiple agreeing); dashboard: HIGH (official shadcn/ui + TanStack docs); kylemanna deprecation: MEDIUM (community assessment) |
| Features | MEDIUM | Competitor features verified from official sites and help docs; some nuances from third-party reviews at LOW confidence |
| Architecture | HIGH | Based on direct codebase analysis of `tunnel/main.go`, `transparentproxy/proxy.go`, `service/`, `openvpn/` — not inferred |
| Pitfalls | HIGH | Critical pitfalls from direct code inspection with specific file and line references; performance patterns from MEDIUM-confidence external guides |

**Overall confidence:** MEDIUM-HIGH. Architecture is solid (direct code analysis). Stack additions are well-documented. Feature gaps are factual (dashboard is a scaffold, specific features not implemented). OpenVPN tuning recommendations converge across multiple community sources. Main uncertainty is whether connection pooling alone resolves the throughput issue or whether there is an additional bottleneck in the Android HTTP proxy or cellular network.

### Gaps to Address

- **Actual throughput bottleneck cause:** Research identifies 5 candidate causes (no connection pooling, MTU mismatch, buffer settings, peek timeout, CGNAT). The fix order should be driven by tcpdump diagnosis, not applied all at once. Phase 1 should start with a measurement step before applying config changes.
- **Android HTTP proxy throughput ceiling:** The Android device's HTTP proxy on port 8080 is a hard bottleneck — its throughput ceiling per device is unknown without testing on a real device under load. This determines whether per-port connection limits are needed.
- **Carrier CGNAT behavior:** UDP blocking and CGNAT timeout behavior varies by carrier. Research gives general guidance (reduce keepalive to 5/30) but real-device testing on target carrier SIMs is the only way to validate. Flag for Phase 1 verification checklist.
- **kylemanna/openvpn replacement timing:** The image is unmaintained (OpenVPN 2.4.9, last update ~2020). OpenVPN 2.5+ adds `data-ciphers` and multi-threaded TLS. Migration is deferred but should be planned as a separate milestone after throughput is stable. The `ncp-ciphers` directive in the current config is a 2.4-era directive that may cause compatibility issues with newer clients.

---

## Sources

### Primary (HIGH confidence)
- Direct codebase analysis: `server/cmd/tunnel/main.go`, `server/internal/transparentproxy/proxy.go`, `server/internal/api/handler/openvpn_handler.go`, `server/deployments/openvpn/client-server.conf`, `server/internal/service/`
- shadcn/ui official docs — https://ui.shadcn.com/docs/installation/next
- shadcn/ui React Hook Form integration — https://ui.shadcn.com/docs/forms/react-hook-form
- shadcn/ui GitHub — Recharts v3 upgrade PR #8486 — https://github.com/shadcn-ui/ui/pull/8486
- TanStack Query v5 official docs — https://tanstack.com/query/v5/docs/react/overview
- Go SNI proxy patterns — https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go
- BBR congestion control — https://lwn.net/Articles/776090/

### Secondary (MEDIUM confidence)
- hamy.io OpenVPN throughput guide — fast-io, mssfix, compression — https://hamy.io/post/0003/optimizing-openvpn-throughput/
- linuxblog.io — fast-io, sndbuf/rcvbuf, txqueuelen — https://linuxblog.io/improving-openvpn-performance-and-throughput/
- ivanvari.com — txqueuelen 1000 documented 47→73 Mbps improvement — https://ivanvari.com/solving-openvpn-poor-throughput-and-packet-loss/
- OpenVPN community forums — sndbuf/rcvbuf=0 discussion — https://community.openvpn.net/openvpn/ticket/461
- OpenVPN MTU/fragment wiki — https://community.openvpn.net/MTU%20and%20Fragments
- kylemanna/docker-openvpn unmaintained status — https://github.com/kylemanna/docker-openvpn
- iProxy.online homepage and feature docs — https://iproxy.online/ and https://iproxy.online/blog/all-the-personal-account-features
- Proxidize official site and help docs — https://proxidize.com/proxy-builder/
- OpenVPN transparent proxy forum — https://forums.openvpn.net/viewtopic.php?t=32422
- iptables REDIRECT vs DNAT vs TPROXY — http://gsoc-blog.ecklm.com/iptables-redirect-vs.-dnat-vs.-tproxy/
- Tigera conntrack article — https://www.tigera.io/blog/when-linux-conntrack-is-no-longer-your-friend/
- Cellular CGNAT handling — https://damow.net/dealing-with-cellular-broadband-cgnat/

### Tertiary (LOW confidence)
- G2 iProxy reviews — aggregated user reviews, not feature documentation — https://www.g2.com/products/iproxy-online/reviews
- Dolphin Anty Proxidize review — third-party review — https://dolphin-anty.com/blog/en/review-of-proxidize-the-all-in-one-mobile-proxy-solution/
- mobileproxy.space farm guide — https://mobileproxy.space/en/pages/how-to-build-your-own-mobile-proxy-farm-a-step-by-step-beginners-guide.html

---

*Research completed: 2026-02-25*
*Ready for roadmap: yes*
