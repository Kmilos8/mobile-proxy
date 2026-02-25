# Stack Research

**Domain:** Mobile proxy platform — OpenVPN throughput + dashboard redesign
**Researched:** 2026-02-25
**Confidence:** MEDIUM (OpenVPN tuning: HIGH based on official forums + verified patterns; dashboard: HIGH based on official docs; kylemanna deprecation: MEDIUM based on community sources)

---

## Context: What Already Exists (Do Not Replace)

This is a subsequent milestone. The following stack is in production and must not be replaced:

| Component | Technology | Version | Status |
|-----------|-----------|---------|--------|
| Go backend | Go + Gin | 1.23 / gin 1.9.1 | Working |
| Database | PostgreSQL | 16 (Docker) | Working |
| Tunnel server | Go (custom UDP + TUN) | — | Working |
| Dashboard | Next.js + TypeScript + Tailwind | 14.1.0 / TS 5.3 | Scaffold only |
| Deployment | Docker Compose | 3.8 | Working |
| Device proxy | Android app | — | Working |
| VPN (device) | OpenVPN server on port 1194 | kylemanna/openvpn 2.4 | Working |
| VPN (client) | OpenVPN server on port 1195 | kylemanna/openvpn 2.4 | Broken throughput |

Research below covers only the two active problems: (1) OpenVPN throughput and (2) dashboard UI libraries.

---

## Recommended Stack

### OpenVPN Performance: What to Fix

The existing `client-server.conf` already has partial tuning. The throughput problem is not a missing library — it is misconfigured or missing OpenVPN parameters. The fix is configuration changes, not new software.

#### Root Cause Diagnosis (from code review + research)

The architecture is:

```
Customer OpenVPN client
  → OpenVPN server (tun1, 10.9.0.0/24, port 1195)
  → iptables REDIRECT → transparent proxy (port 12345)
  → HTTP CONNECT → device's HTTP proxy (192.168.255.y:8080)
  → Android cellular data
```

There are multiple hops where throughput degrades:
1. OpenVPN UDP buffer settings (already present but may not be pushed correctly)
2. The transparent proxy is TCP-only — UDP traffic (DNS, QUIC) is blocked/rejected
3. MTU mismatch: OpenVPN tun1 MTU not set, default 1500 may cause fragmentation over cellular
4. `mssfix 1400` in server config but NOT in client `.ovpn` — must be in both
5. No `fast-io` directive in `client-server.conf` (applies to UDP mode)
6. No `txqueuelen` in the OpenVPN container's tun interface

#### Required OpenVPN Config Changes

**In `client-server.conf` (server-side):**

| Parameter | Current | Recommended | Why |
|-----------|---------|-------------|-----|
| `sndbuf` | 524288 (512KB) | `0` | Let OS/kernel pick optimal; hard values can hurt on some kernels. Known to fix 5 Mbps→60 Mbps issues. |
| `rcvbuf` | 524288 | `0` | Same reason — OS-managed buffers are better tuned for modern Linux |
| `push "sndbuf ..."` | 524288 | `push "sndbuf 0"` | Must match server side |
| `push "rcvbuf ..."` | 524288 | `push "rcvbuf 0"` | Must match server side |
| `txqueuelen` | 1000 | `1000` | Correct — keep as-is |
| `fast-io` | missing | Add `fast-io` | Non-blocking UDP writes; 5-10% CPU gain on UDP mode |
| `mssfix` | 1400 | `1300` | More conservative for cellular (variable MTU); prevents fragmentation |
| `fragment` | missing | Do NOT add | Fragment adds 4-byte overhead and reassembly latency; mssfix alone is preferred |
| `compress` | missing | Add `compress lz4-v2` | Compresses non-encrypted HTTP content; negligible overhead for already-compressed data |
| `tun-mtu` | 1500 | Keep `1500` | Default is correct for TUN; don't change without specific evidence |

**In generated `.ovpn` client configs:**

| Parameter | Current | Recommended | Why |
|-----------|---------|-------------|-----|
| `sndbuf` | 524288 | `sndbuf 0` | Pushed from server, but explicit in file overrides push — set to 0 |
| `rcvbuf` | 524288 | `rcvbuf 0` | Same |
| `mssfix` | missing | `mssfix 1300` | Cannot be pushed; must be in client config explicitly |
| cipher | AES-128-GCM | Keep | AES-128-GCM is correct for throughput (2x faster than AES-256 on most CPUs) |

**Kernel sysctl (already applied in tunnel/main.go — verify these are active):**

The tunnel server already sets BBR congestion control and enlarged TCP buffers via sysctl in `configureTUN()`. Verify this is actually taking effect inside the Docker container (requires `privileged: true` or `SYS_ADMIN` capability). The OpenVPN container does NOT apply these — they need to be applied on the host or in the OpenVPN container entrypoint.

