package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/service"
)

// CustomerAuthHandler handles HTTP requests for all customer authentication
// endpoints: signup, login, email verification, password reset, and Google OAuth.
type CustomerAuthHandler struct {
	customerAuthService *service.CustomerAuthService
}

// NewCustomerAuthHandler constructs a CustomerAuthHandler.
func NewCustomerAuthHandler(s *service.CustomerAuthService) *CustomerAuthHandler {
	return &CustomerAuthHandler{customerAuthService: s}
}

// Signup handles POST /api/auth/customer/signup.
// Creates a new customer account and sends a verification email.
func (h *CustomerAuthHandler) Signup(c *gin.Context) {
	var req domain.CustomerSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.customerAuthService.Signup(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		if err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": "an account with that email already exists"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "check your email to verify your account"})
}

// Login handles POST /api/auth/customer/login.
// Returns a JWT on success. Returns 403 with an "email_not_verified" code if
// the account exists but has not yet verified its email, so the frontend can
// offer a "resend verification" option.
func (h *CustomerAuthHandler) Login(c *gin.Context) {
	var req domain.CustomerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.customerAuthService.Login(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		if err.Error() == "email_not_verified" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Please verify your email first",
				"code":  "email_not_verified",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyEmailCheck handles GET /api/auth/customer/verify-email?token=X.
// Reports whether the given token is valid (exists, not expired, not used),
// so the frontend can show an appropriate confirmation page before the user
// clicks "Confirm".
func (h *CustomerAuthHandler) VerifyEmailCheck(c *gin.Context) {
	rawToken := c.Query("token")
	if rawToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token query parameter is required"})
		return
	}

	valid, err := h.customerAuthService.VerifyEmailCheck(c.Request.Context(), rawToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if valid {
		c.JSON(http.StatusOK, gin.H{"valid": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"valid": false, "reason": "expired or invalid"})
	}
}

// VerifyEmail handles POST /api/auth/customer/verify-email.
// Consumes the token, marks the customer's email as verified, and returns a
// JWT for automatic login.
func (h *CustomerAuthHandler) VerifyEmail(c *gin.Context) {
	var req domain.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.customerAuthService.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired verification token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ResendVerification handles POST /api/auth/customer/resend-verification.
// Always returns a generic 200 message regardless of whether the email exists,
// to prevent email enumeration.
func (h *CustomerAuthHandler) ResendVerification(c *gin.Context) {
	var req domain.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.customerAuthService.ResendVerification(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{"message": "if that email is registered, a verification link has been sent"})
}

// ForgotPassword handles POST /api/auth/customer/forgot-password.
// Always returns a generic 200 message regardless of whether the email exists,
// to prevent email enumeration.
func (h *CustomerAuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.customerAuthService.ForgotPassword(c.Request.Context(), &req, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{"message": "if that email is registered, a password reset link has been sent"})
}

// ResetPassword handles POST /api/auth/customer/reset-password.
// Validates the token and updates the password.
func (h *CustomerAuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.customerAuthService.ResetPassword(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// GoogleLogin handles GET /api/auth/customer/google.
// Redirects the user to the Google OAuth consent screen. Sets a short-lived
// `oauth_state` cookie for CSRF protection.
func (h *CustomerAuthHandler) GoogleLogin(c *gin.Context) {
	authURL, state, err := h.customerAuthService.GoogleAuthURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build auth URL"})
		return
	}

	// Store the state in a short-lived HttpOnly cookie for CSRF verification
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles GET /api/auth/customer/google/callback.
// Verifies the state cookie, exchanges the OAuth code, and redirects the user
// to the dashboard login page with the JWT attached as a query parameter.
func (h *CustomerAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state cookie"})
		return
	}

	if state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	// Delete the state cookie immediately after verification
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	resp, redirectBase, err := h.customerAuthService.GoogleCallback(c.Request.Context(), code, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google authentication failed"})
		return
	}

	// Redirect to dashboard login page; frontend reads token from query param
	redirectURL := fmt.Sprintf("%s/login?token=%s&google=true", redirectBase, resp.Token)
	c.Redirect(http.StatusFound, redirectURL)
}
