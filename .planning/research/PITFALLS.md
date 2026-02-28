# Pitfalls Research: PocketProxy v2.0 SaaS Transformation

**Mode:** Ecosystem — Pitfalls
**Confidence:** HIGH for tenant isolation and auth migration (direct code inspection); MEDIUM for OAuth/email/Turnstile patterns

## Critical Pitfalls

### P1: Tenant Data Leak (CRITICAL)

**Risk:** Every existing handler returns ALL records with no customer scoping. When the customer portal is added, reusing any existing handler without customer-scoped filtering leaks one customer's proxy credentials to another.

**Current code:**
- `connHandler.List()` — returns ALL connections
- `deviceHandler.List()` — returns ALL devices
- `AuthMiddleware` sets `user_id` and `user_role` but has no `customer_id` concept

**Prevention:**
- Create a SEPARATE `/api/portal/*` route group for customer-facing endpoints
- Never expose existing admin endpoints to customer role
- Every portal repo method MUST include `WHERE customer_id = $X`
- Code review checklist: search for any query missing customer_id filter

**Phase:** Multi-tenant isolation (must be 100% before any customer gets portal access)

### P2: Google OAuth Backend Bypass (CRITICAL)

**Risk:** If the backend trusts the frontend's assertion that a Google login occurred without independently verifying the Google ID token, attackers can forge authentication.

**Prevention:**
- Backend must call Google's token verification independently (`go-oidc` library)
- NEVER accept a Google user profile from the frontend — only accept the raw ID token
- Backend verifies token signature, audience (must match your client ID), and expiry
- Create a SEPARATE `/api/auth/google` endpoint — do not mix with email/password login

**Phase:** Auth & Signup

### P3: Email Verification Token Consumed by Corporate Scanners (HIGH)

**Risk:** Microsoft Defender, Proofpoint, and Barracuda pre-fetch GET links in emails before the user clicks. If verification uses a simple GET endpoint, the token is consumed by the scanner and the user sees "token expired/invalid."

**Prevention:**
- Use two-step verification: GET shows a "Confirm your email" page with a button, POST commits the verification
- Never verify on GET alone
- Set token expiry to 24 hours (not 15 minutes)

**Phase:** Auth & Signup

### P4: Unverified Accounts Using Proxies (HIGH)

**Risk:** If a user can sign up and immediately use proxy connections before verifying their email, abuse actors will create throwaway accounts.

**Prevention:**
- Check `email_verified = true` before allowing any proxy-related API calls
- Google OAuth users are auto-verified (Google already verified the email)
- Show a clear "verify your email" gate in the portal UI

**Phase:** Auth & Signup

## High-Risk Pitfalls

### P5: Turnstile Widget Without Server-Side Verification

**Risk:** The Turnstile widget alone does nothing to stop bot API requests. Without server-side verification, bots can POST directly to signup/login endpoints.

**Prevention:**
- Go signup/login handlers MUST call `challenges.cloudflare.com/turnstile/v0/siteverify`
- Reject requests where `success: true` is not in the response
- Pass `cf-turnstile-response` as a request header from frontend

**Phase:** Auth & Signup

### P6: IP Whitelist CIDR Bypass

**Risk:** String-based IP comparison doesn't handle CIDR notation and is trivially bypassed.

**Prevention:**
- Use `net.ParseCIDR()` + `Contains()` for comparison
- For single IPs, append `/32` before parsing
- Only trust `X-Forwarded-For` from known proxy container IPs
- Configure Gin's `SetTrustedProxies()` — default trusts all forwarded headers

**Phase:** IP Whitelist

### P7: Customer Creating Connections on Any Device

**Risk:** If the portal allows customers to create connections by passing a device_id, a customer could pass another customer's device_id.

**Prevention:**
- Portal connection creation must verify: `device.customer_id == jwt.customer_id`
- Better: operator pre-assigns devices to customers, customers can only manage connections on their assigned devices
- Never let customers specify arbitrary device IDs

**Phase:** Customer Portal

### P8: Email Pre-Registration Hijacking

**Risk:** Attacker registers with victim's email before victim signs up. If email verification is weak, attacker owns the account.

**Prevention:**
- Don't create the account until email is verified (or create in "unverified" state with no access)
- If Google OAuth links to existing email, require email verification first
- Prevent duplicate signups for same email

**Phase:** Auth & Signup

## Medium-Risk Pitfalls

### P9: SPF/DKIM Email Delivery

**Risk:** Verification emails land in spam if SPF/DKIM/DMARC are not configured.

**Prevention:**
- Use Resend — handles DKIM automatically after domain verification
- Verify sending domain in Resend dashboard before development
- Test with Gmail and Outlook (both are strict)

**Phase:** Auth & Signup

### P10: Tailwind CSS Purge Breaking Runtime Theme Variables

**Risk:** Tailwind's JIT compiler purges CSS classes not found in source files.

**Prevention:**
- Override CSS variables on `:root`, not class names — shadcn/ui already uses CSS variables
- Test theme switching with production build (not just dev server)

**Phase:** White-Label Theming

### P11: CORS Wildcard Breaking Custom Domains

**Risk:** If white-label supports custom domains, CORS must allow those origins dynamically.

**Prevention:**
- Dynamic CORS: read allowed origins from tenant_themes table
- Or: use same-origin architecture (API and dashboard on same domain)

**Phase:** White-Label Theming

### P12: Unbounded Traffic Log Storage

**Risk:** Per-port traffic logs without retention policy grow unbounded.

**Prevention:**
- Set retention policy (30 days) from day one
- Cron job: `DELETE FROM bandwidth_logs WHERE created_at < NOW() - INTERVAL '30 days'`
- Add index on `created_at` for efficient cleanup

**Phase:** Traffic Logs

## Pitfall-to-Phase Summary

| Phase | Pitfalls to Address |
|-------|---------------------|
| Auth & Signup | P2 (OAuth bypass), P3 (email scanner), P4 (unverified accounts), P5 (Turnstile server-side), P8 (pre-registration hijack), P9 (SPF/DKIM) |
| Multi-Tenant | P1 (tenant data leak — CRITICAL) |
| Customer Portal | P7 (cross-customer device access) |
| IP Whitelist | P6 (CIDR bypass, header spoofing) |
| White-Label | P10 (CSS purge), P11 (CORS custom domains) |
| Traffic Logs | P12 (unbounded storage) |

---
*Researched: 2026-02-27 for v2.0 SaaS milestone*
