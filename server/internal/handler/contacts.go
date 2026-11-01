package handler

import (
	"context"
	"strings"

	"backend/internal/response"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type profileReader interface {
	Profile(ctx context.Context, accessToken string) (service.PublicUser, error)
}

type contactsService interface {
	Add(ctx context.Context, userID, username string) (service.Contact, error)
	List(ctx context.Context, userID string) ([]service.Contact, error)
}

type ContactHandler struct {
	profile  profileReader
	contacts contactsService
}

func NewContactHandler(profile profileReader, contacts contactsService) *ContactHandler {
	return &ContactHandler{profile: profile, contacts: contacts}
}

type addContactRequest struct {
	Username string `json:"username" binding:"required"`
}

func (h *ContactHandler) List(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	contacts, err := h.contacts.List(c.Request.Context(), user.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, contacts)
}

func (h *ContactHandler) Add(c *gin.Context) {
	user, ok := h.currentUser(c)
	if !ok {
		return
	}
	var req addContactRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Username) == "" {
		response.Validation(c, map[string]string{"username": "Enter a username."})
		return
	}
	contact, err := h.contacts.Add(c.Request.Context(), user.ID, req.Username)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, contact)
}

func (h *ContactHandler) currentUser(c *gin.Context) (service.PublicUser, bool) {
	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		response.FailCode(c, response.CodeInvalidSession)
		return service.PublicUser{}, false
	}
	user, err := h.profile.Profile(c.Request.Context(), token)
	if err != nil {
		response.Fail(c, err)
		return service.PublicUser{}, false
	}
	return user, true
}
