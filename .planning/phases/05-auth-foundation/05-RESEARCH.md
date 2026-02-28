# Phase 5: Auth Foundation - Research

**Researched:** 2026-02-28
**Domain:** Go authentication (email/password + Google OAuth2), Resend transactional email, Cloudflare Turnstile bot protection, Next.js 14 auth UI
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Signup flow
- Separate `/signup` page (not a tab on login)
- Fields: email + password only — minimal, no name/company at signup
- "Already have an account? Log in" link on signup page
- Google "Sign up with Google" button on signup page (same as login)
- After successful signup → "Check your email" page telling user to verify before proceeding

#### Login page design
- Keep current layout (PocketProxy logo + email/password fields)
- Add "Sign in with Google" button below the form with a visual divider ("or")
- Google button appears on BOTH signup and login pages
- Add "Don't have an account? Sign up" link on login page
- Add "Forgot password?" link on login page

#### Email verification UX
- After signup, show a "Check your email" page — user cannot access portal until verified
- Verification email contains a link that opens a confirmation page (two-step: GET shows page, POST confirms — prevents corporate email scanners from consuming the token)
- Google OAuth users are auto-verified (skip verification step)
- If unverified user tries to log in, show "Please verify your email first" message with resend option

#### Password reset flow
- "Forgot password?" link on login page
- Clicking it goes to a "Enter your email" page
- Email contains a reset link (24-hour expiry)
- Link opens a "Set new password" page with password + confirm password fields
- After successful reset → redirect to login with "Password updated" banner

#### Cloudflare Turnstile
- Turnstile widget on signup, login, and forgot-password forms
- Server-side verification on all three endpoints (not just frontend widget)

### Claude's Discretion
- Email template design and wording
- Password strength requirements (minimum length, complexity rules)
- Turnstile widget positioning on forms
- Error message wording for all auth flows
- Token expiry times for email verification (research suggested 24 hours)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AUTH-01 | Customer can sign up with email and password | Go signup endpoint + `customers` table auth extension + bcrypt password hashing (already used in codebase) |
| AUTH-02 | Customer receives email verification after signup and must verify before accessing portal | Resend Go SDK for email delivery; crypto/rand token generation; `customer_auth_tokens` migration; two-step GET/POST confirmation flow |
| AUTH-03 | Customer can sign up and log in with Google account (OAuth) | golang.org/x/oauth2 + google.Endpoint; Google userinfo API; auto-verified flag; account deduplication by email |
| AUTH-04 | Customer can reset password via email link | Same token table with `type` discriminator; 24-hour expiry; Resend email; two-step confirm flow |
| AUTH-05 | Signup and login forms protected by Cloudflare Turnstile | @marsidev/react-turnstile on frontend; POST to Cloudflare siteverify in Go middleware; runs on signup, login, and forgot-password |
</phase_requirements>

---

## Summary

The project already has a working operator/admin auth system (Go + JWT + bcrypt + PostgreSQL) that this phase extends with customer-facing self-registration. The key architectural decision already made is that Go handles all auth — no NextAuth.js. This means the Go backend needs new endpoints for customer signup, email verification, Google OAuth callback, password reset, and Turnstile verification. The Next.js frontend needs new pages that mirror the existing login page design.

The existing `customers` table (id, name, email, active) is an operator-managed record with no auth credentials. This phase upgrades it to be a self-registering entity by adding auth columns via migration: `password_hash`, `email_verified`, `google_id`, and a separate `customer_auth_tokens` table for time-limited email verification and password reset tokens.

The three external services — Resend (email), Google OAuth (social login), and Cloudflare Turnstile (bot protection) — all have official Go support and are straightforward to integrate. The critical day-one prerequisite is having all three configured (Resend domain verified, Google Cloud Console project created, Turnstile site/secret key pair created) before writing code.

