package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// CustomerAuthService implements all customer authentication flows:
// signup, login, email verification, password reset, and Google OAuth.
type CustomerAuthService struct {
	customerRepo *repository.CustomerRepository
	tokenRepo    *repository.CustomerTokenRepository
	emailService *EmailService
	jwtConfig    domain.JWTConfig
	googleConfig domain.GoogleOAuthConfig
	turnstileCfg domain.TurnstileConfig
}

// NewCustomerAuthService constructs a CustomerAuthService with all required dependencies.
func NewCustomerAuthService(
	customerRepo *repository.CustomerRepository,
	tokenRepo *repository.CustomerTokenRepository,
	emailService *EmailService,
	jwtConfig domain.JWTConfig,
	googleConfig domain.GoogleOAuthConfig,
	turnstileCfg domain.TurnstileConfig,
) *CustomerAuthService {
	return &CustomerAuthService{
		customerRepo: customerRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		jwtConfig:    jwtConfig,
		googleConfig: googleConfig,
		turnstileCfg: turnstileCfg,
	}
}

// ─── Private helpers ───────────────────────────────────────────────────────────

// generateToken creates a cryptographically random 32-byte token.
// Returns the hex-encoded raw token (sent to the user) and its SHA-256 hash
// (stored in the database). Never uses math/rand.
func generateToken() (raw string, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random token: %w", err)
	}
	raw = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return raw, hashed, nil
}

// hashToken returns the SHA-256 hex digest of the raw token string.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// verifyTurnstile calls the Cloudflare Turnstile siteverify API.
// Returns true if the token is valid or if no secret key is configured
// (development mode).
func (s *CustomerAuthService) verifyTurnstile(token, remoteIP string) (bool, error) {
	if s.turnstileCfg.SecretKey == "" {
		return true, nil
	}

	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify",
		url.Values{
			"secret":   {s.turnstileCfg.SecretKey},
			"response": {token},
			"remoteip": {remoteIP},
		})
	if err != nil {
		return false, fmt.Errorf("turnstile request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read turnstile response: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("parse turnstile response: %w", err)
	}
	return result.Success, nil
}

