package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrAuthNotFound is returned when a User, challenge, or session is absent.
var ErrAuthNotFound = errors.New("not found")

// AuthUser is the persistence view of a User: unique Username, unique Email,
// nullable password hash (social-only Users have none), and verification state.
type AuthUser struct {
	ID            string
	Username      string
	Email         string
	PasswordHash  string
	HasPassword   bool
	EmailVerified bool
}

// VerificationChallenge is a single-use code grant for a purpose such as
// email verification.
type VerificationChallenge struct {
	ID        string
	UserID    string
	Purpose   string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	CreatedAt time.Time
}

// AuthRepository owns User, challenge, and session persistence.
type AuthRepository interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (AuthUser, error)
	GetUserByUsername(ctx context.Context, username string) (AuthUser, error)
	GetUserByID(ctx context.Context, id string) (AuthUser, error)
	MarkUserVerified(ctx context.Context, id string) error
	CreateVerificationChallenge(ctx context.Context, userID, purpose, codeHash string, expiresAt time.Time) (VerificationChallenge, error)
	GetLatestVerificationChallenge(ctx context.Context, userID, purpose string) (VerificationChallenge, error)
	CountRecentVerificationChallenges(ctx context.Context, userID, purpose string, since time.Time) (int, error)
	ConsumeVerificationChallenge(ctx context.Context, id string) error
	IncrementChallengeAttempts(ctx context.Context, id string) error
	CreateSession(ctx context.Context, userID, refreshHash string, expiresAt time.Time, deviceLabel string) (string, error)
}

// PostgresAuthRepository is the PostgreSQL adapter for AuthRepository.
type PostgresAuthRepository struct {
	queries *database.Queries
}

// NewAuthRepository wires the generated queries to a live pool.
func NewAuthRepository(db *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{queries: database.New(db)}
}

func (r *PostgresAuthRepository) CreateUser(ctx context.Context, username, email, passwordHash string) (AuthUser, error) {
	hash := pgtype.Text{Valid: false}
	if passwordHash != "" {
		hash = pgtype.Text{String: passwordHash, Valid: true}
	}
	row, err := r.queries.CreateUser(ctx, database.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		return AuthUser{}, err
	}
	return AuthUser{
		ID:            row.ID,
		Username:      row.Username,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash.String,
		HasPassword:   row.PasswordHash.Valid,
		EmailVerified: row.EmailVerifiedAt.Valid,
	}, nil
}

func (r *PostgresAuthRepository) GetUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthUser{}, mapNoRows(err)
	}
	return userRow(row.ID, row.Username, row.Email, row.PasswordHash, row.EmailVerifiedAt), nil
}

func (r *PostgresAuthRepository) GetUserByUsername(ctx context.Context, username string) (AuthUser, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return AuthUser{}, mapNoRows(err)
	}
	return userRow(row.ID, row.Username, row.Email, row.PasswordHash, row.EmailVerifiedAt), nil
}

func (r *PostgresAuthRepository) GetUserByID(ctx context.Context, id string) (AuthUser, error) {
	key, err := parseUUID(id)
	if err != nil {
		return AuthUser{}, ErrAuthNotFound
	}
	row, err := r.queries.GetUserByID(ctx, key)
	if err != nil {
		return AuthUser{}, mapNoRows(err)
	}
	return userRow(row.ID, row.Username, row.Email, row.PasswordHash, row.EmailVerifiedAt), nil
}

func (r *PostgresAuthRepository) MarkUserVerified(ctx context.Context, id string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.queries.MarkUserVerified(ctx, key)
}

func (r *PostgresAuthRepository) CreateVerificationChallenge(ctx context.Context, userID, purpose, codeHash string, expiresAt time.Time) (VerificationChallenge, error) {
	key, err := parseUUID(userID)
	if err != nil {
		return VerificationChallenge{}, err
	}
	row, err := r.queries.CreateVerificationChallenge(ctx, database.CreateVerificationChallengeParams{
		Column1:   key,
		Purpose:   purpose,
		CodeHash:  codeHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return VerificationChallenge{}, err
	}
	return VerificationChallenge{
		ID:        row.ID,
		UserID:    row.UserID,
		Purpose:   row.Purpose,
		CodeHash:  row.CodeHash,
		ExpiresAt: row.ExpiresAt.Time,
		Attempts:  int(row.Attempts),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *PostgresAuthRepository) GetLatestVerificationChallenge(ctx context.Context, userID, purpose string) (VerificationChallenge, error) {
	key, err := parseUUID(userID)
	if err != nil {
		return VerificationChallenge{}, ErrAuthNotFound
	}
	row, err := r.queries.GetLatestVerificationChallenge(ctx, database.GetLatestVerificationChallengeParams{
		Column1: key,
		Purpose: purpose,
	})
	if err != nil {
		return VerificationChallenge{}, mapNoRows(err)
	}
	return VerificationChallenge{
		ID:        row.ID,
		UserID:    row.UserID,
		Purpose:   row.Purpose,
		CodeHash:  row.CodeHash,
		ExpiresAt: row.ExpiresAt.Time,
		Attempts:  int(row.Attempts),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *PostgresAuthRepository) CountRecentVerificationChallenges(ctx context.Context, userID, purpose string, since time.Time) (int, error) {
	key, err := parseUUID(userID)
	if err != nil {
		return 0, ErrAuthNotFound
	}
	count, err := r.queries.CountRecentVerificationChallenges(ctx, database.CountRecentVerificationChallengesParams{
		Column1:   key,
		Purpose:   purpose,
		CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *PostgresAuthRepository) ConsumeVerificationChallenge(ctx context.Context, id string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.queries.ConsumeVerificationChallenge(ctx, key)
}

func (r *PostgresAuthRepository) IncrementChallengeAttempts(ctx context.Context, id string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.queries.IncrementChallengeAttempts(ctx, key)
}

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, userID, refreshHash string, expiresAt time.Time, deviceLabel string) (string, error) {
	key, err := parseUUID(userID)
	if err != nil {
		return "", err
	}
	row, err := r.queries.CreateSession(ctx, database.CreateSessionParams{
		Column1:     key,
		RefreshHash: refreshHash,
		ExpiresAt:   pgtype.Timestamptz{Time: expiresAt, Valid: true},
		DeviceLabel: deviceLabel,
	})
	if err != nil {
		return "", err
	}
	return row.ID, nil
}

func userRow(id, username, email string, hash pgtype.Text, verifiedAt pgtype.Timestamptz) AuthUser {
	return AuthUser{
		ID:            id,
		Username:      username,
		Email:         email,
		PasswordHash:  hash.String,
		HasPassword:   hash.Valid,
		EmailVerified: verifiedAt.Valid,
	}
}

func parseUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func mapNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAuthNotFound
	}
	return err
}
