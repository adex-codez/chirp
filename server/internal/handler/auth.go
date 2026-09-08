package handler

import (
	"context"
	"errors"
	"net/http"
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
}

// AuthHandler is the HTTP transport for Password sign-in joins, verification,
// and sign-ins. It maps service outcomes to status codes and keeps
// credential failures generic.
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

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUsernameTaken):
		response.Error(c, http.StatusConflict, "username is already taken")
	case errors.Is(err, service.ErrEmailTaken):
		response.Error(c, http.StatusConflict, "email is already in use")
	case errors.Is(err, service.ErrInvalidCredential):
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
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
	case errors.Is(err, service.ErrValidation):
		response.Error(c, http.StatusBadRequest, validationMessage(err))
	default:
		response.Error(c, http.StatusInternalServerError, "something went wrong")
	}
}

func validationMessage(err error) string {
	// Validation errors join the sentinel with the human detail; surface
	// the detail after the sentinel prefix.
	msg := err.Error()
	if rest, ok := strings.CutPrefix(msg, service.ErrValidation.Error()+"\n"); ok {
		return rest
	}
	if rest, ok := strings.CutPrefix(msg, service.ErrValidation.Error()+": "); ok {
		return rest
	}
	return "invalid request"
}
