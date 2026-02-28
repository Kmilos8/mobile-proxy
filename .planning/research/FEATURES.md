# Features Research: PocketProxy v2.0 SaaS Transformation

**Mode:** Ecosystem — Feature Landscape
**Confidence:** HIGH for table stakes (well-documented SaaS patterns); MEDIUM for proxy-specific differentiators

## Feature Categories

### 1. Authentication & Signup

**Table stakes:**
- Email + password signup with validation
- Email verification (confirm your email before access)
- Google OAuth login/signup (one-click)
- Password reset via email link
- Cloudflare Turnstile on all public forms (signup, login, forgot-password)

**Differentiators:**
- Passwordless magic link login
- 2FA/TOTP

**Anti-features:**
- Allowing unverified accounts to use proxies (abuse vector)
- Trusting frontend OAuth assertions without backend verification (security hole)

**Complexity:** MEDIUM — auth is well-understood but touches every layer (DB, API, frontend, middleware)
**Dependencies:** None — foundation for everything else

### 2. Multi-Tenant Access

**Table stakes:**
- Customer sees ONLY their own devices, connections, and data
- Operator (admin) sees everything across all customers
- JWT claims carry `customer_id` for query scoping
- Every repository query filters by `customer_id` for customer role

**Differentiators:**
- Per-customer rate limiting
- Customer usage quotas (max devices, max connections)

**Anti-features:**
- Shared device pools between customers (defeats private proxy value)

**Complexity:** HIGH — must audit every existing handler/query to add customer scoping. Missing a single query = data leak between customers.
**Dependencies:** Auth (customer accounts must exist first)

### 3. Customer Self-Service Portal

**Table stakes:**
- Customer can view their assigned devices and connections
- Customer can see connection credentials (host, port, username, password)
- Customer can download .ovpn configs
- Customer can view bandwidth usage
- Customer can trigger IP rotation

**Differentiators:**
- Customer can create/delete their own connections (within operator-set limits)
- Customer can set auto-rotation intervals
- Customer can configure webhook notifications

**Anti-features:**
- Customer accessing device-level settings (reboot, rename, commands)
- Customer seeing other customers' data

**Complexity:** MEDIUM — reuses existing v1.0 APIs with customer-scoped JWT
**Dependencies:** Auth + Multi-tenant isolation

### 4. Marketing Landing Page

**Table stakes:**
- Hero section with value proposition
- Feature highlights (3 protocols, mobile IPs, dashboard)
- Pricing section or CTA
- Signup/login buttons
- Mobile responsive
- Professional design matching dashboard theme

**Differentiators:**
- Live demo or interactive preview
- Customer testimonials
- Speed test results / uptime stats

**Complexity:** LOW — static page, no backend integration
**Dependencies:** None (but signup buttons need auth system)

### 5. White-Label Dashboard Theming

**Table stakes:**
- Custom logo upload
- Custom brand colors (primary, accent)
- Custom page title / brand name

**Differentiators:**
- Custom domain support (CNAME)
- Custom email templates with operator branding
- Custom CSS injection

**Anti-features:**
- Per-customer theming (massive complexity) — keep it per-operator

**Complexity:** MEDIUM — CSS variables + DB storage for theme config
**Dependencies:** Multi-tenant (operator must exist as entity)

### 6. IP Whitelist Auth

**Table stakes:**
- Operator can add IP addresses/CIDRs to a proxy connection's whitelist
- Whitelisted IPs can use the proxy without username/password
- Whitelist management UI on connection detail page

**Differentiators:**
- Auto-detect client IP and offer to whitelist
- Whitelist expiry (temporary access)

**Anti-features:**
- Trusting X-Forwarded-For without validation (bypass vector)

**Complexity:** LOW-MEDIUM — `IPWhitelist` field already exists in data model, needs enforcement + UI
**Dependencies:** None

### 7. REST API Documentation

**Table stakes:**
- Swagger/OpenAPI spec generated from code annotations
- Interactive API explorer (Swagger UI)
- Authentication documented (JWT flow)
- All CRUD endpoints documented

**Differentiators:**
- API versioning (`/api/v1/`)
- SDK generation from OpenAPI spec
- Rate limiting documentation

**Complexity:** LOW — annotation-based with swaggo, no logic changes
**Dependencies:** None

### 8. Device Grouping & Bulk Actions

**Table stakes:**
- Create named device groups
- Add/remove devices from groups
- Bulk rotate IPs for all devices in a group
- Bulk enable/disable auto-rotation for a group

**Differentiators:**
- Group-level bandwidth aggregation
- Group assignment rules (auto-assign by carrier/region)

**Complexity:** MEDIUM — new tables + join logic + bulk command dispatch
**Dependencies:** None

### 9. Per-Port Traffic Logs

**Table stakes:**
- View per-connection traffic history (bytes in/out over time)
- Time-range filtering (last 24h, 7d, 30d)
- Export/download logs

**Differentiators:**
- Real-time traffic graph (WebSocket)
- Request-level logging (URLs, response codes)

**Anti-features:**
- Unbounded log retention (storage explosion) — MUST have TTL/retention policy
- Request-level content inspection (legal/privacy liability)

**Complexity:** LOW-MEDIUM — existing `bandwidth_logs` table has per-connection data, needs UI + API + retention policy
**Dependencies:** None

## Dependency Graph

```
Auth & Signup ──→ Multi-Tenant ──→ Customer Portal
                                ──→ White-Label Theming

Landing Page (independent)
Auto-rotation bug fix (independent)
IP Whitelist (independent)
API Docs (independent)
Device Grouping (independent)
Traffic Logs (independent)
```

## Suggested Phase Order

1. **Auth foundation** — signup, email verification, Google OAuth, Turnstile, password reset
2. **Multi-tenant isolation** — customer_id in JWT, query scoping in all handlers
3. **Customer self-service portal** — customer-facing UI reusing existing APIs
4. **Landing page + auto-rotation bug fix** — parallel with portal work
5. **White-label theming + IP whitelist** — operator-facing features
6. **API docs + versioning** — wrap existing APIs under `/api/v1/`
7. **Device grouping + bulk actions** — operator-side
8. **Per-port traffic logs** — validate demand, set retention policy

## Competitor Reference

| Feature | iProxy | IPRoyal | Proxidize | PocketProxy v2.0 |
|---------|--------|---------|-----------|-------------------|
| Self-signup | Yes | Yes | No | Planned |
| Google OAuth | No | No | No | Planned (differentiator) |
| Multi-tenant | Yes | Yes | Yes | Planned |
| White-label | No | No | Yes | Planned |
| IP Whitelist | Yes | Yes | Yes | Planned |
| API Docs | Yes | Yes | No | Planned |
| Device Groups | No | N/A | Yes | Planned |
| Traffic Logs | Basic | Yes | Yes | Planned |

---
*Researched: 2026-02-27 for v2.0 SaaS milestone*
