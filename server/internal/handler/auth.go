package handler

import (
	"context"
	"strings"

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
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, code, password string) error
	AddPassword(ctx context.Context, accessToken, password string) error
	ChangeUsername(ctx context.Context, accessToken, username string) (service.PublicUser, error)
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

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type addPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

type changeUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

// SignUp handles POST /auth/signup.
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a username, email and password."})
		return
	}
	result, err := h.service.SignUp(c.Request.Context(), req.Username, req.Email, req.Password, req.DeviceLabel)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// Verify handles POST /auth/verify.
func (h *AuthHandler) Verify(c *gin.Context) {
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with an email and code."})
		return
	}
	result, err := h.service.VerifyEmail(c.Request.Context(), req.Email, req.Code, req.DeviceLabel)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// Resend handles POST /auth/verify/resend.
func (h *AuthHandler) Resend(c *gin.Context) {
	var req resendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with an email."})
		return
	}
	result, err := h.service.ResendVerification(c.Request.Context(), req.Email)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// SignIn handles POST /auth/login.
func (h *AuthHandler) SignIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with an email and password."})
		return
	}
	result, err := h.service.SignIn(c.Request.Context(), req.Email, req.Password, req.DeviceLabel)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// Profile handles GET /auth/me.
func (h *AuthHandler) Profile(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	user, err := h.service.Profile(c.Request.Context(), token)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, user)
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a refresh token."})
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceLabel)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// SignOut handles POST /auth/logout. Unknown tokens succeed idempotently.
func (h *AuthHandler) SignOut(c *gin.Context) {
	var req signOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a refresh token."})
		return
	}
	if err := h.service.SignOut(c.Request.Context(), req.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"signed_out": true})
}

// SignOutAll handles POST /auth/logout-all.
func (h *AuthHandler) SignOutAll(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	if err := h.service.SignOutAll(c.Request.Context(), token); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"signed_out_everywhere": true})
}

// SocialSignIn handles POST /auth/social. It answers either a full session
// or a pending grant for the Username picker.
func (h *AuthHandler) SocialSignIn(c *gin.Context) {
	var req socialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a provider and id token."})
		return
	}
	result, err := h.service.SocialSignIn(c.Request.Context(), req.Provider, req.IDToken, req.Nonce, req.DeviceLabel)
	if err != nil {
		response.Fail(c, err)
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
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	var req setUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a username."})
		return
	}
	result, err := h.service.SetUsername(c.Request.Context(), token, req.Username)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

// ForgotPassword handles POST /auth/password/forgot. It always answers
// success so addresses cannot be enumerated; only verified Emails get a
// challenge.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with an email."})
		return
	}
	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"reset_sent": true})
}

// ResetPassword handles POST /auth/password/reset.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with an email, code and new password."})
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"password_reset": true})
}

// AddPassword handles POST /auth/password for signed-in Users without one.
func (h *AuthHandler) AddPassword(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	var req addPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a new password."})
		return
	}
	if err := h.service.AddPassword(c.Request.Context(), token, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"password_added": true})
}

// ChangeUsername handles POST /auth/username/change for signed-in Users.
func (h *AuthHandler) ChangeUsername(c *gin.Context) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return
	}
	var req changeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Validation(c, map[string]string{"body": "Send a JSON body with a username."})
		return
	}
	user, err := h.service.ChangeUsername(c.Request.Context(), token, req.Username)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, user)
}

func bearerToken(header string) (string, bool) {
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || strings.TrimSpace(token) == "" {
		return "", false
	}
	return token, true
}
