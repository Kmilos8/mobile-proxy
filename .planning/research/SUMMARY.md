# Project Research Summary

**Project:** PocketProxy v2.0 SaaS Transformation
**Domain:** Mobile proxy SaaS platform (B2B, operator-reseller model)
**Researched:** 2026-02-27
**Confidence:** HIGH

## Executive Summary

PocketProxy v2.0 is a transformation of an existing v1.0 operator tool into a multi-tenant SaaS platform where operators can resell mobile proxy access to their own customers. The existing codebase is mature and functional — Go/Gin backend with PostgreSQL, Next.js/shadcn/ui frontend, and a two-layer VPN architecture. The v2.0 work is purely additive: extend the auth system, layer in multi-tenant data isolation, build a customer-facing self-service portal, and add supporting features (landing page, white-label theming, device grouping, traffic logs, API docs). No major rewrites are needed, and the existing infrastructure requires only six new dependencies.

The recommended approach is to treat auth and multi-tenant isolation as the unbreakable foundation before any customer-facing work ships. The dependency is strict: every customer portal feature is only safe after every existing handler is audited for customer-scoping. Google OAuth is handled backend-first (Go verifies the ID token with `go-oidc`, not the frontend) to avoid a critical bypass vulnerability. Email verification uses a two-step flow (GET shows confirmation page, POST commits) to survive corporate email scanners that pre-fetch links. The rest of the features — landing page, white-label theming, device grouping, traffic logs, API docs — are independent and can be parallelized or sequenced by team priority after the foundation is solid.

The highest risk is tenant data leakage: all existing handlers return all records with no customer filtering. Missing a single query during the multi-tenant audit leaks one customer's proxy credentials to another, which is an existential SaaS failure. The second highest risk is the Google OAuth backend bypass, which is trivially exploitable if not correctly implemented. Both risks have clear, well-understood mitigations and must be addressed in the first two phases. All remaining risks (Turnstile enforcement, IP whitelist CIDR parsing, email scanner token consumption, CSS purge for theming, unbounded log storage) are standard SaaS patterns with known solutions documented in PITFALLS.md.

## Key Findings

### Recommended Stack

The existing stack requires only six additions, all lightweight and well-justified. The Go backend needs `golang.org/x/oauth2` and `coreos/go-oidc/v3` for Google OAuth, `resend-go/v3` for email delivery, and the `swaggo` toolchain for Swagger docs generation. The Next.js dashboard needs `@marsidev/react-turnstile` for the Cloudflare challenge widget and `next-themes` for runtime theme switching. Every other requirement — form handling, charting, component library, styling — is already installed and sufficient. Notably, NextAuth.js is explicitly excluded because it conflicts with the existing Go JWT system.

**Core technology additions:**
- `golang.org/x/oauth2` v0.35.0: Google OAuth2 token exchange — official Go client, handles token lifecycle
- `github.com/coreos/go-oidc/v3` v3.17.0: Google ID token verification — OpenID Connect discovery + JWT validation
- `github.com/resend/resend-go/v3` v3.1.1: Email sending — simpler than SendGrid, 3k free/month, handles DKIM automatically
- `github.com/swaggo/swag` + `gin-swagger` + `files`: Swagger UI — annotation-based, no manual spec maintenance
- `@marsidev/react-turnstile` v1.4.2: Cloudflare Turnstile widget — listed on official Cloudflare community resources
- `next-themes` v0.4.6: Theme switching — CSS variable based, already compatible with shadcn/ui

**Manual setup required before coding starts:**
- Register Google OAuth redirect URI in Google Cloud Console (blocks auth phase if delayed)
- Verify sending domain in Resend dashboard (blocks email verification if delayed)
- Create Cloudflare Turnstile site for site key and secret key

### Expected Features

The feature set is well-understood SaaS. Auth and multi-tenant isolation are table stakes — without them the product cannot function as a SaaS platform. Customer portal, landing page, and IP whitelist are expected at launch. White-label theming, device grouping, traffic logs, and API docs are competitive differentiators that round out the product. Google OAuth is a genuine differentiator — no current competitor (iProxy, IPRoyal, Proxidize) offers it.

**Must have (table stakes):**
- Email + password signup with email verification and Turnstile — prevents account abuse
- Google OAuth login — reduces signup friction, strong differentiator versus all competitors
- Password reset via email — basic account hygiene
- Multi-tenant isolation — customer sees only their own devices and connections
- Customer self-service portal — view credentials, rotate IPs, download .ovpn configs, view bandwidth
- IP whitelist authentication — per-connection CIDR-based IP auth without username/password
- Marketing landing page — product cannot be sold without it