**Primary recommendation:** Extend the existing `customers` table with auth columns via a numbered migration (011_customer_auth.up.sql), create a new `customer_auth_tokens` table for one-time tokens, build Go handlers following the same pattern as the existing `AuthHandler`, and add frontend pages that reuse the existing login page design system.

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.org/x/oauth2` | latest | Google OAuth2 web server flow | Official Go OAuth2 library; already in go.sum ecosystem; google.Endpoint built in |
| `github.com/resend/resend-go/v3` | v3 | Transactional email (verification, password reset) | Official Resend Go SDK; decided in STATE.md as the email provider |
| `crypto/rand` (stdlib) | Go 1.23 | Secure token generation for email verification and password reset | Standard library; never use math/rand for security tokens |
| `golang.org/x/crypto/bcrypt` | already in go.mod | Password hashing for customer accounts | Already in go.mod; same as operator auth |
| `github.com/golang-jwt/jwt/v5` | already in go.mod | JWT generation for customer sessions | Already in go.mod; extend existing JWTClaims to include `customer_id` + `role: customer` |

### Supporting (Frontend)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `@marsidev/react-turnstile` | latest | Cloudflare Turnstile React widget | Most widely adopted React wrapper; works with Next.js 14; handles token lifecycle |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| @marsidev/react-turnstile | Raw script tag + cf-turnstile div | Script tag works but requires manual token extraction; the package handles Next.js re-renders on route change automatically |
| resend-go/v3 | net/smtp directly | SMTP requires domain/relay setup on VPS; Resend manages deliverability, SPF/DKIM |
| golang.org/x/oauth2 | google.golang.org/api/idtoken | idtoken is for verifying Google-issued ID tokens (service accounts); x/oauth2 is correct for the 3-legged user flow |

**Installation (new dependencies):**
```bash
# Go (in server/)
go get golang.org/x/oauth2
go get github.com/resend/resend-go/v3

# Next.js (in dashboard/)
npm install @marsidev/react-turnstile
```

---

## Architecture Patterns

### Database Migration Strategy

Follow the existing numbered migration pattern. New migration file: `011_customer_auth.up.sql`.

**What changes to the `customers` table:**
- Add `password_hash VARCHAR(255)` (nullable — Google-only users have no password)
- Add `email_verified BOOLEAN NOT NULL DEFAULT FALSE`
- Add `google_id VARCHAR(255) UNIQUE` (nullable — email-only users have no Google ID)
- Add `google_email VARCHAR(255)` (for display; may differ from login email)

**New table: `customer_auth_tokens`** — stores all time-limited tokens (email verification AND password reset, distinguished by `type`):
```sql
CREATE TABLE customer_auth_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 of the random token
    type        VARCHAR(50) NOT NULL,           -- 'email_verify' or 'password_reset'
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,                    -- NULL = unused
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_customer_auth_tokens_hash ON customer_auth_tokens(token_hash);
CREATE INDEX idx_customer_auth_tokens_customer ON customer_auth_tokens(customer_id, type);
```

**Why hash tokens in DB:** If the `customer_auth_tokens` table is compromised, raw tokens cannot be replayed. The unhashed token travels only in the email link. Verify by hashing the incoming token and comparing with `token_hash`.

### Recommended Project Structure (additions only)

```
server/
├── internal/
│   ├── domain/
│   │   └── models.go               # Add CustomerAuthToken, CustomerSignupRequest, etc.
│   ├── repository/
│   │   └── customer_repo.go        # Add auth-related methods (GetByEmail, Update, etc.)
│   │   └── customer_token_repo.go  # New: CRUD for customer_auth_tokens
│   ├── service/
│   │   └── customer_auth_service.go  # New: signup, verify, google OAuth, password reset
│   │   └── email_service.go          # New: Resend client wrapper
│   └── api/handler/
│       └── customer_auth_handler.go  # New: HTTP handlers for all customer auth endpoints
│       └── router.go                 # Add new public routes
├── migrations/
│   └── 011_customer_auth.up.sql
│   └── 011_customer_auth.down.sql

