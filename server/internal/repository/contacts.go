package repository

import (
	"context"

	"backend/internal/database/generated"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Contact struct {
	ID                string
	Username          string
	ProfilePictureURL string
}

type ContactRepository interface {
	FindContactableUserByUsername(ctx context.Context, username string) (Contact, error)
	AddContact(ctx context.Context, userID, contactID string) error
	ListContacts(ctx context.Context, userID string) ([]Contact, error)
}

type PostgresContactRepository struct {
	queries *database.Queries
}

func NewContactRepository(db *pgxpool.Pool) *PostgresContactRepository {
	return &PostgresContactRepository{queries: database.New(db)}
}

func (r *PostgresContactRepository) FindContactableUserByUsername(ctx context.Context, username string) (Contact, error) {
	row, err := r.queries.GetContactableUserByUsername(ctx, username)
	if err != nil {
		return Contact{}, mapNoRows(err)
	}
	return Contact{ID: row.ID, Username: row.Username}, nil
}

func (r *PostgresContactRepository) AddContact(ctx context.Context, userID, contactID string) error {
	userKey, err := parseUUID(userID)
	if err != nil {
		return err
	}
	contactKey, err := parseUUID(contactID)
	if err != nil {
		return err
	}
	return r.queries.AddContact(ctx, database.AddContactParams{
		UserID:    userKey,
		ContactID: contactKey,
	})
}

func (r *PostgresContactRepository) ListContacts(ctx context.Context, userID string) ([]Contact, error) {
	userKey, err := parseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListContacts(ctx, userKey)
	if err != nil {
		return nil, err
	}
	contacts := make([]Contact, 0, len(rows))
	for _, row := range rows {
		contacts = append(contacts, Contact{ID: row.ID, Username: row.Username.String, ProfilePictureURL: row.ProfilePictureUrl.String})
	}
	return contacts, nil
}
