---
phase: 05-auth-foundation
verified: 2026-02-28T07:00:00Z
status: passed
score: 17/17 must-haves verified
re_verification: false
---

# Phase 5: Auth Foundation Verification Report

**Phase Goal:** Customers can self-register and log in securely using email/password or Google, with email verification and bot protection on all public forms
**Verified:** 2026-02-28
**Status:** PASSED
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | customers table has password_hash, email_verified, google_id, google_email columns | VERIFIED | `server/migrations/011_customer_auth.up.sql` — four ALTER TABLE statements confirmed in file |
| 2 | customer_auth_tokens table exists with token_hash, type, expires_at columns | VERIFIED | `011_customer_auth.up.sql` CREATE TABLE with all required columns + used_at + indexes |
| 3 | Customer domain struct includes all auth fields | VERIFIED | `models.go` lines 95-98: PasswordHash *string, EmailVerified bool, GoogleID *string, GoogleEmail *string |
| 4 | Repository methods exist for GetByEmail, GetByGoogleID, CreateToken, GetValidToken, MarkTokenUsed | VERIFIED | `customer_repo.go`: GetByEmail, GetByGoogleID, CreateWithAuth, UpdateEmailVerified, UpdatePasswordHash, LinkGoogleAccount; `customer_token_repo.go`: Create, GetByHash, MarkUsed, DeleteByCustomerAndType |
| 5 | POST /api/auth/customer/signup creates customer with hashed password and sends verification email | VERIFIED | `customer_auth_service.go` Signup(): bcrypt.GenerateFromPassword + customerRepo.CreateWithAuth + emailService.SendVerification |
| 6 | POST /api/auth/customer/login returns JWT only if email is verified | VERIFIED | `customer_auth_service.go` Login(): checks customer.EmailVerified == false → returns "email_not_verified" error, skipping JWT |
| 7 | GET /api/auth/customer/verify-email?token=X returns token validity status | VERIFIED | `customer_auth_handler.go` VerifyEmailCheck(): queries tokenRepo.GetByHash and returns {"valid": true/false} |
| 8 | POST /api/auth/customer/verify-email consumes token and marks email verified | VERIFIED | `customer_auth_service.go` VerifyEmail(): MarkUsed + UpdateEmailVerified(true) + returns JWT |
| 9 | POST /api/auth/customer/forgot-password sends reset email | VERIFIED | `customer_auth_service.go` ForgotPassword(): Turnstile check + DeleteByCustomerAndType + Create token + emailService.SendPasswordReset |
| 10 | POST /api/auth/customer/reset-password changes password using valid token | VERIFIED | `customer_auth_service.go` ResetPassword(): GetByHash + MarkUsed + bcrypt.GenerateFromPassword + UpdatePasswordHash |
| 11 | GET /api/auth/customer/google redirects to Google OAuth consent screen | VERIFIED | `customer_auth_handler.go` GoogleLogin(): generates state, sets HttpOnly cookie, c.Redirect to authURL |
| 12 | GET /api/auth/customer/google/callback handles OAuth code exchange and returns JWT | VERIFIED | `customer_auth_service.go` GoogleCallback(): three-case logic (existing Google ID / existing email / new user), returns JWT |
| 13 | All three public endpoints (signup, login, forgot-password) verify Turnstile token server-side | VERIFIED | `customer_auth_service.go`: verifyTurnstile() called in Signup(), Login(), ForgotPassword(); dev-mode bypass when SecretKey=="" |
| 14 | Login page has Google sign-in button, Forgot password link, and Sign up link | VERIFIED | `login/page.tsx` lines 197-218: Google button with 4-path SVG logo, /forgot-password link, /signup link |
| 15 | Signup page has email/password fields, Google button, Turnstile widget, and link to login | VERIFIED | `signup/page.tsx`: email/password form, Turnstile conditional render, Google sign-up anchor, login link |
| 16 | Turnstile widget appears on signup, login, and forgot-password forms | VERIFIED | `Turnstile` from `@marsidev/react-turnstile` conditionally rendered (NEXT_PUBLIC_TURNSTILE_SITE_KEY) in all three pages |
| 17 | Google OAuth callback stores token and redirects to dashboard | VERIFIED | `login/page.tsx` useEffect: decodes JWT from ?token=X&google=true via atob(), calls setAuth(), router.push('/devices') |

**Score:** 17/17 truths verified

---

## Required Artifacts

### Plan 01 — DB Schema and Data Access Layer

| Artifact | Status | Details |
|----------|--------|---------|
| `server/migrations/011_customer_auth.up.sql` | VERIFIED | 19 lines; ALTER TABLE + CREATE TABLE + 2 indexes — fully substantive |
| `server/migrations/011_customer_auth.down.sql` | VERIFIED | DROP TABLE + DROP COLUMN rollback present |
| `server/internal/domain/models.go` | VERIFIED | CustomerAuthToken struct, all 7 auth request/response types, auth fields on Customer struct |
| `server/internal/domain/config.go` | VERIFIED | GoogleOAuthConfig, ResendConfig, TurnstileConfig all present; DefaultConfig populated |
| `server/internal/repository/customer_repo.go` | VERIFIED | 6 new auth methods: GetByEmail, GetByGoogleID, CreateWithAuth, UpdateEmailVerified, UpdatePasswordHash, LinkGoogleAccount |
| `server/internal/repository/customer_token_repo.go` | VERIFIED | 58 lines; Create, GetByHash (with validity SQL filter), MarkUsed, DeleteByCustomerAndType |