dashboard/
├── src/app/
│   ├── signup/page.tsx              # New: email+password signup form
│   ├── signup/verify/page.tsx       # New: "Check your email" confirmation page
│   ├── verify-email/page.tsx        # New: GET shows confirm button, POST confirms token
│   ├── forgot-password/page.tsx     # New: Enter email form
│   └── reset-password/page.tsx      # New: New password + confirm form
├── src/lib/
│   └── api.ts                       # Add customer auth API calls
```

### Pattern 1: Customer JWT Claims Extension

The existing `JWTClaims` struct carries `UserID`, `Email`, and `Role`. Customer JWTs use the same structure, with `Role = "customer"` and `UserID` pointing to `customers.id`. No separate claims struct needed — the middleware reads `user_role` from context to distinguish admin from customer.

```go
// Source: server/internal/service/auth_service.go (existing pattern)
// Customer token generation — same generateToken pattern, role = "customer"
func (s *CustomerAuthService) generateCustomerToken(c *domain.Customer) (string, error) {
    claims := service.JWTClaims{
        UserID: c.ID,
        Email:  c.Email,
        Role:   "customer",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "mobileproxy",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.Secret))
}
```

### Pattern 2: Secure Token Generation and Storage

```go
// Source: Go stdlib crypto/rand + encoding/hex
import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
)

// Generate: produce a 32-byte random token, return as hex string (64 chars)
func generateToken() (raw string, hashed string, err error) {
    b := make([]byte, 32)
    if _, err = rand.Read(b); err != nil {
        return
    }
    raw = hex.EncodeToString(b)
    sum := sha256.Sum256([]byte(raw))
    hashed = hex.EncodeToString(sum[:])
    return
}

// Store `hashed` in DB, send `raw` in the email link URL
// On verification: hash the incoming raw token, look up by token_hash
```

### Pattern 3: Google OAuth2 Web Server Flow

```go
// Source: pkg.go.dev/golang.org/x/oauth2/google
import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  "https://yourdomain.com/api/auth/customer/google/callback",
    Scopes: []string{
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/userinfo.profile",
    },
    Endpoint: google.Endpoint,
}