// generateCustomerJWT mints a signed JWT for a customer with role "customer".
// Expires in 24 hours.
func (s *CustomerAuthService) generateCustomerJWT(customer *domain.Customer) (string, error) {
	claims := JWTClaims{
		UserID: customer.ID,
		Email:  customer.Email,
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mobileproxy",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

// googleOAuthConfig builds the oauth2.Config for Google.
func (s *CustomerAuthService) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.googleConfig.ClientID,
		ClientSecret: s.googleConfig.ClientSecret,
		RedirectURL:  s.googleConfig.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// ─── Public methods ────────────────────────────────────────────────────────────

// Signup creates a new customer account and sends a verification email.
// Returns an error if the Turnstile check fails or the email is already taken.
func (s *CustomerAuthService) Signup(ctx context.Context, req *domain.CustomerSignupRequest, remoteIP string) error {
	ok, err := s.verifyTurnstile(req.TurnstileToken, remoteIP)
	if err != nil {
		return fmt.Errorf("turnstile check: %w", err)
	}
	if !ok {
		return fmt.Errorf("turnstile verification failed")
	}

	email := strings.ToLower(req.Email)

	// Conflict check
	if _, err := s.customerRepo.GetByEmail(ctx, email); err == nil {
		return fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	hashStr := string(hash)

	// Derive display name from email prefix
	namePart := email
	if idx := strings.Index(email, "@"); idx > 0 {
		namePart = email[:idx]
	}

	customer := &domain.Customer{
		ID:            uuid.New(),
		Name:          namePart,
		Email:         email,
		Active:        true,
		PasswordHash:  &hashStr,
		EmailVerified: false,
	}
	if err := s.customerRepo.CreateWithAuth(ctx, customer); err != nil {
		return fmt.Errorf("create customer: %w", err)
	}

	return s.issueVerificationToken(ctx, customer.ID, email)
}

// issueVerificationToken deletes old email_verify tokens and issues a new one,
// then sends the verification email.
func (s *CustomerAuthService) issueVerificationToken(ctx context.Context, customerID uuid.UUID, email string) error {
	if err := s.tokenRepo.DeleteByCustomerAndType(ctx, customerID, "email_verify"); err != nil {
		return fmt.Errorf("delete old verify tokens: %w", err)
	}

	raw, hashed, err := generateToken()
	if err != nil {
		return err
	}

	token := &domain.CustomerAuthToken{
		ID:         uuid.New(),
		CustomerID: customerID,
		TokenHash:  hashed,
		Type:       "email_verify",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("store verify token: %w", err)
	}

	return s.emailService.SendVerification(email, raw)
}

// Login authenticates a customer with email and password. Returns a JWT and
// customer record on success.
//
// Returns a specific "email not verified" error (code "email_not_verified") if
// the account exists but has not yet verified its email, so the frontend can
// offer a resend option.
func (s *CustomerAuthService) Login(ctx context.Context, req *domain.CustomerLoginRequest, remoteIP string) (*domain.CustomerLoginResponse, error) {
	ok, err := s.verifyTurnstile(req.TurnstileToken, remoteIP)
	if err != nil {
		return nil, fmt.Errorf("turnstile check: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}

	customer, err := s.customerRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if customer.PasswordHash == nil {
		// Google-only account — no password set
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*customer.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !customer.EmailVerified {
		return nil, fmt.Errorf("email_not_verified")
	}

	jwtToken, err := s.generateCustomerJWT(customer)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &domain.CustomerLoginResponse{
		Token:    jwtToken,
		Customer: *customer,
	}, nil
}

// VerifyEmailCheck reports whether a raw email-verification token is valid
// (exists, not expired, not used). Used by the GET endpoint to drive the
// confirmation page UI before the user clicks "Confirm".
func (s *CustomerAuthService) VerifyEmailCheck(ctx context.Context, rawToken string) (bool, error) {
	hashed := hashToken(rawToken)
	token, err := s.tokenRepo.GetByHash(ctx, hashed)
	if err != nil || token == nil {
		return false, nil
	}
	return true, nil
}

// VerifyEmail consumes an email-verification token, marks the customer's email
// as verified, and returns a JWT so the user is automatically logged in.
func (s *CustomerAuthService) VerifyEmail(ctx context.Context, rawToken string) (*domain.CustomerLoginResponse, error) {
	hashed := hashToken(rawToken)
	token, err := s.tokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	if err := s.tokenRepo.MarkUsed(ctx, token.ID); err != nil {
		return nil, fmt.Errorf("mark token used: %w", err)
	}

	if err := s.customerRepo.UpdateEmailVerified(ctx, token.CustomerID, true); err != nil {
		return nil, fmt.Errorf("update email verified: %w", err)
	}

	customer, err := s.customerRepo.GetByID(ctx, token.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("fetch customer: %w", err)
	}

	jwtToken, err := s.generateCustomerJWT(customer)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &domain.CustomerLoginResponse{
		Token:    jwtToken,
		Customer: *customer,
	}, nil
}

// ResendVerification re-issues an email-verification link. Silently succeeds
// when the email is not found to prevent email enumeration.
func (s *CustomerAuthService) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(email)

	customer, err := s.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		// Not found — silently succeed
		return nil
	}

	if customer.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	return s.issueVerificationToken(ctx, customer.ID, email)
}

// ForgotPassword sends a password-reset email. Silently succeeds when the
// email is not found to prevent email enumeration.
func (s *CustomerAuthService) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest, remoteIP string) error {
	ok, err := s.verifyTurnstile(req.TurnstileToken, remoteIP)
	if err != nil {
		return fmt.Errorf("turnstile check: %w", err)
	}
	if !ok {
		return fmt.Errorf("turnstile verification failed")
	}

	email := strings.ToLower(req.Email)

	customer, err := s.customerRepo.GetByEmail(ctx, email)
	if err != nil {
		// Not found — silently succeed (prevent enumeration)
		return nil
	}

	if err := s.tokenRepo.DeleteByCustomerAndType(ctx, customer.ID, "password_reset"); err != nil {
		return fmt.Errorf("delete old reset tokens: %w", err)
	}

	raw, hashed, err := generateToken()
	if err != nil {
		return err
	}

	token := &domain.CustomerAuthToken{
		ID:         uuid.New(),
		CustomerID: customer.ID,
		TokenHash:  hashed,
		Type:       "password_reset",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	return s.emailService.SendPasswordReset(email, raw)
}

// ResetPassword validates a password-reset token and updates the customer's
// password hash.
func (s *CustomerAuthService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	hashed := hashToken(req.Token)
	token, err := s.tokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	if err := s.tokenRepo.MarkUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.customerRepo.UpdatePasswordHash(ctx, token.CustomerID, string(newHash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

// GoogleAuthURL returns the Google OAuth consent-screen URL and the random
// state token (which the handler must persist in a short-lived cookie for CSRF
// protection).
func (s *CustomerAuthService) GoogleAuthURL() (authURL string, state string, err error) {
	rawState, _, genErr := generateToken()
	if genErr != nil {
		return "", "", fmt.Errorf("generate oauth state: %w", genErr)
	}
	cfg := s.googleOAuthConfig()
	authURL = cfg.AuthCodeURL(rawState, oauth2.AccessTypeOnline)
	return authURL, rawState, nil
}

// GoogleCallback exchanges the OAuth code for a token, fetches the Google user
// profile, and returns a JWT. Handles three cases:
//  1. Existing Google-linked account — generates JWT directly.
//  2. Existing email account — links Google and generates JWT.
//  3. New user — creates an account with email_verified=true and generates JWT.
//
// The second return value is the redirect URL the handler should send the user
// to after storing the JWT.
func (s *CustomerAuthService) GoogleCallback(ctx context.Context, code, state string) (*domain.CustomerLoginResponse, string, error) {
	cfg := s.googleOAuthConfig()

	oauthToken, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("exchange oauth code: %w", err)
	}

	client := cfg.Client(ctx, oauthToken)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, "", fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, "", fmt.Errorf("decode userinfo: %w", err)
	}

	email := strings.ToLower(userInfo.Email)

	// Case 1: existing Google-linked account
	customer, err := s.customerRepo.GetByGoogleID(ctx, userInfo.ID)
	if err == nil {
		jwtToken, err := s.generateCustomerJWT(customer)
		if err != nil {
			return nil, "", fmt.Errorf("generate token: %w", err)
		}
		return &domain.CustomerLoginResponse{Token: jwtToken, Customer: *customer}, s.emailService.baseURL + "/devices", nil
	}

	// Case 2: existing email account — link Google
	customer, err = s.customerRepo.GetByEmail(ctx, email)
	if err == nil {
		if err := s.customerRepo.LinkGoogleAccount(ctx, customer.ID, userInfo.ID, userInfo.Email); err != nil {
			return nil, "", fmt.Errorf("link google account: %w", err)
		}
		// Refresh customer record to pick up updated fields
		customer, err = s.customerRepo.GetByID(ctx, customer.ID)
		if err != nil {
			return nil, "", fmt.Errorf("fetch customer after link: %w", err)
		}
		jwtToken, err := s.generateCustomerJWT(customer)
		if err != nil {
			return nil, "", fmt.Errorf("generate token: %w", err)
		}
		return &domain.CustomerLoginResponse{Token: jwtToken, Customer: *customer}, s.emailService.baseURL + "/devices", nil
	}

	// Case 3: new user
	namePart := userInfo.Name
	if namePart == "" {
		namePart = email
		if idx := strings.Index(email, "@"); idx > 0 {
			namePart = email[:idx]
		}
	}
	googleID := userInfo.ID
	googleEmail := userInfo.Email
	customer = &domain.Customer{
		ID:            uuid.New(),
		Name:          namePart,
		Email:         email,
		Active:        true,
		PasswordHash:  nil,
		EmailVerified: true, // Google has confirmed email ownership
		GoogleID:      &googleID,
		GoogleEmail:   &googleEmail,
	}
	if err := s.customerRepo.CreateWithAuth(ctx, customer); err != nil {
		return nil, "", fmt.Errorf("create google customer: %w", err)
	}

	jwtToken, err := s.generateCustomerJWT(customer)
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	return &domain.CustomerLoginResponse{Token: jwtToken, Customer: *customer}, s.emailService.baseURL + "/devices", nil
}
