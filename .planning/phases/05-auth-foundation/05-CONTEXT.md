# Phase 5: Auth Foundation - Context

**Gathered:** 2026-02-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Enable customer self-registration and login using email/password or Google OAuth, with email verification before portal access, password reset via email, and Cloudflare Turnstile bot protection on all public forms. This phase builds the auth infrastructure — the customer portal UI is Phase 7.

</domain>

<decisions>
## Implementation Decisions

### Signup flow
- Separate `/signup` page (not a tab on login)
- Fields: email + password only — minimal, no name/company at signup
- "Already have an account? Log in" link on signup page
- Google "Sign up with Google" button on signup page (same as login)
- After successful signup → "Check your email" page telling user to verify before proceeding

### Login page design
- Keep current layout (PocketProxy logo + email/password fields)
- Add "Sign in with Google" button below the form with a visual divider ("or")
- Google button appears on BOTH signup and login pages
- Add "Don't have an account? Sign up" link on login page
- Add "Forgot password?" link on login page

### Email verification UX
- After signup, show a "Check your email" page — user cannot access portal until verified
- Verification email contains a link that opens a confirmation page (two-step: GET shows page, POST confirms — prevents corporate email scanners from consuming the token)
- Google OAuth users are auto-verified (skip verification step)
- If unverified user tries to log in, show "Please verify your email first" message with resend option

### Password reset flow
- "Forgot password?" link on login page
- Clicking it goes to a "Enter your email" page
- Email contains a reset link (24-hour expiry)
- Link opens a "Set new password" page with password + confirm password fields
- After successful reset → redirect to login with "Password updated" banner

### Cloudflare Turnstile
- Turnstile widget on signup, login, and forgot-password forms
- Server-side verification on all three endpoints (not just frontend widget)

### Claude's Discretion
- Email template design and wording
- Password strength requirements (minimum length, complexity rules)
- Turnstile widget positioning on forms
- Error message wording for all auth flows
- Token expiry times for email verification (research suggested 24 hours)

</decisions>

<specifics>
## Specific Ideas

- Current login page style should be preserved — just extended with Google button and links
- "Check your email" page should be clean and simple, matching the login page design
- Google button should use standard Google branding (white button with Google logo)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-auth-foundation*
*Context gathered: 2026-02-28*