### Plan 02 — Service Layer and HTTP Handlers

| Artifact | Status | Details |
|----------|--------|---------|
| `server/internal/service/customer_auth_service.go` | VERIFIED | 499 lines; all 9 flows implemented: Signup, Login, VerifyEmailCheck, VerifyEmail, ResendVerification, ForgotPassword, ResetPassword, GoogleAuthURL, GoogleCallback |
| `server/internal/service/email_service.go` | VERIFIED | 145 lines; Resend SDK wrapper, SendVerification, SendPasswordReset, HTML templates, dev-mode fallback |
| `server/internal/api/handler/customer_auth_handler.go` | VERIFIED | 209 lines; all 9 gin.HandlerFunc handlers wired to service, CSRF cookie pattern on Google flows |
| `server/internal/api/handler/router.go` | VERIFIED | customerAuthHandler parameter added; 9 routes registered under /api/auth/customer/ (public) |
| `server/cmd/api/main.go` | VERIFIED | CustomerTokenRepository, EmailService, CustomerAuthService, CustomerAuthHandler instantiated; 8 env vars loaded |

### Plan 03 — Frontend Auth Pages

| Artifact | Status | Details |
|----------|--------|---------|
| `dashboard/src/app/login/page.tsx` | VERIFIED | Extended with Turnstile, Google button, forgot-password link, signup link, Google OAuth callback useEffect, password_updated banner, customer-first login with admin fallback, inline email_not_verified resend UI |
| `dashboard/src/app/signup/page.tsx` | VERIFIED | 135 lines; email/password form, min-8-char validation, Turnstile conditional render, Google sign-up anchor, login link |
| `dashboard/src/app/signup/verify/page.tsx` | VERIFIED | 97 lines; "Check your email" with email from URL param, 60s cooldown resend, success/error messages |
| `dashboard/src/app/verify-email/page.tsx` | VERIFIED | 136 lines; two-step flow: GET check on mount → loading/valid/invalid/no-token states → Verify button POSTs, auto-login via setAuth |
| `dashboard/src/app/forgot-password/page.tsx` | VERIFIED | 108 lines; email + Turnstile, generic success (no enumeration), 429-only real error |
| `dashboard/src/app/reset-password/page.tsx` | VERIFIED | 142 lines; password + confirm, client-side validation, redirects to /login?message=password_updated |
| `dashboard/src/lib/api.ts` | VERIFIED | AuthCustomer interface exported; customerAuth namespace with 7 typed methods: signup, login, verifyEmailCheck, verifyEmail, resendVerification, forgotPassword, resetPassword |

---

## Key Link Verification