// Step 1: Generate state + redirect
func (h *CustomerAuthHandler) GoogleLogin(c *gin.Context) {
    state := generateStateToken() // store in session or signed cookie
    url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// Step 2: Callback — exchange code for token, fetch user info
func (h *CustomerAuthHandler) GoogleCallback(c *gin.Context) {
    code := c.Query("code")
    token, err := googleOAuthConfig.Exchange(c.Request.Context(), code)
    // Use token to call userinfo endpoint
    client := googleOAuthConfig.Client(c.Request.Context(), token)
    resp, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    // Parse: { "id": "...", "email": "...", "verified_email": true, "name": "..." }
    // Upsert customer: if email exists → link google_id; if new → create verified customer
}
```

### Pattern 4: Cloudflare Turnstile Server-Side Verification

```go
// Source: developers.cloudflare.com/turnstile/get-started/server-side-validation/
// No external library needed — simple HTTP POST

func verifyTurnstile(token string, remoteIP string) (bool, error) {
    data := url.Values{}
    data.Set("secret", os.Getenv("TURNSTILE_SECRET_KEY"))
    data.Set("response", token)
    data.Set("remoteip", remoteIP)

    resp, err := http.PostForm(
        "https://challenges.cloudflare.com/turnstile/v0/siteverify",
        data,
    )
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    var result struct {
        Success bool `json:"success"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Success, nil
}
```

Create a `TurnstileMiddleware` or call this as a helper at the top of each handler that needs it. The frontend sends the token in the JSON body as `turnstile_token`.

### Pattern 5: Resend Email Sending

```go
// Source: resend.com/docs/send-with-go
import "github.com/resend/resend-go/v3"

type EmailService struct {
    client *resend.Client
    from   string // e.g. "PocketProxy <noreply@yourdomain.com>"
}

func NewEmailService(apiKey string) *EmailService {
    return &EmailService{
        client: resend.NewClient(apiKey),
        from:   "PocketProxy <noreply@yourdomain.com>",
    }
}

func (s *EmailService) SendVerification(to, token string) error {
    link := fmt.Sprintf("https://yourdomain.com/verify-email?token=%s", token)
    params := &resend.SendEmailRequest{
        From:    s.from,
        To:      []string{to},
        Subject: "Verify your PocketProxy account",
        Html:    fmt.Sprintf(`<p>Click <a href="%s">here</a> to verify your email. Link expires in 24 hours.</p>`, link),
    }
    _, err := s.client.Emails.Send(params)
    return err
}
```

### Pattern 6: Turnstile React Widget (Frontend)

```tsx
// Source: github.com/marsidev/react-turnstile
import { Turnstile } from '@marsidev/react-turnstile'

// Inside the form component:
const [turnstileToken, setTurnstileToken] = useState<string>('')

<Turnstile
  siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY!}
  onSuccess={(token) => setTurnstileToken(token)}
  onError={() => setTurnstileToken('')}
  onExpire={() => setTurnstileToken('')}
/>

// Include in form submission body:
body: { email, password, turnstile_token: turnstileToken }
```

### Pattern 7: Two-Step Email Verification (Scanner-Safe)

The two-step design (GET shows confirmation page, POST actually consumes token) prevents corporate email scanners from accidentally consuming the verification token by prefetching the link URL.

```
GET  /verify-email?token=<raw>  → shows "Click to verify" page
POST /verify-email              → body: {token: "<raw>"} → marks used, logs in user
```

The GET handler looks up the token (no side effect) just to show a valid/expired state. The POST handler is the actual consumption point.

### Anti-Patterns to Avoid

- **Using the `customers` table alone for auth:** The existing `customers` table has no `password_hash` or auth fields — add via migration, do NOT create a separate `customer_users` table that would break foreign keys in `proxy_connections.customer_id`.
- **Storing raw verification tokens in the DB:** If DB is compromised, attacker can immediately use any valid reset link. Always store SHA-256 hash.
- **Verifying Turnstile on the frontend only:** The `cf-turnstile-response` field is trivially forgeable by a client that calls the API directly. Server-side verification via siteverify is mandatory.
- **Using math/rand for tokens:** Never. Use `crypto/rand`. math/rand is predictable and seeded from time.
- **Consuming verification token on GET:** Corporate mail scanners, link preview bots, and Outlook Safe Links all prefetch URLs. A GET-based verification would mark the token used before the human ever clicks.
- **Auto-linking Google account to email-only account silently:** If customer@example.com already exists (email/password), a Google login with the same email should link the accounts, not create a duplicate. Show a "linked" message rather than treating it as a new signup.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Google OAuth2 web flow | Custom HTTP client for Google token exchange | `golang.org/x/oauth2` with `google.Endpoint` | Token refresh, error handling, PKCE support already implemented |
| Transactional email delivery | Direct SMTP from VPS | Resend Go SDK | SPF/DKIM alignment, deliverability reputation, bounce handling — not trivial to self-manage |
| Bot protection | Custom rate-limiting or reCAPTCHA integration | Cloudflare Turnstile | Free, no user friction on most traffic, server-side verify is a single HTTP call |
| Password hashing | Custom hash scheme | `golang.org/x/crypto/bcrypt` (already in go.mod) | bcrypt is already used for operator passwords — use same pattern |
| Secure random tokens | UUID or time-based tokens | `crypto/rand` (stdlib) | UUIDs are not secret enough; time-based tokens are predictable |

**Key insight:** The entire auth stack (hashing, JWT, OAuth2) is solved problems in Go. The only custom code needed is the business logic wiring: which endpoints exist, what the token table looks like, and what the email HTML says.

---

## Common Pitfalls

### Pitfall 1: `customers` table has no auth columns yet
**What goes wrong:** Attempting to store `password_hash` or `email_verified` fails because the columns don't exist in the current schema. The `Domain.Customer` struct has no auth fields.
**Why it happens:** The current `customers` table is an operator-managed record (name, email, active). Auth columns were never added.
**How to avoid:** Write migration `011_customer_auth.up.sql` as the very first task. Add `password_hash`, `email_verified`, `google_id` to `customers` table. Extend `domain.Customer` struct to match.
**Warning signs:** Compile errors referencing undefined struct fields on `Customer`.

### Pitfall 2: Google OAuth redirect URI mismatch
**What goes wrong:** Google returns `redirect_uri_mismatch` error during OAuth callback.
**Why it happens:** The redirect URI in the Google Cloud Console must exactly match the `RedirectURL` in `oauth2.Config`, including scheme, host, and path.
**How to avoid:** Register the exact URL `https://yourdomain.com/api/auth/customer/google/callback` in Google Cloud Console before testing. For local dev, also register `http://localhost:8080/api/auth/customer/google/callback` as a separate authorized URI.
**Warning signs:** OAuth callback returns 400 with "Error 400: redirect_uri_mismatch".

### Pitfall 3: Duplicate customer on Google signup when email already exists
**What goes wrong:** A customer who signed up with email/password tries Google login — a second customer record is created with a conflicting email.
**Why it happens:** The callback handler only checks `google_id`, not email.
**How to avoid:** In the Google callback: first check `GetByGoogleID`, then check `GetByEmail`. If email match found, link `google_id` to existing account and mark `email_verified = true`. Never create two records with the same email.
**Warning signs:** Unique constraint violation on `customers.email` during Google callback.

### Pitfall 4: Resend domain not verified before testing
**What goes wrong:** Emails sent from an unverified domain are rejected or land in spam.
**Why it happens:** Resend requires DNS records (SPF, DKIM, DMARC) to be configured and verified before emails can be sent from a custom domain.
**How to avoid:** Verify the domain in the Resend dashboard before writing any email code. Use the `resend.dev` test address only for initial SDK validation.
**Warning signs:** Resend API returns error about unverified sender domain.

### Pitfall 5: Turnstile token is 5-minute TTL — form abandonment invalidates it
**What goes wrong:** User loads the signup form, gets distracted for 6 minutes, then submits — Turnstile token expired.
**Why it happens:** Turnstile tokens are single-use and expire after 300 seconds.
**How to avoid:** Configure `@marsidev/react-turnstile` `onExpire` callback to reset the token state. Display a "Please complete the security check again" message if the user submits with an empty/expired token. The Turnstile widget auto-refreshes itself when `onExpire` fires.
**Warning signs:** Server returns 400 "captcha verification failed" on long-idle form submissions.

### Pitfall 6: Login for customer vs operator — same endpoint or different?
**What goes wrong:** Using the same `/api/auth/login` endpoint means a customer with the same email as an operator could collide. Or the JWT `role` field is not checked, allowing customers to access admin routes.
**Why it happens:** The existing `users` table (operators) and `customers` table are separate, but the login endpoint only checks `users`.
**How to avoid:** Add a separate `/api/auth/customer/login` endpoint that checks the `customers` table, not `users`. The middleware already reads `user_role` from JWT — ensure admin-only routes check for `role = "admin"` or `role = "operator"`, not just "authenticated".
**Warning signs:** Customers able to reach `/api/devices` endpoints; operators getting 401 on login.

### Pitfall 7: Google state parameter CSRF protection
**What goes wrong:** Without a validated `state` parameter, an attacker can craft a malicious callback URL and force-link a victim's account to an attacker's Google account.
**Why it happens:** OAuth2 CSRF is easy to overlook in simple implementations.
**How to avoid:** Generate a random `state` value, store it in a short-lived signed cookie or server-side session before redirecting to Google. Verify the `state` parameter in the callback matches before processing. `golang.org/x/oauth2` does not manage state automatically.
**Warning signs:** Missing state verification in the callback handler.

---

## Code Examples

### Customer Signup Handler (Go)

```go
// Source: pattern follows server/internal/api/handler/auth_handler.go
func (h *CustomerAuthHandler) Signup(c *gin.Context) {
    var req struct {
        Email          string `json:"email" binding:"required,email"`
        Password       string `json:"password" binding:"required,min=8"`
        TurnstileToken string `json:"turnstile_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 1. Verify Turnstile
    ok, err := h.turnstile.Verify(req.TurnstileToken, c.ClientIP())
    if err != nil || !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "captcha verification failed"})
        return
    }

    // 2. Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }

    // 3. Create customer (email_verified = false)
    customer := &domain.Customer{
        ID:            uuid.New(),
        Email:         strings.ToLower(req.Email),
        PasswordHash:  string(hash),
        EmailVerified: false,
        Active:        true,
    }
    if err := h.customerRepo.Create(c.Request.Context(), customer); err != nil {
        // Handle unique constraint on email
        c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
        return
    }

    // 4. Generate verification token, send email
    raw, hashed, _ := generateToken()
    h.tokenRepo.CreateToken(c.Request.Context(), customer.ID, hashed, "email_verify", 24*time.Hour)
    h.emailService.SendVerification(customer.Email, raw)

    c.JSON(http.StatusCreated, gin.H{"message": "check your email to verify your account"})
}
```

### Migration 011 (SQL)

```sql
-- 011_customer_auth.up.sql
ALTER TABLE customers
    ADD COLUMN password_hash   VARCHAR(255),          -- NULL for Google-only accounts
    ADD COLUMN email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN google_id       VARCHAR(255) UNIQUE,   -- NULL for email-only accounts
    ADD COLUMN google_email    VARCHAR(255);           -- Raw email from Google profile

