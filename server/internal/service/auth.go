package service

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTokenTTL matches ADR-0002: short-lived access, renewed by refresh.
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL matches ADR-0002: 30-day rotating refresh.
	RefreshTokenTTL = 30 * 24 * time.Hour
	// VerificationCodeTTL bounds how long a code stays usable.
	VerificationCodeTTL = 20 * time.Minute
	// VerificationMaxAttempts caps guesses before the challenge locks.
	VerificationMaxAttempts = 5
	// VerificationResendCooldown spaces out re-sends.
	VerificationResendCooldown = time.Minute
)

const ChallengePurposeVerifyEmail = "verify_email"

var (
	ErrValidation        = errors.New("validation failed")
	ErrUsernameTaken     = errors.New("username is already taken")
	ErrEmailTaken        = errors.New("email is already in use")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrEmailNotVerified  = errors.New("email is not verified")
	ErrInvalidChallenge  = errors.New("invalid or expired code")
	ErrChallengeLocked   = errors.New("too many attempts, request a new code")
	ErrNoPendingVerify   = errors.New("no pending verification for this email")
	ErrAlreadyVerified   = errors.New("email is already verified")
	ErrResendTooSoon     = errors.New("verification code was just sent")
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.]{3,20}$`)

// PublicUser is the outward view of a User: handle, contact, verification state.
type PublicUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

// AuthResult is a signed-in session: the User plus its first token pair.
type AuthResult struct {
	User         PublicUser `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresIn    int        `json:"expires_in"`
}

// SignUpResult is the answer to a join: the User plus verification state.
// DevCode is populated only outside release builds, until a mail sender exists.
type SignUpResult struct {
	User             PublicUser `json:"user"`
	VerificationSent bool       `json:"verification_sent"`
	DevCode          string     `json:"dev_verification_code,omitempty"`
}

// AuthService contains the application behavior for joining, verifying,
// and signing in with Password sign-in.
type AuthService struct {
	repository     repository.AuthRepository
	jwtSecret      string
	devExposeCodes bool
	now            func() time.Time
}

// NewAuthService wires the service. now is nil in production (uses time.Now)
// and overridden in tests.
func NewAuthService(repo repository.AuthRepository, jwtSecret string, devExposeCodes bool) *AuthService {
	return &AuthService{
		repository:     repo,
		jwtSecret:      jwtSecret,
		devExposeCodes: devExposeCodes,
		now:            nil,
	}
}

func (s *AuthService) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// SignUp creates an unverified User and issues its first verification challenge.
// No session is issued until the Email is verified.
func (s *AuthService) SignUp(ctx context.Context, username, email, password, deviceLabel string) (SignUpResult, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if !usernamePattern.MatchString(username) {
		return SignUpResult{}, errors.Join(ErrValidation, errors.New("username must be 3-20 characters of letters, numbers, underscore, or dot"))
	}
	if err := validateEmail(email); err != nil {
		return SignUpResult{}, err
	}
	if len(password) < 8 {
		return SignUpResult{}, errors.Join(ErrValidation, errors.New("password must be at least 8 characters"))
	}
	if len(password) > 72 {
		return SignUpResult{}, errors.Join(ErrValidation, errors.New("password must be at most 72 characters"))
	}

	if _, err := s.repository.GetUserByUsername(ctx, username); err == nil {
		return SignUpResult{}, ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return SignUpResult{}, err
	}
	if _, err := s.repository.GetUserByEmail(ctx, email); err == nil {
		return SignUpResult{}, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return SignUpResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return SignUpResult{}, err
	}
	created, err := s.repository.CreateUser(ctx, username, email, string(hash))
	if err != nil {
		return SignUpResult{}, mapTaken(err)
	}

	code, err := s.issueChallenge(ctx, created.ID)
	if err != nil {
		return SignUpResult{}, err
	}

	result := SignUpResult{
		User:             publicUser(created),
		VerificationSent: true,
	}
	if s.devExposeCodes {
		result.DevCode = code
	}
	return result, nil
}

