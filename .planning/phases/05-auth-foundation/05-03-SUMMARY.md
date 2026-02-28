---
phase: 05-auth-foundation
plan: 03
subsystem: auth-frontend
tags: [nextjs, typescript, react, turnstile, google-oauth, email-verification, password-reset]

# Dependency graph
requires:
  - phase: 05-auth-foundation
    plan: 02
    provides: Backend auth endpoints (/api/auth/customer/*) for all 9 flows

provides:
  - Login page extended with Google sign-in button, Turnstile widget, forgot-password and signup links, Google OAuth callback handler, password_updated success banner, customer login with email_not_verified resend UI
  - Signup page with email/password form, Turnstile, Google button, validation
  - signup/verify page: "Check your email" with 60s cooldown resend button
  - verify-email page: two-step verification (GET checks token validity, POST confirms with auto-login)
  - forgot-password page: email + Turnstile form, generic success (no enumeration)
  - reset-password page: password + confirm fields, client-side match validation, redirects to login with success banner
  - customerAuth namespace in api.ts with 7 typed methods (signup, login, verifyEmailCheck, verifyEmail, resendVerification, forgotPassword, resetPassword)
  - AuthCustomer interface exported from api.ts

affects:
  - Future customer portal pages (Phase 7) that build on these auth flows
  - Any page that imports customerAuth from api.ts

# Tech tracking
tech-stack:
  added:
    - "@marsidev/react-turnstile ^0.x (Cloudflare Turnstile React widget)"
  patterns:
    - Turnstile conditional render: only shows when NEXT_PUBLIC_TURNSTILE_SITE_KEY env var is set (graceful dev mode)
    - Google OAuth callback via query params: /login?token=X&google=true, JWT decoded with atob()
    - Two-step email verification: GET to check token validity before showing Verify button (prevents email scanner consumption)
    - Customer login with fallback to admin login: tries customerAuth.login first, falls through to api.auth.login on non-verification failure
    - 60s resend cooldown via useEffect countdown timer
    - Email enumeration prevention: forgot-password always shows generic success regardless of email existence
    - password_updated query param pattern: reset-password redirects to /login?message=password_updated, login page shows green banner and clears URL param

key-files:
  created:
    - dashboard/src/app/signup/page.tsx
    - dashboard/src/app/signup/verify/page.tsx
    - dashboard/src/app/verify-email/page.tsx
    - dashboard/src/app/forgot-password/page.tsx
    - dashboard/src/app/reset-password/page.tsx
  modified:
    - dashboard/src/app/login/page.tsx
    - dashboard/src/lib/api.ts
    - dashboard/package.json

key-decisions:
  - "Turnstile widget conditionally rendered via process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY check — no widget in dev without env var, forms still work"
  - "Customer login tried first, admin login as fallback — preserves operator login without requiring separate login URL"
  - "email_not_verified sentinel string from backend drives resend verification UI inline in login error box"
  - "verify-email uses two-step pattern: GET check on mount shows state, user clicks Verify button to POST — prevents link scanners from consuming token"
  - "forgot-password shows generic success even on errors (except 429) to prevent email enumeration"

patterns-established:
  - "Conditional Turnstile render: check process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY before mounting widget"
  - "Google OAuth redirect: /login?token=X&google=true with JWT decoded via atob() on client"
  - "Two-step token verification: GET check validity first, POST to consume — prevents scanners"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05]

# Metrics
duration: 20min
completed: 2026-02-28
---

# Phase 5 Plan 03: Auth Foundation — Customer Frontend Auth Pages Summary

**Six Next.js auth pages (login extended, signup, verify confirmation, email verification two-step, forgot-password, reset-password) with Turnstile bot protection and Google OAuth integration**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-02-28T06:06:46Z
- **Completed:** 2026-02-28T06:26:00Z
- **Tasks:** 3 complete (Tasks 1+2 auto, Task 3 checkpoint:human-verify — approved by user)
- **Files modified:** 8 (5 created, 3 modified)

## Accomplishments

- Installed `@marsidev/react-turnstile` package
- Extended `dashboard/src/lib/api.ts` with `AuthCustomer` interface and `customerAuth` namespace (7 methods)
- Updated login page: Turnstile widget (conditional on env var), Google sign-in button with 4-color G logo (per Google branding), "Forgot password?" link, "Sign up" link, Google OAuth callback handler (decodes JWT from ?token=X&google=true), password_updated green banner, customer login with fallback to admin login, inline resend verification UI for email_not_verified errors
- Created signup page: email/password with 8-char validation, Turnstile, Google button, links to login
- Created signup/verify page: "Check your email" with email from URL param, 60-second resend cooldown
- Created verify-email page: two-step flow — GET checks token on mount (loading state), shows Verify button for valid tokens, invalid/expired/no-token states all handled, POST confirms with auto-login via setAuth + redirect to /devices
- Created forgot-password page: email + Turnstile form, generic success message (prevents email enumeration even on unexpected errors)
- Created reset-password page: password + confirm fields, client-side length and match validation, redirects to /login?message=password_updated on success, inline "request new link" prompt on expired token errors
- `npm run build` passes: all 6 auth routes compile successfully (15 pages total)
- User visually confirmed all auth pages render correctly on localhost:3001 (Task 3 checkpoint approved)

## Task Commits

Each task was committed atomically:

1. **Task 1: Install Turnstile, extend API client, build signup + login pages** - `8d0b5f6` (feat)
2. **Task 2: Verify-email, forgot-password, and reset-password pages** - `65c0026` (feat)
3. **Task 3: Verify complete auth flow end-to-end** - checkpoint:human-verify approved

**Plan metadata:** `ab42431` (docs: complete frontend auth pages plan)

## Files Created/Modified

- `dashboard/src/app/login/page.tsx` - Extended with: Turnstile widget, Google sign-in button (proper 4-path SVG logo), Forgot password link, Sign up link, Google OAuth callback useEffect (decodes JWT, calls setAuth, redirects), password_updated banner useEffect, customer login with admin fallback, email_not_verified inline resend UI
- `dashboard/src/lib/api.ts` - Added AuthCustomer interface; added customerAuth namespace with signup, login, verifyEmailCheck, verifyEmail, resendVerification, forgotPassword, resetPassword methods
- `dashboard/src/app/signup/page.tsx` - Create account form with email/password (min 8 chars), Turnstile, Google button, links; redirects to /signup/verify?email=X on success; 409 conflict handled
- `dashboard/src/app/signup/verify/page.tsx` - Check your email page; reads email from URL param; resend with 60s cooldown countdown; success/error messages
- `dashboard/src/app/verify-email/page.tsx` - Two-step verification; GET check on mount sets loading/valid/invalid/no-token state; valid state shows Verify Email button; POST confirm calls setAuth + redirects to /devices
- `dashboard/src/app/forgot-password/page.tsx` - Email form with Turnstile; submitted state shows generic confirmation (no email enumeration); 429 shown as only real error
- `dashboard/src/app/reset-password/page.tsx` - New password + confirm fields; validates length and match client-side; on success redirects to /login?message=password_updated; expired/used token error shows "request new link" link
- `dashboard/package.json` - Added @marsidev/react-turnstile dependency

## Decisions Made

- **Conditional Turnstile render**: Widget only renders when `NEXT_PUBLIC_TURNSTILE_SITE_KEY` env var is set. When not set (dev), widget is hidden and forms pass empty string as turnstile_token — backend accepts this in dev mode (no secret key = auto-pass).
- **Single login form with customer-first fallback**: Tries `api.customerAuth.login` first, falls back to `api.auth.login` for operator/admin. No separate login URL needed — operators never have email_not_verified errors so the fallback is transparent.
- **email_not_verified inline resend**: Rather than redirecting to a separate page, the resend button appears inline in the login error box. Cleaner UX for the expected recovery path.
- **Two-step email verification**: GET on mount checks token validity (shows loading spinner), then shows a button the user must click to POST-confirm. Prevents corporate email link scanners from consuming tokens.
- **Generic forgot-password success**: Even on unexpected errors (not 429), the page transitions to the success state. Email enumeration prevention outweighs showing fetch errors.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

External services require manual configuration before the full auth flow can be tested in production:

- **Cloudflare Turnstile**: Create widget in Cloudflare Dashboard, add `NEXT_PUBLIC_TURNSTILE_SITE_KEY` env var
- **Google OAuth**: Create OAuth credentials in Google Cloud Console, add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`
- **Resend**: Verify sending domain, add `RESEND_API_KEY`

All three services have dev-mode fallbacks: pages work without env vars set (Turnstile hidden, emails logged to console, Google button links to backend which handles missing credentials gracefully).

## Next Phase Readiness

Phase 5 (Auth Foundation) is complete. All three plans delivered:
- 05-01: DB schema (customer_accounts, email_verification_tokens, password_reset_tokens tables)
- 05-02: Go backend auth service layer and API handlers for all 9 auth endpoints
- 05-03: Next.js frontend auth pages (6 pages), Turnstile integration, Google OAuth

Ready for Phase 6: Tenant Isolation (customer scoping on proxy resources, devices, traffic). The auth foundation provides JWT tokens with customer_id claims that Phase 6 will use for row-level authorization.

---
*Phase: 05-auth-foundation*
*Completed: 2026-02-28*

## Self-Check: PASSED

Created files verified:
- dashboard/src/app/signup/page.tsx: FOUND (git: 8d0b5f6)
- dashboard/src/app/signup/verify/page.tsx: FOUND (git: 8d0b5f6)
- dashboard/src/app/verify-email/page.tsx: FOUND (git: 65c0026)
- dashboard/src/app/forgot-password/page.tsx: FOUND (git: 65c0026)
- dashboard/src/app/reset-password/page.tsx: FOUND (git: 65c0026)
- dashboard/src/app/login/page.tsx: Modified (git: 8d0b5f6)
- dashboard/src/lib/api.ts: Modified (git: 8d0b5f6)

npm run build: PASSED (15 routes compiled, 0 errors)
Task commits: 8d0b5f6 (Task 1), 65c0026 (Task 2) — verified in git log
Checkpoint Task 3: approved by user 2026-02-28
