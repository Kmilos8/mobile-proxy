# Stack Research: PocketProxy v2.0 SaaS Transformation

**Mode:** Ecosystem — Stack additions only
**Confidence:** HIGH (all libraries verified against pkg.go.dev and npm registries)

## New Libraries Required

### Go Backend

| Library | Version | Purpose | Rationale |
|---------|---------|---------|-----------|
| `golang.org/x/oauth2` | v0.35.0 | Google OAuth2 flow | Official Go OAuth2 client, handles token exchange |
| `github.com/coreos/go-oidc/v3` | v3.17.0 | Google ID token verification | OpenID Connect discovery + JWT validation |
| `github.com/resend/resend-go/v3` | v3.1.1 | Email sending (verification) | Simple API, 3k free/month, handles DKIM/deliverability |
| `github.com/swaggo/gin-swagger` | v1.6.1 | Swagger UI middleware for Gin | Standard Gin pattern for API docs |
| `github.com/swaggo/files` | v1.0.1 | Swagger static files | Required by gin-swagger |
| `github.com/swaggo/swag` (CLI) | v1.16.6 | Generate Swagger JSON from annotations | Annotation-based, no manual spec maintenance |

```bash
go get golang.org/x/oauth2@v0.35.0
go get github.com/coreos/go-oidc/v3@v3.17.0
go get github.com/resend/resend-go/v3@v3.1.1
go get github.com/swaggo/gin-swagger@v1.6.1
go get github.com/swaggo/files@v1.0.1
go install github.com/swaggo/swag/cmd/swag@v1.16.6
```

### Next.js Dashboard

| Library | Version | Purpose | Rationale |
|---------|---------|---------|-----------|
| `@marsidev/react-turnstile` | v1.4.2 | Cloudflare Turnstile widget | Listed on Cloudflare's official community resources |
| `next-themes` | v0.4.6 | Theme switching (white-label) | Recommended by shadcn/ui, CSS variable based |

```bash
npm install @marsidev/react-turnstile next-themes
```

### No Changes Needed

These are already installed and sufficient:
- react-hook-form, zod, @hookform/resolvers — form handling
- shadcn/ui — component library (already uses CSS variables for theming)
- Tailwind CSS — styling
- lucide-react — icons
- recharts — charts

## Stack Decisions

### Google OAuth: Go backend, NOT NextAuth.js

NextAuth.js conflicts with the existing Go JWT system. The correct pattern:
1. Frontend shows Google Sign-In button (Google's client-side SDK)
2. User authenticates with Google → gets ID token
3. Frontend POSTs ID token to Go backend `/api/auth/google`
4. Go backend verifies ID token with `go-oidc`, creates/finds user, issues PocketProxy JWT

### Email: Resend, not SendGrid

- Simpler API surface
- Better developer experience
- 3,000 free emails/month covers verification use case
- Handles deliverability/DKIM automatically
- v3.1.1 published Feb 26 2026

### Cloudflare Turnstile: Widget + server-side verification

- React widget: `@marsidev/react-turnstile` (renders invisible challenge)
- Server-side: No library needed — simple `net/http` POST to `challenges.cloudflare.com/turnstile/v0/siteverify` (20 lines)
- Only apply to unauthenticated public forms (signup, login, forgot-password)

### Email Verification Tokens

Use `crypto/rand.Text()` from Go stdlib (available Go 1.22+, project uses Go 1.23). Store in DB with expiry column. No external library needed.

### Multi-tenant, Traffic Logs, Device Grouping

All solved with SQL migrations + Go service layer changes. No new libraries required.

## What NOT to Add

| Library | Why Not |
|---------|---------|
| NextAuth.js | Conflicts with existing Go JWT auth system |
| Passport.js | Not applicable (Go backend, not Node) |
| PostgreSQL RLS | App-level scoping simpler at this scale, works with pgx pool |
| Zod v4 | Available but don't upgrade mid-milestone — stay on v3 |
| Redis | Not needed yet — session storage via JWT, no caching requirement |

## Integration Points

- Google OAuth → extends existing `AuthService` with new `GoogleLogin` method
- Resend → new `EmailService` injected into `AuthService` for verification flow
- Turnstile → new Gin middleware on signup/login routes
- Swagger → annotations on existing handler methods, new `/swagger/*` route
- next-themes → wraps existing layout, shadcn/ui already CSS-variable ready

## Setup Requirements (Manual Steps)

1. **Resend**: Domain verification in Resend dashboard before first send
2. **Google OAuth**: Register redirect URI in Google Cloud Console (blocks Phase 1 if not done early)
3. **Cloudflare Turnstile**: Create site in Cloudflare dashboard, get site key + secret key

---
*Researched: 2026-02-27 for v2.0 SaaS milestone*
