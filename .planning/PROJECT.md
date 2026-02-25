# PocketProxy

## What This Is

A mobile proxy platform that turns Android devices into mobile proxies, offering HTTP, SOCKS5, and OpenVPN direct access. Operates as both a direct proxy-selling business and a white-label SaaS product that others can deploy with their own device fleets.

## Core Value

Customers can reliably route traffic through real mobile devices via any of the three proxy protocols (HTTP, SOCKS5, OpenVPN), managed through a clean dashboard.

## Requirements

### Validated

- ✓ HTTP proxy port — functional, working
- ✓ SOCKS5 proxy port — functional, working
- ✓ Android device connects to server and serves as proxy endpoint
- ✓ Go backend with API server
- ✓ Dashboard scaffold (Next.js 14 + Tailwind CSS)
- ✓ Docker Compose deployment

### Active

- [ ] OpenVPN direct access — tunnel connects and shows correct mobile IP, but throughput is unusable (pages barely load, speed tests fail)
- [ ] Dashboard full redesign — device monitoring (status, IP, carrier, battery, signal) and proxy port management (create/delete HTTP, SOCKS5, VPN connections)
- [ ] All three protocols stable and production-ready

### Out of Scope

- Native iOS app — Android only for v1
- Payment/billing integration — handle externally for now
- Multi-tenant SaaS isolation — single-tenant first, SaaS architecture later

## Context

- Server is written in Go (go.mod in server/)
- Android app handles device-side proxy serving
- Dashboard is Next.js 14 + TypeScript + Tailwind CSS with recharts, lucide-react, qrcode.react
- Docker Compose orchestrates deployment
- OpenVPN configs are generated for client connections
- Similar to iProxy in concept — mobile proxy platform
- OpenVPN issue: tunnel establishes, correct IP shows, but bandwidth is severely limited (can't load pages fully, speed tests fail)

## Constraints

- **Protocol**: Must support HTTP, SOCKS5, and OpenVPN — all three are customer-facing
- **Platform**: Android devices only for v1
- **Deployment**: Docker Compose based
- **Dashboard**: Next.js 14 + Tailwind (existing stack, full redesign not rewrite)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go for server | Already built, performant for proxy workloads | — Pending |
| Next.js + Tailwind for dashboard | Already scaffolded, modern stack | — Pending |
| OpenVPN for direct access | Industry standard for tunnel-based proxy access | — Pending |
| Fix protocols before dashboard | Stable proxy = core product, dashboard is management layer | — Pending |

---
*Last updated: 2026-02-25 after initialization*