### Plan 01 Key Links

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `domain/models.go` | `011_customer_auth.up.sql` | struct fields match DB columns | VERIFIED | PasswordHash/*string matches password_hash VARCHAR; EmailVerified/bool matches email_verified BOOLEAN |
| `customer_token_repo.go` | `domain/models.go` | uses CustomerAuthToken struct | VERIFIED | `domain.CustomerAuthToken` used in Create, GetByHash return type |

### Plan 02 Key Links

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `customer_auth_handler.go` | `customer_auth_service.go` | handler calls service methods | VERIFIED | `h.customerAuthService.Signup/Login/VerifyEmailCheck/VerifyEmail/ResendVerification/ForgotPassword/ResetPassword/GoogleAuthURL/GoogleCallback` — all 9 calls present |
| `customer_auth_service.go` | `customer_repo.go` | service queries customer data | VERIFIED | `s.customerRepo.GetByEmail`, `GetByGoogleID`, `CreateWithAuth`, `UpdateEmailVerified`, `UpdatePasswordHash`, `LinkGoogleAccount`, `GetByID` |
| `customer_auth_service.go` | `email_service.go` | service sends verification and reset emails | VERIFIED | `s.emailService.SendVerification` in Signup/issueVerificationToken; `s.emailService.SendPasswordReset` in ForgotPassword |
| `router.go` | `customer_auth_handler.go` | route registration | VERIFIED | 9 routes registered: customerAuthHandler.Signup/Login/VerifyEmailCheck/VerifyEmail/ResendVerification/ForgotPassword/ResetPassword/GoogleLogin/GoogleCallback |
| `cmd/api/main.go` | `customer_auth_service.go` | dependency injection in main | VERIFIED | `service.NewCustomerAuthService(customerRepo, customerTokenRepo, emailService, cfg.JWT, cfg.Google, cfg.Turnstile)` at line 82 |

### Plan 03 Key Links

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `login/page.tsx` | `/api/auth/customer/login` | api.customerAuth.login() | VERIFIED | Called in handleSubmit; response used to call setAuth and route to /devices |
| `signup/page.tsx` | `/api/auth/customer/signup` | api.customerAuth.signup() | VERIFIED | Called in handleSubmit; on success routes to /signup/verify?email=X |
| `verify-email/page.tsx` | `/api/auth/customer/verify-email` (GET+POST) | verifyEmailCheck + verifyEmail | VERIFIED | GET called on mount to check validity; POST called in handleVerify, response used to setAuth + route to /devices |
| `forgot-password/page.tsx` | `/api/auth/customer/forgot-password` | api.customerAuth.forgotPassword() | VERIFIED | Called in handleSubmit with Turnstile token; success shows generic confirmation |
| `reset-password/page.tsx` | `/api/auth/customer/reset-password` | api.customerAuth.resetPassword() | VERIFIED | Called in handleSubmit; success routes to /login?message=password_updated |
| `signup/verify/page.tsx` | `/api/auth/customer/resend-verification` | api.customerAuth.resendVerification() | VERIFIED | Called in handleResend with 60s cooldown; silently succeeds |
| `login/page.tsx` | Google OAuth | /api/auth/customer/google redirect + ?token callback | VERIFIED | Anchor href points to backend Google endpoint; useEffect decodes JWT from callback URL params |

---

## Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| AUTH-01 | 01, 02, 03 | Customer can sign up with email and password | SATISFIED | Migration adds password_hash column; Signup() hashes password with bcrypt; signup/page.tsx submits to /api/auth/customer/signup |
| AUTH-02 | 01, 02, 03 | Customer receives email verification after signup and must verify before accessing portal | SATISFIED | customer_auth_tokens table; EmailService.SendVerification(); Login() blocks with email_not_verified; verify-email two-step page |
| AUTH-03 | 01, 02, 03 | Customer can sign up and log in with Google account (OAuth) | SATISFIED | google_id/google_email columns; GoogleCallback() three-case logic; GoogleLogin/GoogleCallback handlers; Google buttons on login and signup pages |
| AUTH-04 | 01, 02, 03 | Customer can reset password via email link | SATISFIED | customer_auth_tokens with type='password_reset'; ForgotPassword() + ResetPassword() service methods; forgot-password and reset-password pages |
| AUTH-05 | 02, 03 | Signup and login forms are protected by Cloudflare Turnstile bot protection | SATISFIED | verifyTurnstile() called server-side in Signup, Login, ForgotPassword; @marsidev/react-turnstile widget conditionally rendered on all three frontend forms |

No orphaned requirements — all five AUTH-0x IDs declared in plans match the five AUTH IDs mapped to Phase 5 in REQUIREMENTS.md.

---

## Anti-Patterns Found

None. Scanned all 15 modified/created files. No TODO/FIXME/PLACEHOLDER comments, no empty implementations, no stub handlers.

---

## Human Verification Required

### 1. End-to-End Email Verification Flow

**Test:** Sign up with a real email address, receive verification email, click link, confirm account activates and auto-login works.
**Expected:** Verification email arrives within ~30 seconds; link opens verify-email page; clicking "Verify Email" logs user in and redirects to /devices.
**Why human:** Requires live RESEND_API_KEY and configured sending domain. Cannot verify email delivery programmatically.

### 2. Google OAuth Full Roundtrip

**Test:** Click "Sign in with Google" on login page; complete Google consent; confirm redirect back to dashboard with JWT.
**Expected:** Google OAuth redirect works; callback URL receives ?token=X&google=true; user lands on /devices authenticated.
**Why human:** Requires live GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and authorized redirect URI configured in Google Cloud Console.

### 3. Turnstile Bot Protection Activation

**Test:** Set NEXT_PUBLIC_TURNSTILE_SITE_KEY and TURNSTILE_SECRET_KEY, submit signup form with and without completing Turnstile widget.
**Expected:** Widget appears; submission without completing widget rejected with error; valid completion allows signup.
**Why human:** Widget only renders when env var is set; requires live Cloudflare Turnstile credentials. Dev-mode always passes (SecretKey=="").

### 4. Customer Login Does Not Expose Operator Dashboard

**Test:** Log in as a customer (role="customer"); verify no admin-only pages are accessible.
**Expected:** Customer JWT contains role="customer"; future middleware (Phase 6) enforces scoping. Current phase only generates the JWT with correct role claim.
**Why human:** Role-based access control enforcement is Phase 6 scope; verify JWT payload role field is "customer" by decoding returned token.

---

## Gaps Summary

No gaps. All 17 observable truths verified. All artifacts are substantive (not stubs). All key links are wired end-to-end from DB schema through Go service through HTTP handler through frontend form. All five requirement IDs (AUTH-01 through AUTH-05) are satisfied by concrete implementation evidence.

The four human verification items above require live external service credentials (Resend, Google OAuth, Cloudflare Turnstile) that are expected to be configured at deployment time. The code is fully implemented and correct; the dev-mode fallbacks allow local testing without these credentials.

---

_Verified: 2026-02-28_
_Verifier: Claude (gsd-verifier)_
