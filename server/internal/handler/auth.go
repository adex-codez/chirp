package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backend/internal/auth"
	"backend/internal/response"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type authService interface {
	SignUp(ctx context.Context, username, email, password, deviceLabel string) (service.SignUpResult, error)
	VerifyEmail(ctx context.Context, email, code, deviceLabel string) (service.AuthResult, error)
	ResendVerification(ctx context.Context, email string) (service.SignUpResult, error)
	SignIn(ctx context.Context, email, password, deviceLabel string) (service.AuthResult, error)
	Profile(ctx context.Context, accessToken string) (service.PublicUser, error)
	Refresh(ctx context.Context, refreshToken, deviceLabel string) (service.AuthResult, error)
	SignOut(ctx context.Context, refreshToken string) error
	SignOutAll(ctx context.Context, accessToken string) error
	SocialSignIn(ctx context.Context, provider, idToken, nonce, deviceLabel string) (service.SocialResult, error)
	SetUsername(ctx context.Context, pendingToken, username string) (service.AuthResult, error)
}

// AuthHandler is the HTTP transport for Password sign-in joins, verification,
// sign-ins, and the current-User read. It maps service outcomes to status
// codes and keeps sign-in failures generic.
type AuthHandler struct {
	service authService
}

// NewAuthHandler wires the handler to its service.
func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{service: svc}
}

type signUpRequest struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DeviceLabel string `json:"device_label"`
}

type verifyRequest struct {
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	DeviceLabel string `json:"device_label"`
}

type resendRequest struct {
	Email string `json:"email" binding:"required"`
}

type signInRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DeviceLabel string `json:"device_label"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	DeviceLabel  string `json:"device_label"`
}

type signOutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type socialRequest struct {
	Provider    string `json:"provider" binding:"required"`
	IDToken     string `json:"id_token" binding:"required"`
	Nonce       string `json:"nonce"`
	DeviceLabel string `json:"device_label"`
}

type setUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

// SignUp handles POST /auth/signup.
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "username, email, and password are required")
		return
	}
	result, err := h.service.SignUp(c.Request.Context(), req.Username, req.Email, req.Password, req.DeviceLabel)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

// Verify handles POST /auth/verify.
func (h *AuthHandler) Verify(c *gin.Context) {
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "email and code are required")
		return
	}
	result, err := h.service.VerifyEmail(c.Request.Context(), req.Email, req.Code, req.DeviceLabel)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

// Resend handles POST /auth/verify/resend.
func (h *AuthHandler) Resend(c *gin.Context) {
	var req resendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "email is required")
		return
	}
	result, err := h.service.ResendVerification(c.Request.Context(), req.Email)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

// SignIn handles POST /auth/login.
func (h *AuthHandler) SignIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "email and password are required")
		return
	}
	result, err := h.service.SignIn(c.Request.Context(), req.Email, req.Password, req.DeviceLabel)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

// Profile handles GET /auth/me.
func (h *AuthHandler) Profile(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	user, err := h.service.Profile(c.Request.Context(), token)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, user)
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "refresh token is required")
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceLabel)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

// SignOut handles POST /auth/logout. Unknown tokens succeed idempotently.
func (h *AuthHandler) SignOut(c *gin.Context) {
	var req signOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "refresh token is required")
		return
	}
	if err := h.service.SignOut(c.Request.Context(), req.RefreshToken); err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, gin.H{"signed_out": true})
}

// SignOutAll handles POST /auth/logout-all.
func (h *AuthHandler) SignOutAll(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	if err := h.service.SignOutAll(c.Request.Context(), token); err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, gin.H{"signed_out_everywhere": true})
}

// SocialSignIn handles POST /auth/social. It answers either a full session
// or a pending grant for the Username picker.
func (h *AuthHandler) SocialSignIn(c *gin.Context) {
	var req socialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "provider and id token are required")
		return
	}
	result, err := h.service.SocialSignIn(c.Request.Context(), req.Provider, req.IDToken, req.Nonce, req.DeviceLabel)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	if result.Pending {
		response.Success(c, gin.H{
			"user":          result.User,
			"pending":       true,
			"pending_token": result.PendingToken,
		})
		return
	}
	response.Success(c, result.Session)
}

// SetUsername handles POST /auth/username. Only a pending grant authorizes
// it; full sessions manage the Username elsewhere.
func (h *AuthHandler) SetUsername(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	var req setUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "username is required")
		return
	}
	result, err := h.service.SetUsername(c.Request.Context(), token, req.Username)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, result)
}

func bearerToken(header string) (string, bool) {
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || strings.TrimSpace(token) == "" {
		return "", false
	}
	return token, true
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUsernameTaken):
		response.Error(c, http.StatusConflict, "username is already taken")
	case errors.Is(err, service.ErrEmailTaken):
		response.Error(c, http.StatusConflict, "email is already in use")
	case errors.Is(err, service.ErrInvalidSignIn):
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, service.ErrInvalidSession):
		response.Error(c, http.StatusUnauthorized, "invalid or expired session")
	case errors.Is(err, service.ErrSessionRevoked):
		response.Error(c, http.StatusUnauthorized, "session was revoked, sign in again")
	case errors.Is(err, service.ErrUserNotFound):
		response.Error(c, http.StatusNotFound, "user not found")
	case errors.Is(err, service.ErrEmailNotVerified):
		response.Error(c, http.StatusForbidden, "email is not verified, check your inbox for the code")
	case errors.Is(err, service.ErrInvalidChallenge):
		response.Error(c, http.StatusBadRequest, "invalid or expired code")
	case errors.Is(err, service.ErrChallengeLocked):
		response.Error(c, http.StatusTooManyRequests, "too many attempts, request a new code")
	case errors.Is(err, service.ErrNoPendingVerify):
		response.Error(c, http.StatusNotFound, "no pending verification for this email")
	case errors.Is(err, service.ErrAlreadyVerified):
		response.Error(c, http.StatusBadRequest, "email is already verified")
	case errors.Is(err, service.ErrResendTooSoon):
		response.Error(c, http.StatusTooManyRequests, "verification code was just sent, wait a minute")
	case errors.Is(err, service.ErrResendLimit):
		response.Error(c, http.StatusTooManyRequests, "too many codes requested, try again later")
	case errors.Is(err, service.ErrSocialMisconfigured):
		response.Error(c, http.StatusServiceUnavailable, "social sign-in is not configured")
	case errors.Is(err, service.ErrSocialNoEmail):
		response.Error(c, http.StatusBadRequest, "no email address came with this social sign-in")
	case errors.Is(err, service.ErrSocialUnverified):
		response.Error(c, http.StatusForbidden, "this email is not verified by the provider")
	case errors.Is(err, service.ErrSocialConflict):
		response.Error(c, http.StatusConflict, "an account with this email already exists, sign in with your password first")
	case errors.Is(err, auth.ErrSocialTokenInvalid):
		response.Error(c, http.StatusUnauthorized, "invalid social token")
	case errors.Is(err, auth.ErrSocialUnavailable):
		response.Error(c, http.StatusBadGateway, "could not verify social token")
	case errors.Is(err, service.ErrValidation):
		response.Error(c, http.StatusBadRequest, validationMessage(err))
	default:
		response.Error(c, http.StatusInternalServerError, "something went wrong")
	}
}

func validationMessage(err error) string {
	var validation service.ValidationError
	if errors.As(err, &validation) {
		return validation.Detail
	}
	return "invalid request"
}
