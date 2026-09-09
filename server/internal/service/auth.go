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
	// VerificationMaxChallengesPerHour caps how many challenges a User can
	// mint, so resend cannot be abused to flood an inbox.
	VerificationMaxChallengesPerHour = 5
)

const ChallengePurposeVerifyEmail = "verify_email"

var (
	ErrValidation       = errors.New("validation failed")
	ErrUsernameTaken    = errors.New("username is already taken")
	ErrEmailTaken       = errors.New("email is already in use")
	ErrInvalidSignIn    = errors.New("invalid email or password")
	ErrInvalidSession   = errors.New("invalid or expired session")
	ErrEmailNotVerified = errors.New("email is not verified")
	ErrInvalidChallenge = errors.New("invalid or expired code")
	ErrChallengeLocked  = errors.New("too many attempts, request a new code")
	ErrNoPendingVerify  = errors.New("no pending verification for this email")
	ErrAlreadyVerified  = errors.New("email is already verified")
	ErrResendTooSoon    = errors.New("verification code was just sent")
	ErrResendLimit      = errors.New("too many codes requested, try again later")
	ErrUserNotFound     = errors.New("user not found")
)

// ValidationError carries the field-level detail behind ErrValidation so the
// transport can report it without parsing error strings.
type ValidationError struct {
	Detail string
}

func (e ValidationError) Error() string { return "validation failed: " + e.Detail }

func (e ValidationError) Is(target error) bool { return target == ErrValidation }

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.]{3,20}$`)

var (
	passwordUpperPattern   = regexp.MustCompile(`[A-Z]`)
	passwordLowerPattern   = regexp.MustCompile(`[a-z]`)
	passwordDigitPattern   = regexp.MustCompile(`[0-9]`)
	passwordSpecialPattern = regexp.MustCompile(`[^A-Za-z0-9]`)
)

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
		return SignUpResult{}, ValidationError{Detail: "username must be 3-20 characters of letters, numbers, underscore, or dot"}
	}
	if err := validateEmail(email); err != nil {
		return SignUpResult{}, err
	}
	if err := validatePassword(password); err != nil {
		return SignUpResult{}, err
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
		// Leave the challenge in place so further guesses keep returning
		// locked until expiry; only a resend clears it.
		return AuthResult{}, ErrChallengeLocked
	}
	if auth.HashVerificationCode(strings.TrimSpace(code)) != challenge.CodeHash {
		_ = s.repository.IncrementChallengeAttempts(ctx, challenge.ID)
		if challenge.Attempts+1 >= VerificationMaxAttempts {
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

	recent, err := s.repository.CountRecentVerificationChallenges(ctx, user.ID, ChallengePurposeVerifyEmail, s.currentTime().Add(-time.Hour))
	if err != nil {
		return SignUpResult{}, err
	}
	if recent >= VerificationMaxChallengesPerHour {
		return SignUpResult{}, ErrResendLimit
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

// SignIn signs a verified User in with Password sign-in. The password is
// checked before verification state so wrong passwords always return the same
// generic error and Users cannot be enumerated; only a correct password on an
// unverified Email reveals that verification is pending.
func (s *AuthService) SignIn(ctx context.Context, email, password, deviceLabel string) (AuthResult, error) {
	user, err := s.repository.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrInvalidSignIn
		}
		return AuthResult{}, err
	}
	if !user.HasPassword || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidSignIn
	}
	if !user.EmailVerified {
		return AuthResult{}, ErrEmailNotVerified
	}
	return s.issueSession(ctx, user, deviceLabel)
}

// Profile reads the current User from a presented access token.
func (s *AuthService) Profile(ctx context.Context, accessToken string) (PublicUser, error) {
	userID, _, err := auth.VerifyAccessToken(s.jwtSecret, accessToken, s.currentTime())
	if err != nil {
		return PublicUser{}, ErrInvalidSession
	}
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return PublicUser{}, ErrUserNotFound
		}
		return PublicUser{}, err
	}
	return publicUser(user), nil
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

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 15 {
		return ValidationError{Detail: "password must be 8-15 characters"}
	}
	if !passwordUpperPattern.MatchString(password) ||
		!passwordLowerPattern.MatchString(password) ||
		!passwordDigitPattern.MatchString(password) ||
		!passwordSpecialPattern.MatchString(password) {
		return ValidationError{Detail: "password must include an uppercase letter, a lowercase letter, a number, and a special character"}
	}
	return nil
}

func validateEmail(email string) error {
	if len(email) == 0 || len(email) > 254 {
		return ValidationError{Detail: "email must be 1-254 characters"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ValidationError{Detail: "email must be a valid address"}
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
