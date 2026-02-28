# PocketProxy

## What This Is

A mobile proxy platform that turns Android devices into mobile proxies, offering HTTP, SOCKS5, and OpenVPN direct access. Evolving from an operator-only tool into a customer-facing SaaS product where end customers can self-register, manage their own proxies, and operators can white-label the platform.

## Core Value

Customers can reliably route traffic through real mobile devices via any of the three proxy protocols (HTTP, SOCKS5, OpenVPN), managed through a clean dashboard.

## Current Milestone: v2.0 SaaS Platform

**Goal:** Transform PocketProxy from operator-only to a customer-facing SaaS product with self-signup, multi-tenant access, and a public landing page.

**Target features:**
- Marketing/sales landing page
- Customer self-signup (email + Google OAuth)
- Email verification on signup
- Cloudflare Turnstile bot protection
- Multi-tenant access (customers see only their proxies)
- Customer self-service portal
- White-label dashboard theming
- IP whitelist auth per proxy port
- REST API docs & versioning
- Device grouping with bulk actions
- Per-port traffic logs

## Requirements

### Validated

- ✓ HTTP proxy port — v1.0
- ✓ SOCKS5 proxy port — v1.0
- ✓ Android device connects to server and serves as proxy endpoint — v1.0
- ✓ Go backend with API server — v1.0
- ✓ Dashboard scaffold (Next.js 14 + Tailwind CSS) — v1.0
- ✓ Docker Compose deployment — v1.0
- ✓ OpenVPN direct access with usable throughput — v1.0 Phase 1
- ✓ Dashboard full redesign (shadcn/ui, device table, connection CRUD) — v1.0 Phase 2
- ✓ All three protocols stable and production-ready — v1.0 Phase 1
- ✓ Bcrypt auth (PasswordPlain removed) — v1.0 Phase 3
- ✓ Bandwidth enforcement per connection — v1.0 Phase 3
- ✓ Device offline/recovery webhooks — v1.0 Phase 4

### Active

- [ ] Marketing/sales landing page
- [ ] Customer self-signup with email and Google OAuth
- [ ] Email verification on signup
- [ ] Cloudflare Turnstile on login/signup
- [ ] Multi-tenant access (customer isolation)
- [ ] Customer self-service portal
- [ ] White-label dashboard theming
- [ ] IP whitelist auth per proxy port
- [ ] REST API documentation and versioning
- [ ] Device grouping with bulk actions
- [ ] Per-port traffic logs

### Out of Scope

- Native iOS app — Android only
- Payment/billing integration — handle externally (Stripe Checkout) until billing milestone
- Telegram bot — Dashboard + API + rotation links cover all use cases
- Real-time traffic interception — Legal/privacy liability, storage costs
- Shared IP pools — Defeats core value of private mobile proxies
- Geolocation targeting — Architecture assumes specific devices, not location pools
- SMS forwarding — Legal risk; requires deliberate TOS design
- Passive OS fingerprint spoofing — High complexity, limited audience

## Context

- Server is written in Go (go.mod in server/)
- Android app handles device-side proxy serving
- Dashboard is Next.js 14 + TypeScript + Tailwind CSS with shadcn/ui, recharts, lucide-react, qrcode.react
- Docker Compose orchestrates deployment on two VPSes (relay + dashboard)
- OpenVPN configs generated for client connections
- v1.0 shipped: protocol stability, dashboard redesign, security hardening, monitoring
- Auto-rotation exists but interval may not be firing correctly — needs investigation

## Constraints

- **Protocol**: Must support HTTP, SOCKS5, and OpenVPN — all three are customer-facing
- **Platform**: Android devices only
- **Deployment**: Docker Compose based, two VPSes
- **Dashboard**: Next.js 14 + Tailwind + shadcn/ui (extend, not rewrite)
- **Auth**: Google OAuth + email/password, Cloudflare Turnstile for bot protection

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go for server | Already built, performant for proxy workloads | ✓ Good |
| Next.js + Tailwind for dashboard | Already scaffolded, modern stack | ✓ Good |
| OpenVPN for direct access | Industry standard for tunnel-based proxy access | ✓ Good |
| Fix protocols before dashboard | Stable proxy = core product, dashboard is management layer | ✓ Good |
| v2.0 = SaaS transformation | Customer self-signup, multi-tenant, landing page — major scope change | — Pending |

---
*Last updated: 2026-02-27 after v2.0 milestone start*