CREATE TABLE customer_auth_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    type        VARCHAR(50) NOT NULL CHECK (type IN ('email_verify', 'password_reset')),
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customer_auth_tokens_hash ON customer_auth_tokens(token_hash);
CREATE INDEX idx_customer_auth_tokens_customer ON customer_auth_tokens(customer_id, type);
```

### Turnstile React Component (Next.js)

```tsx
// dashboard/src/app/signup/page.tsx — excerpt
// Source: github.com/marsidev/react-turnstile
'use client'
import { Turnstile } from '@marsidev/react-turnstile'
import { useState, FormEvent } from 'react'

export default function SignupPage() {
  const [turnstileToken, setTurnstileToken] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!turnstileToken) {
      setError('Please complete the security check')
      return
    }
    await api.customerAuth.signup(email, password, turnstileToken)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* ... email and password fields ... */}
      <Turnstile
        siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY!}
        onSuccess={setTurnstileToken}
        onError={() => setTurnstileToken('')}
        onExpire={() => setTurnstileToken('')}
        className="mt-2"
      />
      <button type="submit" disabled={!turnstileToken || loading}>
        Create Account
      </button>
    </form>
  )
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| reCAPTCHA v2/v3 | Cloudflare Turnstile | 2022-present | Turnstile is free, no user friction for legitimate users, no Google dependency |
| Nodemailer / SMTP from app server | Managed transactional email (Resend, Postmark) | ~2020-present | Deliverability, bounce handling, and SPF/DKIM complexity managed by provider |
| Raw session cookies for OAuth state | Signed JWT or HMAC-signed cookie for state | Standard practice | Prevents CSRF; stateless server is easier to scale |
| Verifying email token via GET | Two-step GET (show page) + POST (consume) | Best practice since ~2018 | Email scanners consume GET links; two-step prevents token theft |

