package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/database/generated"
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

// Session is the persistence view of a refresh grant: its hash, expiry,
// revocation state, and rotation lineage (ReplacedBy is empty until rotated).
type Session struct {
	ID         string
	UserID     string
	ExpiresAt  time.Time
	Revoked    bool
	ReplacedBy string
}

// LinkedIdentity is a provider identity attached to a User: which provider
// and which subject, plus the address seen at link time.
type LinkedIdentity struct {
	ID          string
	UserID      string
	Provider    string
	ProviderSub string
	Email       string
}

// UserRepository owns User persistence: lookup, creation, verification
// state, and credential changes.
type UserRepository interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (AuthUser, error)
	GetUserByEmail(ctx context.Context, email string) (AuthUser, error)
	GetUserByUsername(ctx context.Context, username string) (AuthUser, error)
	GetUserByID(ctx context.Context, id string) (AuthUser, error)
	MarkUserVerified(ctx context.Context, id string) error
	UpdateUsername(ctx context.Context, id, username string) (AuthUser, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
}

// ChallengeRepository owns single-use verification-code grants.
type ChallengeRepository interface {
	CreateVerificationChallenge(ctx context.Context, userID, purpose, codeHash string, expiresAt time.Time) (VerificationChallenge, error)
	GetLatestVerificationChallenge(ctx context.Context, userID, purpose string) (VerificationChallenge, error)
	CountRecentVerificationChallenges(ctx context.Context, userID, purpose string, since time.Time) (int, error)
	ConsumeVerificationChallenge(ctx context.Context, id string) error
	IncrementChallengeAttempts(ctx context.Context, id string) error
}

// SessionRepository owns refresh grants: issuance, rotation, and revocation.
type SessionRepository interface {
	CreateSession(ctx context.Context, userID, refreshHash string, expiresAt time.Time, deviceLabel string) (string, error)
	GetSessionByRefreshHash(ctx context.Context, refreshHash string) (Session, error)
	ReplaceSession(ctx context.Context, id, replacedBy string) error
	RevokeSession(ctx context.Context, id string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
}

// IdentityRepository owns provider identities linked to Users.
type IdentityRepository interface {
	GetIdentity(ctx context.Context, provider, providerSub string) (LinkedIdentity, error)
	CreateIdentity(ctx context.Context, userID, provider, providerSub, email string) (LinkedIdentity, error)
}

// AuthRepository is the full persistence contract: the four narrow stores
// above. PostgresAuthRepository implements all of them; consumers depend on
// the narrow store they actually use.
type AuthRepository interface {
	UserRepository
	ChallengeRepository
	SessionRepository
	IdentityRepository
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
		Column1:      username,
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

func (r *PostgresAuthRepository) GetSessionByRefreshHash(ctx context.Context, refreshHash string) (Session, error) {
	row, err := r.queries.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return Session{}, mapNoRows(err)
	}
	return Session{
		ID:         row.ID,
		UserID:     row.UserID,
		ExpiresAt:  row.ExpiresAt.Time,
		Revoked:    row.RevokedAt.Valid,
		ReplacedBy: row.ReplacedBy.String,
	}, nil
}

func (r *PostgresAuthRepository) ReplaceSession(ctx context.Context, id, replacedBy string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	replacement, err := parseUUID(replacedBy)
	if err != nil {
		return err
	}
	return r.queries.ReplaceSession(ctx, database.ReplaceSessionParams{
		Column1: key,
		Column2: replacement,
	})
}

func (r *PostgresAuthRepository) RevokeSession(ctx context.Context, id string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.queries.RevokeSession(ctx, key)
}

func (r *PostgresAuthRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	key, err := parseUUID(userID)
	if err != nil {
		return err
	}
	return r.queries.RevokeAllUserSessions(ctx, key)
}

func (r *PostgresAuthRepository) GetIdentity(ctx context.Context, provider, providerSub string) (LinkedIdentity, error) {
	row, err := r.queries.GetIdentity(ctx, database.GetIdentityParams{
		Provider:    provider,
		ProviderSub: providerSub,
	})
	if err != nil {
		return LinkedIdentity{}, mapNoRows(err)
	}
	return LinkedIdentity{
		ID:          row.ID,
		UserID:      row.UserID,
		Provider:    row.Provider,
		ProviderSub: row.ProviderSub,
		Email:       row.Email.String,
	}, nil
}

func (r *PostgresAuthRepository) CreateIdentity(ctx context.Context, userID, provider, providerSub, email string) (LinkedIdentity, error) {
	key, err := parseUUID(userID)
	if err != nil {
		return LinkedIdentity{}, err
	}
	mail := pgtype.Text{Valid: false}
	if email != "" {
		mail = pgtype.Text{String: email, Valid: true}
	}
	row, err := r.queries.CreateIdentity(ctx, database.CreateIdentityParams{
		Column1:     key,
		Provider:    provider,
		ProviderSub: providerSub,
		Email:       mail,
	})
	if err != nil {
		return LinkedIdentity{}, err
	}
	return LinkedIdentity{
		ID:          row.ID,
		UserID:      row.UserID,
		Provider:    row.Provider,
		ProviderSub: row.ProviderSub,
		Email:       row.Email.String,
	}, nil
}

func (r *PostgresAuthRepository) UpdateUsername(ctx context.Context, id, username string) (AuthUser, error) {
	key, err := parseUUID(id)
	if err != nil {
		return AuthUser{}, err
	}
	row, err := r.queries.UpdateUsername(ctx, database.UpdateUsernameParams{
		Column1:  key,
		Username: pgtype.Text{String: username, Valid: true},
	})
	if err != nil {
		return AuthUser{}, mapNoRows(err)
	}
	return userRow(row.ID, row.Username, row.Email, row.PasswordHash, row.EmailVerifiedAt), nil
}

func (r *PostgresAuthRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	key, err := parseUUID(id)
	if err != nil {
		return err
	}
	return r.queries.UpdatePassword(ctx, database.UpdatePasswordParams{
		Column1:      key,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
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