// VerifyEmail consumes a verification challenge and, on success, signs the
// User in with its first token pair.
func (s *AuthService) VerifyEmail(ctx context.Context, email, code, deviceLabel string) (AuthResult, error) {
	email = strings.TrimSpace(email)
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrInvalidChallenge
		}
		return AuthResult{}, err
	}
	if user.EmailVerified {
		return AuthResult{}, ErrAlreadyVerified
	}

	challenge, err := s.repository.GetLatestVerificationChallenge(ctx, user.ID, ChallengePurposeVerifyEmail)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrInvalidChallenge
		}
		return AuthResult{}, err
	}

	now := s.currentTime()
	if now.After(challenge.ExpiresAt) {
		_ = s.repository.ConsumeVerificationChallenge(ctx, challenge.ID)
		return AuthResult{}, ErrInvalidChallenge
	}
	if challenge.Attempts >= VerificationMaxAttempts {
		_ = s.repository.ConsumeVerificationChallenge(ctx, challenge.ID)
		return AuthResult{}, ErrChallengeLocked
	}
	if auth.HashVerificationCode(strings.TrimSpace(code)) != challenge.CodeHash {
		_ = s.repository.IncrementChallengeAttempts(ctx, challenge.ID)
		if challenge.Attempts+1 >= VerificationMaxAttempts {
			_ = s.repository.ConsumeVerificationChallenge(ctx, challenge.ID)
			return AuthResult{}, ErrChallengeLocked
		}
		return AuthResult{}, ErrInvalidChallenge
	}

	if err := s.repository.ConsumeVerificationChallenge(ctx, challenge.ID); err != nil {
		return AuthResult{}, err
	}
	if err := s.repository.MarkUserVerified(ctx, user.ID); err != nil {
		return AuthResult{}, err
	}
	user.EmailVerified = true
	return s.issueSession(ctx, user, deviceLabel)
}

// ResendVerification invalidates the pending challenge and issues a fresh one.
func (s *AuthService) ResendVerification(ctx context.Context, email string) (SignUpResult, error) {
	email = strings.TrimSpace(email)
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return SignUpResult{}, ErrNoPendingVerify
		}
		return SignUpResult{}, err
	}
	if user.EmailVerified {
		return SignUpResult{}, ErrAlreadyVerified
	}

	if pending, err := s.repository.GetLatestVerificationChallenge(ctx, user.ID, ChallengePurposeVerifyEmail); err == nil {
		if s.currentTime().Sub(pending.CreatedAt) < VerificationResendCooldown {
			return SignUpResult{}, ErrResendTooSoon
		}
		_ = s.repository.ConsumeVerificationChallenge(ctx, pending.ID)
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return SignUpResult{}, err
	}

	code, err := s.issueChallenge(ctx, user.ID)
	if err != nil {
		return SignUpResult{}, err
	}
	result := SignUpResult{User: publicUser(user), VerificationSent: true}
	if s.devExposeCodes {
		result.DevCode = code
	}
	return result, nil
}

// SignIn signs a verified User in with Password sign-in. Unknown addresses,
// missing passwords, and wrong passwords all return the same generic error so
// Users cannot be enumerated.
func (s *AuthService) SignIn(ctx context.Context, email, password, deviceLabel string) (AuthResult, error) {
	user, err := s.repository.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrInvalidCredential
		}
		return AuthResult{}, err
	}
	if !user.HasPassword || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredential
	}
	if !user.EmailVerified {
		return AuthResult{}, ErrEmailNotVerified
	}
	return s.issueSession(ctx, user, deviceLabel)
}

func (s *AuthService) issueChallenge(ctx context.Context, userID string) (string, error) {
	code, hash, err := auth.NewVerificationCode()
	if err != nil {
		return "", err
	}
	_, err = s.repository.CreateVerificationChallenge(ctx, userID, ChallengePurposeVerifyEmail, hash, s.currentTime().Add(VerificationCodeTTL))
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *AuthService) issueSession(ctx context.Context, user repository.AuthUser, deviceLabel string) (AuthResult, error) {
	now := s.currentTime()
	access, err := auth.SignAccessToken(s.jwtSecret, user.ID, user.Username, AccessTokenTTL, now)
	if err != nil {
		return AuthResult{}, err
	}
	refresh, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}
	if _, err := s.repository.CreateSession(ctx, user.ID, refreshHash, now.Add(RefreshTokenTTL), deviceLabel); err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User:         publicUser(user),
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
	}, nil
}

func publicUser(user repository.AuthUser) PublicUser {
	return PublicUser{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Verified: user.EmailVerified,
	}
}

func validateEmail(email string) error {
	if len(email) == 0 || len(email) > 254 {
		return errors.Join(ErrValidation, errors.New("email must be 1-254 characters"))
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.Join(ErrValidation, errors.New("email must be a valid address"))
	}
	return nil
}

// mapTaken converts a storage uniqueness violation into the taken error the
// caller can name. The pre-checks above name the field; this is the race guard.
func mapTaken(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "users_username_unique") || strings.Contains(msg, "username") && strings.Contains(msg, "duplicate") {
		return ErrUsernameTaken
	}
	if strings.Contains(msg, "users_email_unique") || strings.Contains(msg, "email") && strings.Contains(msg, "duplicate") {
		return ErrEmailTaken
	}
	return err
}