**Deprecated/outdated:**
- `crypto/rand` seed from time.Now(): Never acceptable. Go's `crypto/rand` is CSPRNG — use directly.
- Storing raw tokens in DB: Use SHA-256 hash. If DB leaks, attacker cannot use the tokens.
- `golang.org/x/oauth2` with `oauth2.NoContext`: Deprecated. Use `context.Background()` or request context.

---

## Open Questions

1. **Email sender domain**
   - What we know: Resend requires a verified sending domain; the `from` address must use that domain
   - What's unclear: Which domain is configured for PocketProxy email? The VPS is at 178.156.240.184 — what domain is pointed at it?
   - Recommendation: Confirm domain before starting. Use `noreply@[your-domain]` as sender. If domain is not set up, Resend allows `onboarding@resend.dev` for testing only.

2. **Google OAuth redirect domain**
   - What we know: Redirect URI must be registered in Google Cloud Console exactly
   - What's unclear: Is there a custom domain for the dashboard? Or is it accessed via IP?
   - Recommendation: Set up a proper domain (even a subdomain) before implementing OAuth. Google does not allow IP addresses as redirect URIs for production.

3. **Customer login endpoint — separate or merged with admin login?**
   - What we know: The existing `/api/auth/login` checks the `users` table (operators). Customers are in `customers` table.
   - What's unclear: Should there be one login endpoint that checks both tables, or two separate endpoints?
   - Recommendation: Two separate endpoints (`/api/auth/login` for operators, `/api/auth/customer/login` for customers) is cleaner and avoids role collision. The frontend login page can try the appropriate endpoint based on context, or just use the customer endpoint by default (the admin dashboard can keep its own login flow).

