package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"backend/internal/profilepicture"
)

var ErrProfilePictureUnavailable = errors.New("profile picture uploads are unavailable")

func (s *AuthService) UploadProfilePicture(ctx context.Context, token string, body io.Reader) (PublicUser, error) {
	user, err := s.Profile(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidSession) || errors.Is(err, ErrUserNotFound) {
			return PublicUser{}, ErrInvalidSession
		}
		return PublicUser{}, fmt.Errorf("%w: authorize picture upload: %w", ErrProfilePictureUnavailable, err)
	}
	if s.pictures == nil {
		return PublicUser{}, ErrProfilePictureUnavailable
	}
	picture, err := profilepicture.ReadJPEG(body)
	if err != nil {
		return PublicUser{}, ValidationError{Fields: map[string]string{"picture": err.Error()}}
	}
	newURL, err := s.pictures.Upload(ctx, user.ID, picture)
	if err != nil {
		return PublicUser{}, fmt.Errorf("%w: upload picture: %w", ErrProfilePictureUnavailable, err)
	}
	replacement, err := s.users.ReplaceProfilePicture(ctx, user.ID, newURL)
	if err != nil {
		// A failed commit can already be durable. Deleting this candidate could delete the current picture.
		return PublicUser{}, fmt.Errorf("%w: save picture reference %q, candidate retained: %w", ErrProfilePictureUnavailable, newURL, err)
	}
	if replacement.PreviousURL != "" {
		if err := s.pictures.DeleteOwned(context.WithoutCancel(ctx), user.ID, replacement.PreviousURL); err != nil {
			slog.WarnContext(ctx, "retired picture cleanup failed", "user_id", user.ID, "error", err)
		}
	}
	return publicUser(replacement.User), nil
}