**Should have (competitive differentiators):**
- White-label dashboard theming — custom logo, colors, brand name per operator
- Device grouping with bulk IP rotation — operational efficiency for multi-device setups
- Per-port traffic logs with time-range filtering — visibility into usage, aids billing conversations
- REST API documentation (Swagger UI) — enables customer API integrations

**Defer to v2.1+:**
- Magic link / passwordless login — differentiator but not blocking launch
- 2FA/TOTP — adds complexity, not table stakes at launch scale
- Per-customer rate limiting and quotas — needed eventually, not at launch
- Custom domain support for white-label — CORS complexity, defer until operator demand validates
- Real-time traffic graphs (WebSocket) — basic charts sufficient for launch
- SDK generation from OpenAPI — defer until API is stable and adoption justified

### Architecture Approach

The architecture is extend-in-place, not rewrite. The existing Gin router, JWT middleware, PostgreSQL schema, and Next.js layout are all preserved. New auth endpoints extend the existing `AuthService`. A new `/api/portal/*` route group with `RequireRole("customer")` middleware provides customer-scoped access via a dedicated `PortalHandler` that applies `WHERE customer_id = $X` filtering on every query. The `JWTClaims` struct gains a `customer_id` field. White-label theming uses CSS custom property overrides at runtime — shadcn/ui already uses CSS variables so no component rewriting is needed. Device grouping adds two tables and reuses the existing heartbeat command delivery path for bulk operations. Traffic logs require no new infrastructure; the existing `bandwidth_logs` table already holds per-connection byte data.

**Major components and changes:**
1. Auth extension — new `AuthService` methods (`Signup`, `VerifyEmail`, `GoogleLogin`, `ForgotPassword`, `ResetPassword`), `TurnstileMiddleware`, `EmailService` wrapping Resend, new DB columns on `users` table
2. Multi-tenant isolation — `customer_id` added to `JWTClaims`, `WHERE customer_id` filter on all customer-facing repo methods, separate `/api/portal/*` route group
3. Customer portal — new `PortalHandler`, new Next.js portal pages (`/portal/devices`, `/portal/connections`, `/portal/settings`)
4. White-label theming — new `tenant_themes` table, `GET /api/theme` endpoint, `ThemeProvider` React component, operator theme settings page
5. Device grouping — new `DeviceGroupService/Repo/Handler`, two new tables (`device_groups`, `device_group_members`), bulk commands via existing heartbeat path
6. Traffic logs — query existing `bandwidth_logs` table, add 30-day retention cron, add portal chart UI using existing recharts

### Critical Pitfalls

1. **Tenant data leak (P1, CRITICAL)** — All existing handlers return all records with no customer filtering. Prevention: create a completely separate `/api/portal/*` route group, never expose admin endpoints to customer role, every portal repo method must include `WHERE customer_id = $X`, mandatory code review checklist before any customer gets portal access.

2. **Google OAuth backend bypass (P2, CRITICAL)** — Backend must independently verify the Google ID token using `go-oidc`, never trust a user profile sent from the frontend. Backend verifies token signature, audience (must match client ID), and expiry. Use a separate `/api/auth/google` endpoint — do not mix with email/password login.

3. **Email verification consumed by corporate scanners (P3, HIGH)** — Microsoft Defender, Proofpoint, and Barracuda pre-fetch GET links in emails, consuming single-use tokens before the user clicks. Prevention: two-step flow — GET shows "Confirm your email" page with a button, POST commits the verification. Set token expiry to 24 hours, not 15 minutes.

4. **Unverified accounts accessing proxies (P4, HIGH)** — Check `email_verified = true` before allowing any proxy API calls. Google OAuth users are auto-verified. Show a clear email verification gate in portal UI before granting access.

5. **Turnstile widget without server-side verification (P5, HIGH)** — The React widget alone stops nothing. Go signup/login handlers must call `challenges.cloudflare.com/turnstile/v0/siteverify` and reject requests where `success` is not `true`.

6. **IP whitelist CIDR bypass (P6, HIGH)** — String comparison fails for CIDR ranges and is trivially bypassed. Use `net.ParseCIDR()` + `Contains()`. Configure Gin's `SetTrustedProxies()` — default trusts all forwarded headers.

## Implications for Roadmap

Based on the dependency graph in FEATURES.md and the build order confirmed in ARCHITECTURE.md, phases 1 and 2 are strictly sequential and non-negotiable. Phases 3 through 7 are largely independent and can be reordered by business priority.

### Phase 1: Auth Foundation