4. **Password strength requirements (Claude's discretion)**
   - Recommendation: Minimum 8 characters, no complexity rules (research shows complexity rules increase "Password1!" patterns; length matters more). Use `binding:"required,min=8"` on the signup endpoint.

5. **Email verification token expiry (Claude's discretion)**
   - Recommendation: 24 hours for email verification tokens (matches password reset expiry, reduces support requests from slow responders). Store `expires_at = NOW() + INTERVAL '24 hours'` at token creation.

---

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/golang.org/x/oauth2/google` — Google OAuth2 web server flow, Endpoint constant, Config struct
- `developers.cloudflare.com/turnstile/get-started/server-side-validation/` — Siteverify endpoint, request format, response format, token TTL (5 min), single-use constraint
- `resend.com/docs/send-with-go` — Go SDK v3 installation, SendEmailRequest struct, client usage
- Go stdlib `crypto/rand`, `crypto/sha256` — Secure token generation pattern
- Project codebase (existing auth_service.go, auth_handler.go, router.go, models.go) — Exact patterns to follow

### Secondary (MEDIUM confidence)
- `github.com/marsidev/react-turnstile` — Most widely adopted React Turnstile wrapper; npm package exists and is actively maintained; verified via npm registry search
- `developers.cloudflare.com/turnstile/get-started/client-side-rendering/` — `cf-turnstile-response` field name, `data-sitekey` attribute, implicit rendering behavior
- WebSearch results for Google userinfo endpoint (`https://www.googleapis.com/oauth2/v2/userinfo`) — confirmed by multiple Go OAuth2 example implementations

### Tertiary (LOW confidence)
- State management for OAuth CSRF: Recommended signed cookie approach based on general Go web patterns; specific implementation (cookie vs in-memory) should be validated against project's session strategy

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries verified via official docs and pkg.go.dev; existing codebase already uses bcrypt + jwt which are reused
- Architecture: HIGH — migration pattern matches existing 001-010 migrations; handler/service/repo pattern matches existing codebase exactly
- Pitfalls: HIGH — Google redirect URI mismatch, scanner-consuming GET links, and duplicate account on OAuth are well-documented failure modes verified across multiple sources

**Research date:** 2026-02-28
**Valid until:** 2026-03-30 (Resend and Turnstile APIs are stable; golang.org/x/oauth2 is stable)
