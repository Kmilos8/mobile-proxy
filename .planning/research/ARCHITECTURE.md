# Architecture Research: PocketProxy v2.0 SaaS Transformation

**Mode:** Ecosystem — Architecture integration
**Confidence:** HIGH (based on direct codebase inspection + authoritative sources)

## Existing Architecture (What We Have)

- **Go backend**: Gin router, JWT auth middleware, PostgreSQL via pgx
- **Models**: User (with role field), Device, ProxyConnection (has customer_id FK), BandwidthLog
- **Auth**: `AuthService` with bcrypt, `JWTClaims` with user_id/role, `AuthMiddleware` sets gin context
- **Dashboard**: Next.js 14 + shadcn/ui + Tailwind CSS, `api.ts` client, `auth.ts` token management
- **Deployment**: Two VPSes (relay: tunnel+OpenVPN, dashboard: API+Next.js+PostgreSQL)
- **Real-time**: WebSocket for device status updates

## Integration Plan

### 1. Auth System Extension

**Modified components:**
- `User` model — add fields: `email_verified bool`, `google_id *string`, `verification_token *string`, `verification_token_expires_at *time.Time`
- `JWTClaims` — add `customer_id *string` field
- `AuthService` — add methods: `Signup()`, `VerifyEmail()`, `GoogleLogin()`, `ForgotPassword()`, `ResetPassword()`
- `AuthMiddleware` — branch on role: admin/operator gets full access, customer gets scoped access
- `router.go` — add routes: `POST /api/auth/signup`, `POST /api/auth/google`, `POST /api/auth/verify-email`, `POST /api/auth/forgot-password`, `POST /api/auth/reset-password`

**New components:**
- `TurnstileMiddleware` — Gin middleware that validates Turnstile token on public routes
- `TurnstileService` — calls Cloudflare siteverify API
- `EmailService` — wraps Resend SDK for sending verification/reset emails

**Data flow:**
```
Email Signup:
  Frontend form + Turnstile → POST /api/auth/signup → Create user (unverified) → Send verification email via Resend → User clicks link → POST /api/auth/verify-email → Mark verified → Issue JWT

Google OAuth:
  Frontend Google Sign-In → Get ID token → POST /api/auth/google → Verify ID token with go-oidc → Find/create user (auto-verified) → Issue JWT
```

### 2. Multi-Tenant Isolation

**Approach:** Application-level scoping (NOT PostgreSQL RLS)

**Rationale:** RLS adds complexity with pgx connection pooling. App-level scoping is simpler, debuggable, and sufficient at this scale.

**Modified components:**
- `AuthMiddleware` — extract `customer_id` from JWT, set in Gin context
- ALL repository methods for customer-facing endpoints — add `WHERE customer_id = $X` filter
- `connection_repo.go` — already has `customer_id` column, needs filtering
- `device_repo.go` — needs `customer_id` association (devices assigned to customers by operator)

**Critical audit required:**
Every handler that returns data must be reviewed:
- `connHandler.List()` — currently returns ALL connections
- `deviceHandler.List()` — currently returns ALL devices
- `customerHandler.List()` — operator-only, not exposed to customers

**New: Customer-scoped API group:**
```go
portal := r.Group("/api/portal")
portal.Use(authMiddleware.RequireRole("customer"))
portal.GET("/connections", portalHandler.ListMyConnections)
portal.GET("/devices", portalHandler.ListMyDevices)
portal.POST("/connections/:id/rotate", portalHandler.RotateIP)
// etc.
```

### 3. Customer Self-Service Portal

**New components:**
- `PortalHandler` — customer-facing API handler (parallel to existing admin handlers)
- Next.js portal pages — `/portal/devices`, `/portal/connections`, `/portal/settings`
- Portal layout — reuses shadcn/ui components but with customer-scoped data

**Data flow:**
```
Customer logs in → JWT has role=customer, customer_id=X
→ Portal pages call /api/portal/* endpoints
→ PortalHandler queries with customer_id filter
→ Customer sees only their assigned devices/connections
```

### 4. Cloudflare Turnstile

**New components:**
- `TurnstileMiddleware` (Go) — validates `cf-turnstile-response` header
- Turnstile widget component (React) — wraps `@marsidev/react-turnstile`

**Integration:**
```go
// Apply only to public unauthenticated routes
auth := r.Group("/api/auth")
auth.POST("/signup", turnstileMiddleware.Validate(), authHandler.Signup)
auth.POST("/login", turnstileMiddleware.Validate(), authHandler.Login)
auth.POST("/forgot-password", turnstileMiddleware.Validate(), authHandler.ForgotPassword)
```

### 5. White-Label Theming

**Approach:** CSS custom properties overridden at runtime

**New components:**
- `tenant_themes` DB table — stores brand colors, logo URL, brand name per operator
- Theme API endpoint — `GET /api/theme` returns CSS variables for current tenant
- `ThemeProvider` component — injects CSS variables on `:root` at runtime
- Theme settings page — operator uploads logo, picks colors

**Integration with shadcn/ui:**
shadcn/ui already uses CSS variables (`--primary`, `--secondary`, etc.). Override these at runtime.

### 6. IP Whitelist Enforcement

**Modified components:**
- Tunnel server (`main.go`) — check client IP against connection's whitelist before allowing traffic
- `ProxyConnection` model — `IPWhitelist []string` already exists
- Connection detail page — add whitelist management UI

**Critical:** Use `net.ParseCIDR()` + `Contains()` for comparison, NOT string matching. Trust `X-Forwarded-For` only from known proxy IPs.

### 7. Device Grouping

**New tables:**
```sql
CREATE TABLE device_groups (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE device_group_members (
    group_id UUID REFERENCES device_groups(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, device_id)
);
```

**New components:**
- `DeviceGroupService` + `DeviceGroupRepo` + `DeviceGroupHandler`
- Bulk commands use existing `DeviceCommand` + heartbeat delivery path (fire-and-forget per device)

### 8. Per-Port Traffic Logs

**No new infrastructure needed.** Existing `bandwidth_logs` table already contains per-connection byte data.

**New:**
- Customer-accessible API endpoint to query bandwidth_logs by connection_id
- Portal UI chart (reuse existing recharts setup)
- Retention policy: `DELETE FROM bandwidth_logs WHERE created_at < NOW() - INTERVAL '30 days'` (cron or pg_cron)

### 9. Landing Page

**New:** Static Next.js page at `/` (currently redirects to `/devices`)

**No backend changes.** Pure frontend: hero section, feature highlights, pricing CTA, signup/login buttons.

## Database Migrations Summary

```sql
-- Auth extension
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN google_id TEXT UNIQUE;
ALTER TABLE users ADD COLUMN verification_token TEXT;
ALTER TABLE users ADD COLUMN verification_token_expires_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN password_reset_token TEXT;
ALTER TABLE users ADD COLUMN password_reset_expires_at TIMESTAMPTZ;

-- Tenant theming
CREATE TABLE tenant_themes (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    brand_name TEXT,
    logo_url TEXT,
    primary_color TEXT,
    accent_color TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Device grouping
CREATE TABLE device_groups (...);
CREATE TABLE device_group_members (...);

-- Bandwidth log retention
CREATE INDEX idx_bandwidth_logs_created_at ON bandwidth_logs(created_at);
```

## Suggested Build Order

1. DB migrations → JWT extension → Signup/email verification → Google OAuth → Turnstile
2. Customer portal (read-only) → Customer portal (writes) → Theming
3. Landing page → IP whitelist → Device grouping → API docs → Traffic logs

---
*Researched: 2026-02-27 for v2.0 SaaS milestone*