**Rationale:** Nothing else in the product is safely buildable without auth. Google OAuth redirect URI must be registered in Google Cloud Console on day one or it blocks this entire phase. Resend domain verification must be done before coding email delivery.
**Delivers:** Working signup flow (email + Google), email verification gate, password reset, Turnstile protection on all public forms, JWT with `customer_id` claim, new DB columns on `users` table
**Addresses:** All Auth & Signup features from FEATURES.md
**Avoids:** P2 (OAuth bypass — use `go-oidc` server-side), P3 (email scanner — two-step GET/POST flow), P4 (unverified access gate), P5 (Turnstile server-side verification), P8 (pre-registration hijack — duplicate email check), P9 (SPF/DKIM — handled by Resend after domain verification)

### Phase 2: Multi-Tenant Isolation

**Rationale:** Must be 100% complete before any customer-facing endpoint is exposed. A single missed query is a data leak. This is the highest-risk phase in the entire roadmap — the risk is execution discipline, not technical complexity.
**Delivers:** `customer_id` in JWT, `WHERE customer_id` filter on all customer-facing handlers, separate `/api/portal/*` route group with `RequireRole("customer")` enforcement, mandatory code audit of every handler that returns data
**Addresses:** Multi-Tenant features from FEATURES.md
**Avoids:** P1 (tenant data leak — the entire phase exists to prevent this), P7 (cross-customer device access — verify `device.customer_id == jwt.customer_id` in portal)

### Phase 3: Customer Self-Service Portal

**Rationale:** Depends on Phase 1 (auth) and Phase 2 (isolation). Once the foundation is safe, the portal largely reuses existing APIs with customer-scoped JWT — the heavy lifting is in the frontend pages.
**Delivers:** Customer portal pages (devices, connections, settings), credential display, .ovpn download, bandwidth view, IP rotation trigger from portal
**Addresses:** Customer Self-Service Portal features from FEATURES.md
**Avoids:** P7 (cross-customer device access — do not allow customers to specify arbitrary device IDs)

### Phase 4: Landing Page + IP Whitelist

**Rationale:** Both are independent of the auth/tenant foundation. Landing page is pure frontend with zero backend changes. IP whitelist enforcement uses data already in the `ProxyConnection` model (`IPWhitelist []string` field exists). Both are table stakes for launch.
**Delivers:** Marketing landing page at `/` (hero, features, pricing CTA, signup/login buttons), IP whitelist management UI on connection detail page, backend CIDR enforcement in tunnel server
**Addresses:** Marketing Landing Page and IP Whitelist Auth features from FEATURES.md
**Avoids:** P6 (CIDR bypass — use `net.ParseCIDR()` + `Contains()`, not string matching; configure `SetTrustedProxies()` in Gin)

### Phase 5: White-Label Theming

**Rationale:** Depends on multi-tenant (operator must exist as an entity in the system). Scope is per-operator theming only — per-customer theming is explicitly deferred as too complex.
**Delivers:** `tenant_themes` table, `GET /api/theme` endpoint, `ThemeProvider` React component, operator theme settings page (logo upload, color picker, brand name)
**Uses:** `next-themes` v0.4.6 for theme switching; CSS variable overrides on `:root`
**Avoids:** P10 (Tailwind CSS purge — override CSS variables on `:root`, not class names; test with production build not just dev server), P11 (CORS for custom domains — use same-origin architecture or dynamic CORS from `tenant_themes` table if custom domains are needed)

### Phase 6: Device Grouping + API Docs

**Rationale:** Both are independent of the customer portal. Device grouping adds new tables and reuses the existing heartbeat command path. API docs are annotation-only changes with no logic changes. Group these together as operator-facing tooling.
**Delivers:** Named device groups, bulk IP rotation per group, bulk enable/disable auto-rotation per group, Swagger UI at `/swagger/` with all endpoints documented and JWT auth flow explained
**Uses:** `swaggo` toolchain — annotations on existing handler methods, one new `/swagger/*` route
**Addresses:** Device Grouping and REST API Documentation features from FEATURES.md

### Phase 7: Per-Port Traffic Logs

**Rationale:** Last because it needs a validated retention policy in place from day one to prevent unbounded storage growth. The underlying data already exists in `bandwidth_logs` — this phase is primarily a UI and API exposure effort.
**Delivers:** Traffic history chart in customer portal (existing recharts), time-range filtering (24h/7d/30d), export endpoint, 30-day retention cron job (`DELETE FROM bandwidth_logs WHERE created_at < NOW() - INTERVAL '30 days'`)
**Addresses:** Per-Port Traffic Logs features from FEATURES.md
**Avoids:** P12 (unbounded storage — retention policy and `created_at` index must ship with first query exposure)

### Phase Ordering Rationale