| Sysctl | Value | Why |
|--------|-------|-----|
| `net.core.rmem_max` | 16777216 | 16MB receive socket buffer maximum |
| `net.core.wmem_max` | 16777216 | 16MB send socket buffer maximum |
| `net.ipv4.tcp_congestion_control` | bbr | BBR handles mobile packet loss gracefully (model-based, not loss-based) |
| `net.ipv4.tcp_mtu_probing` | 1 | Enables path MTU discovery — critical for variable mobile MTU |

#### What NOT to do for OpenVPN

| Avoid | Why | Instead |
|-------|-----|---------|
| `fragment N` | Adds 4-byte overhead per packet + reassembly latency | Use `mssfix 1300` alone |
| Hard-coded `sndbuf 524288` | Can hurt performance on some kernels worse than defaults | Use `sndbuf 0` |
| `proto tcp` for the client server | TCP-over-TCP doubles retransmission; catastrophic for mobile | Keep `proto udp` |
| `ncp-ciphers AES-128-GCM` (restrictive) | Breaks compatibility with some clients | Use `data-ciphers AES-128-GCM:AES-256-GCM:CHACHA20-POLY1305` for OpenVPN 2.5+ |
| Switching to WireGuard | Requires rewriting the Android app and server | Fix OpenVPN config first |
| Changing kylemanna/openvpn image now | Risky mid-debug; fix config first, then evaluate image | Replace image only after throughput is fixed |

---

### Dashboard UI: New Libraries to Add

The existing stack (Next.js 14 + TypeScript + Tailwind + Recharts + lucide-react) is the correct foundation. The redesign requires adding UI component and form management libraries.

#### Core Technologies (existing — keep)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Next.js | 14.1.0 | App framework | Already deployed; App Router used |
| TypeScript | 5.3.3 | Type safety | Already present |
| Tailwind CSS | 3.4.1 | Styling | Already present |
| Recharts | 3.7.0 (in package.json as ^3.7.0) | Charts | Already present; shadcn/ui chart uses Recharts |
| lucide-react | 0.312.0 | Icons | Already present |

**Note on Recharts version:** The package.json has `recharts: "^3.7.0"`. shadcn/ui's chart component targets Recharts v3 (PR #8486 in progress as of 2025). This is compatible. HIGH confidence.

#### New Libraries to Add

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| `shadcn/ui` | latest (CLI: `shadcn@latest`) | Component library | Official recommendation for Next.js + Tailwind; built on Radix UI; copy-paste model means no version lock-in; has Card, Table, Badge, Dialog, Select, Button all needed for proxy dashboard |
| `@radix-ui/react-*` | (installed by shadcn init) | Accessible primitives | shadcn/ui dependency; provides accessible dropdowns, dialogs, tooltips |
| `react-hook-form` | ^7.54.0 | Form management | shadcn/ui official integration; needed for proxy creation/edit forms |
| `zod` | ^3.24.0 | Schema validation | Pairs with react-hook-form + shadcn; validates proxy config inputs client + server side |
| `@hookform/resolvers` | ^3.10.0 | Connects zod to react-hook-form | Required bridge package |
| `@tanstack/react-query` | ^5.67.0 | Data fetching + polling | Server state management; replaces manual fetch in useEffect; provides refetchInterval for device status polling every 5-10s; handles stale/loading/error states cleanly |
| `class-variance-authority` | ^0.7.1 | Component variants | shadcn/ui dependency; already likely installed if shadcn init runs |
| `cmdk` | ^1.0.4 | Command palette | shadcn/ui Command component; useful for device search |

**Note:** `clsx` and `tailwind-merge` are already in package.json — shadcn/ui needs these and they are present.

#### What NOT to Add

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Tremor | Built on Recharts but is a higher-level abstraction; reduces flexibility; team already has Recharts directly | shadcn/ui chart (thin Recharts wrapper) |
| Next.js 15 upgrade | App Router patterns change; existing code targets Next.js 14; upgrade is non-trivial | Stay on 14.1.x for this milestone |
| Tailwind v4 | shadcn/ui migrated tailwindcss-animate in March 2025 for v4; existing config is v3; migration is a separate effort | Stay on Tailwind v3 |
| Redux / Zustand | Overkill for dashboard state; TanStack Query handles server state; React state handles UI state | TanStack Query + useState |
| Socket.io | WebSocket already exists via gorilla/websocket on the backend; adding Socket.io requires server changes | Use native WebSocket with a custom hook |
| React Query v4 | v5 is current; v4 is EOL | @tanstack/react-query v5 |
| SWR | TanStack Query v5 is more featured for this use case; both are valid but pick one | TanStack Query v5 |

#### Supporting Libraries (dev)

| Library | Version | Purpose | Notes |
|---------|---------|---------|-------|
| `@types/react` | ^18.3.0 | React types | Upgrade from 18.2.x — minor |
| `eslint-config-next` | 14.x | Linting | Already present via next lint |

---

### Infrastructure: No Changes Required

