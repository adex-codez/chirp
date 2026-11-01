package service

import (
	"context"
	"errors"
	"strings"

	"backend/internal/repository"
)

type Contact struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ContactsService struct {
	contacts repository.ContactRepository
}

func NewContactsService(contacts repository.ContactRepository) *ContactsService {
	return &ContactsService{contacts: contacts}
}

func (s *ContactsService) Add(ctx context.Context, userID, username string) (Contact, error) {
	username = strings.TrimSpace(username)
	if err := validateUsername(username); err != nil {
		return Contact{}, err
	}

	contact, err := s.contacts.FindContactableUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return Contact{}, ErrUserNotFound
		}
		return Contact{}, err
	}
	if contact.ID == userID {
		return Contact{}, ValidationError{Fields: map[string]string{"username": "you cannot add yourself"}}
	}
	if err := s.contacts.AddContact(ctx, userID, contact.ID); err != nil {
		return Contact{}, err
	}
	return Contact{ID: contact.ID, Username: contact.Username}, nil
}

func (s *ContactsService) List(ctx context.Context, userID string) ([]Contact, error) {
	contacts, err := s.contacts.ListContacts(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]Contact, 0, len(contacts))
	for _, contact := range contacts {
		result = append(result, Contact{ID: contact.ID, Username: contact.Username})
	}
	return result, nil
}