- Phases 1 and 2 are strictly sequential and non-negotiable. The entire SaaS model is unsafe until both are complete.
- Phase 3 (customer portal) depends on Phases 1 and 2 but unlocks the core customer value proposition.
- Phases 4, 5, 6, and 7 are largely independent and can be reordered by business priority. The suggested order prioritizes customer-visible value (landing page, whitelist) over operator tooling (theming, groups) over developer tooling (API docs, logs).
- The marketing landing page has zero dependency on auth being complete and can be built during Phase 1 in parallel if resources allow.

### Research Flags

No phase in this roadmap requires a `/gsd:research-phase` deeper research sprint. All patterns were resolved during initial research.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Auth):** Two-step email verification pattern, `go-oidc` ID token verification, and Turnstile middleware are all fully documented. Implement directly from research findings in STACK.md and ARCHITECTURE.md.
- **Phase 2 (Multi-tenant):** App-level scoping with `WHERE customer_id` is a well-established pattern. The risk is execution discipline (audit completeness), not technical uncertainty.
- **Phase 3 (Portal):** Reuses existing handler patterns with scoped JWT. No novel patterns.
- **Phase 4 (Landing + Whitelist):** Static Next.js page and `net.ParseCIDR()` CIDR enforcement are standard.
- **Phase 5 (Theming):** CSS variable overrides at `:root` with `next-themes` is well-documented.
- **Phase 6 (Groups + Docs):** New join tables and annotation-based Swagger are straightforward.
- **Phase 7 (Logs):** Existing data + recharts charts + DELETE cron — no novel patterns.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All libraries verified against pkg.go.dev and npm registries with exact version numbers published as of 2026-02-27 |
| Features | HIGH (table stakes) / MEDIUM (differentiators) | SaaS auth and multi-tenant patterns are well-established; proxy-specific differentiators (device grouping UX, traffic log granularity) less documented in the wild |
| Architecture | HIGH | Based on direct codebase inspection of existing models, handlers, and auth system; all integration points are explicit and concrete |
| Pitfalls | HIGH (P1–P2) / MEDIUM (P3–P12) | P1 and P2 confirmed by direct code inspection (handlers verified to return all records); P3–P12 from community consensus across multiple sources |

**Overall confidence:** HIGH

### Gaps to Address

- **White-label custom domain CORS:** If operators request custom domain support (CNAME), the CORS approach needs a decision before Phase 5 ships: dynamic CORS from `tenant_themes` table vs. same-origin architecture. Decide during Phase 5 planning if any operator requests it.
- **Customer quota enforcement:** Per-customer limits on max devices and connections are a differentiator. The enforcement mechanism (DB column vs. plan-tier vs. Stripe integration) needs a decision during Phase 2 planning if quota enforcement is included in launch scope.
- **Traffic log granularity:** The `bandwidth_logs` table has per-connection byte data but the exact schema and refresh interval were not fully audited. Validate that the schema supports time-range filtering before beginning Phase 7 implementation.
- **Google Cloud Console setup:** Google OAuth client ID registration must happen before Phase 1 coding starts. If the Google Cloud Console project does not already exist, this is a day-one prerequisite that can block the entire auth phase.

## Sources

### Primary (HIGH confidence)
- Direct codebase inspection (`server/internal/`, `dashboard/`, `server/cmd/tunnel/`) — verified existing models, handlers, auth system structure, `ProxyConnection.IPWhitelist` field existence
- pkg.go.dev — `golang.org/x/oauth2` v0.35.0, `coreos/go-oidc/v3` v3.17.0, `resend-go/v3` v3.1.1, `swaggo` toolchain versions
- npmjs.com — `@marsidev/react-turnstile` v1.4.2, `next-themes` v0.4.6 versions
- Cloudflare Turnstile docs — server-side `siteverify` endpoint, widget integration pattern
- go-oidc GitHub — ID token verification pattern and audience validation

### Secondary (MEDIUM confidence)
- SaaS auth pattern consensus — two-step email verification to survive corporate email scanner pre-fetching (Microsoft Defender, Proofpoint, Barracuda)
- shadcn/ui docs — CSS variable theming approach confirming shadcn/ui uses `--primary`, `--secondary` etc. at `:root`
- Gin framework docs — `SetTrustedProxies()` configuration for `X-Forwarded-For` header trust
- Competitor feature matrix (iProxy, IPRoyal, Proxidize) — verified from official sites and help documentation

### Tertiary (LOW confidence / needs validation during implementation)
- Traffic log retention: 30-day default is a reasonable industry starting point but actual operator billing and compliance requirements may differ — validate with intended operators before Phase 7 ships

---
*Research completed: 2026-02-27*
*Ready for roadmap: yes*