The existing Docker Compose infrastructure (postgres, openvpn-client, tunnel, api, worker, dashboard, nginx) is the correct deployment. No new infrastructure components are needed for this milestone.

The kylemanna/openvpn image is unmaintained (last updated ~2020, OpenVPN 2.4.9). However, replacing it during an active throughput investigation adds risk. **Defer image replacement to a future milestone.** The config changes needed work within the existing image constraints.

---

## Installation

```bash
# Navigate to dashboard directory
cd dashboard

# Initialize shadcn/ui (run in the existing Next.js project)
npx shadcn@latest init

# Add specific shadcn/ui components needed for dashboard redesign
npx shadcn@latest add card
npx shadcn@latest add table
npx shadcn@latest add badge
npx shadcn@latest add button
npx shadcn@latest add dialog
npx shadcn@latest add select
npx shadcn@latest add input
npx shadcn@latest add label
npx shadcn@latest add separator
npx shadcn@latest add skeleton
npx shadcn@latest add tooltip
npx shadcn@latest add chart

# TanStack Query and form libs
npm install @tanstack/react-query@^5.67.0
npm install react-hook-form@^7.54.0 zod@^3.24.0 @hookform/resolvers@^3.10.0
```

**Note:** `npx shadcn@latest init` will ask about Tailwind config and install `class-variance-authority`, `clsx` (already present), `tailwind-merge` (already present), `@radix-ui/react-slot`. Accept defaults.

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| Next.js 14.1.0 | React 18.x | Do NOT upgrade to React 19 without Next.js 15 |
| shadcn/ui (latest) | Next.js 14 + Tailwind 3 | Works on v3; Tailwind v4 support is separate |
| Recharts ^3.7.0 | shadcn/ui chart | shadcn chart PR for v3 is merged; compatible |
| @tanstack/react-query v5 | React 18+ | v5 requires React 18 minimum; compatible |
| zod ^3.24.0 | @hookform/resolvers ^3.x | Must use resolvers v3+ for zod v3 |
| react-hook-form ^7.54.0 | @hookform/resolvers ^3.10.0 | Match resolvers to RHF 7.x |

---

## Stack Patterns by Variant

**For device status cards (real-time):**
- Use TanStack Query `useQuery` with `refetchInterval: 5000` for polling
- Use native WebSocket hook (the backend already provides `/ws`) for live updates
- Display signal strength, carrier, battery, IP in a Card grid layout using shadcn/ui Card

**For proxy port management (CRUD):**
- Use TanStack Query `useMutation` for create/delete operations
- Use react-hook-form + zod for the creation dialog form
- shadcn/ui Dialog + Form components for the modal

**For connection list (table):**
- Use shadcn/ui Table component
- Data fetched with TanStack Query, cached for 30s, refetch on window focus
- Download `.ovpn` link triggers the existing `GET /api/connections/:id/ovpn` endpoint

**For charts (bandwidth history):**
- Use shadcn/ui chart component (wraps Recharts)
- AreaChart for bandwidth over time; already using Recharts directly

---

## Sources

- OpenVPN community forums (multiple threads): sndbuf/rcvbuf=0 discussion — MEDIUM confidence (community, verified by multiple sources)
  - https://github.com/angristan/openvpn-install/issues/352
  - https://community.openvpn.net/openvpn/ticket/461
- ivanvari.com — txqueuelen=1000 documented 47→73 Mbps improvement — MEDIUM confidence (single case study, verified by pattern)
  - https://ivanvari.com/solving-openvpn-poor-throughput-and-packet-loss/
- hamy.io OpenVPN throughput guide — fast-io, mssfix, compression recommendations — MEDIUM confidence
  - https://hamy.io/post/0003/optimizing-openvpn-throughput/
- linuxblog.io — fast-io, sndbuf/rcvbuf 512000, txqueuelen 2000 — MEDIUM confidence
  - https://linuxblog.io/improving-openvpn-performance-and-throughput/
- community.openvpn.net MTU/fragment wiki — fragment vs mssfix trade-offs — MEDIUM confidence
  - https://community.openvpn.net/MTU%20and%20Fragments
- shadcn/ui official docs — installation, chart component, React Hook Form integration — HIGH confidence
  - https://ui.shadcn.com/docs/installation/next
  - https://ui.shadcn.com/docs/forms/react-hook-form
- shadcn/ui GitHub — Recharts v3 upgrade PR #8486 — HIGH confidence
  - https://github.com/shadcn-ui/ui/pull/8486
- TanStack Query v5 official docs — polling, useQuery/useMutation — HIGH confidence
  - https://tanstack.com/query/v5/docs/react/overview
- kylemanna/docker-openvpn GitHub — unmaintained status — MEDIUM confidence (community assessment)
  - https://github.com/kylemanna/docker-openvpn

---
*Stack research for: Mobile proxy platform — OpenVPN throughput + dashboard redesign*
*Researched: 2026-02-25*
