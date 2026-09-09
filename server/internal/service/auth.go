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
	// PendingTokenTTL bounds the Username-setup grant.
	PendingTokenTTL = 30 * time.Minute
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
const ChallengePurposeResetPassword = "reset_password"

var (
	ErrValidation          = errors.New("validation failed")
	ErrUsernameTaken       = errors.New("username is already taken")
	ErrEmailTaken          = errors.New("email is already in use")
	ErrInvalidSignIn       = errors.New("invalid email or password")
	ErrInvalidSession      = errors.New("invalid or expired session")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrInvalidChallenge    = errors.New("invalid or expired code")
	ErrChallengeLocked     = errors.New("too many attempts, request a new code")
	ErrNoPendingVerify     = errors.New("no pending verification for this email")
	ErrAlreadyVerified     = errors.New("email is already verified")
	ErrResendTooSoon       = errors.New("verification code was just sent")
	ErrResendLimit         = errors.New("too many codes requested, try again later")
	ErrUserNotFound        = errors.New("user not found")
	ErrSessionRevoked      = errors.New("session was revoked, sign in again")
	ErrSocialMisconfigured = errors.New("social sign-in is not configured")
	ErrSocialNoEmail       = errors.New("no Email came with this Social sign-in")
	ErrSocialUnverified    = errors.New("this Email is not verified by the provider")
	ErrSocialConflict      = errors.New("a User with this Email already exists, sign in with your password first")
	ErrSocialTokenInvalid  = errors.New("invalid social token")
	ErrSocialUnavailable   = errors.New("could not verify social token")
	ErrUsernameAlreadySet  = errors.New("username is already set")
	ErrPasswordAlreadySet  = errors.New("password is already set, use reset instead")
)

// ValidationError carries the field-level detail behind ErrValidation so the
// transport can report it without parsing error strings.
type ValidationError struct {
	Detail string
}

func (e ValidationError) Error() string { return "validation failed: " + e.Detail }

func (e ValidationError) Is(target error) bool { return target == ErrValidation }

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.]{3,20}$`)

func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return ValidationError{Detail: "username must be 3-20 characters of letters, numbers, underscore, or dot"}
	}
	return nil
}

var (
	passwordUpperPattern   = regexp.MustCompile(`[A-Z]`)
	passwordLowerPattern   = regexp.MustCompile(`[a-z]`)
	passwordDigitPattern   = regexp.MustCompile(`[0-9]`)
	passwordSpecialPattern = regexp.MustCompile(`[^A-Za-z0-9]`)
)

// PublicUser is the outward view of a User: handle, contact, verification
// state, and whether Password sign-in is available.
type PublicUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Verified    bool   `json:"verified"`
	HasPassword bool   `json:"has_password"`
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
// signing in, and Social sign-in.
type AuthService struct {
	repository      repository.AuthRepository
	jwtSecret       string
	devExposeCodes  bool
	appleAudience   string
	googleAudiences []string
	now             func() time.Time
}

// NewAuthService wires the service from plain values so the composition
// point owns configuration. now is nil in production (uses time.Now) and
// overridden in tests.
func NewAuthService(repo repository.AuthRepository, jwtSecret string, devExposeCodes bool, appleAudience string, googleAudiences []string) *AuthService {
	return &AuthService{
		repository:      repo,
		jwtSecret:       jwtSecret,
		devExposeCodes:  devExposeCodes,
		appleAudience:   appleAudience,
		googleAudiences: googleAudiences,
		now:             nil,
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
	if err := validateUsername(username); err != nil {
		return SignUpResult{}, err
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

	if err := s.checkChallenge(ctx, user.ID, ChallengePurposeVerifyEmail, code); err != nil {
		return AuthResult{}, err
	}

	if err := s.repository.MarkUserVerified(ctx, user.ID); err != nil {
		return AuthResult{}, err
	}
	user.EmailVerified = true
	result, _, err := s.issueSession(ctx, user, deviceLabel)
	return result, err
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
	result, _, err := s.issueSession(ctx, user, deviceLabel)
	return result, err
}

// Refresh rotates a refresh grant: the presenting token is single-use, and a
// fresh pair is issued. Presenting an already-rotated or revoked token signals
// theft, so every session of the User is revoked instead.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, deviceLabel string) (AuthResult, error) {
	session, err := s.repository.GetSessionByRefreshHash(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrInvalidSession
		}
		return AuthResult{}, err
	}
	if session.Revoked || session.ReplacedBy != "" {
		_ = s.repository.RevokeAllUserSessions(ctx, session.UserID)
		return AuthResult{}, ErrSessionRevoked
	}
	if !s.currentTime().Before(session.ExpiresAt) {
		_ = s.repository.RevokeSession(ctx, session.ID)
		return AuthResult{}, ErrInvalidSession
	}

	user, err := s.repository.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrUserNotFound
		}
		return AuthResult{}, err
	}
	result, newSessionID, err := s.issueSession(ctx, user, deviceLabel)
	if err != nil {
		return AuthResult{}, err
	}
	if err := s.repository.ReplaceSession(ctx, session.ID, newSessionID); err != nil {
		return AuthResult{}, err
	}
	return result, nil
}

// SignOut revokes the presenting session only. Unknown tokens succeed
// idempotently: there is nothing left to cut off.
func (s *AuthService) SignOut(ctx context.Context, refreshToken string) error {
	session, err := s.repository.GetSessionByRefreshHash(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return nil
		}
		return err
	}
	return s.repository.RevokeSession(ctx, session.ID)
}

// SignOutAll revokes every session of the User behind the access token.
func (s *AuthService) SignOutAll(ctx context.Context, accessToken string) error {
	user, err := s.Profile(ctx, accessToken)
	if err != nil {
		return err
	}
	return s.repository.RevokeAllUserSessions(ctx, user.ID)
}

// SocialResult answers a Social sign-in: either a full session, or a pending
// grant that authorizes only picking the Username.
type SocialResult struct {
	User         PublicUser `json:"user"`
	Pending      bool       `json:"pending"`
	PendingToken string     `json:"pending_token,omitempty"`
	Session      AuthResult `json:"session,omitempty"`
}

// SocialSignIn verifies a provider identity token server-side and signs the
// User in: known provider subjects sign straight in, verified social Emails
// matching an existing User link to it, and first-time joins receive a
// pending grant for the Username picker.
func (s *AuthService) SocialSignIn(ctx context.Context, provider, idToken, nonce, deviceLabel string) (SocialResult, error) {
	identity, err := s.verifySocialToken(ctx, provider, idToken, nonce)
	if err != nil {
		return SocialResult{}, err
	}

	// Known provider subjects sign straight in, even when the provider
	// omits the Email on repeat visits (Apple does after first consent).
	if link, err := s.repository.GetIdentity(ctx, identity.Provider, identity.Subject); err == nil {
		return s.socialSession(ctx, link.UserID, deviceLabel)
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return SocialResult{}, err
	}

	if identity.Email == "" {
		return SocialResult{}, ErrSocialNoEmail
	}
	if !identity.EmailVerified {
		if _, err := s.repository.GetUserByEmail(ctx, identity.Email); err == nil {
			return SocialResult{}, ErrSocialConflict
		} else if !errors.Is(err, repository.ErrAuthNotFound) {
			return SocialResult{}, err
		}
		return SocialResult{}, ErrSocialUnverified
	}

	if user, err := s.repository.GetUserByEmail(ctx, identity.Email); err == nil {
		// Link only when the existing User already proved the Email too:
		// otherwise a provider-verified address could take over an
		// unverified Password sign-in account.
		if !user.EmailVerified {
			return SocialResult{}, ErrSocialConflict
		}
		if _, err := s.repository.CreateIdentity(ctx, user.ID, identity.Provider, identity.Subject, identity.Email); err != nil {
			return SocialResult{}, err
		}
		return s.socialSession(ctx, user.ID, deviceLabel)
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return SocialResult{}, err
	}

	created, err := s.repository.CreateUser(ctx, "", identity.Email, "")
	if err != nil {
		return SocialResult{}, mapTaken(err)
	}
	if !created.EmailVerified {
		// Provider-vouched addresses arrive verified.
		if err := s.repository.MarkUserVerified(ctx, created.ID); err != nil {
			return SocialResult{}, err
		}
		created.EmailVerified = true
	}
	if _, err := s.repository.CreateIdentity(ctx, created.ID, identity.Provider, identity.Subject, identity.Email); err != nil {
		return SocialResult{}, err
	}
	pending, err := auth.SignPendingToken(s.jwtSecret, created.ID, PendingTokenTTL, s.currentTime())
	if err != nil {
		return SocialResult{}, err
	}
	return SocialResult{
		User:         publicUser(created),
		Pending:      true,
		PendingToken: pending,
	}, nil
}

// SetUsername picks the Username for a social join holding a pending grant,
// then signs the User in with a full session. The grant is one-shot in
// effect: a User that already has a Username cannot pick again with it.
func (s *AuthService) SetUsername(ctx context.Context, pendingToken, username string) (AuthResult, error) {
	userID, _, scope, err := auth.VerifyAccessToken(s.jwtSecret, pendingToken, s.currentTime())
	if err != nil || scope != auth.ScopeUsernameSetup {
		return AuthResult{}, ErrInvalidSession
	}
	if err := validateUsername(username); err != nil {
		return AuthResult{}, err
	}
	username = strings.TrimSpace(username)
	holder, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return AuthResult{}, ErrUserNotFound
		}
		return AuthResult{}, err
	}
	if holder.Username != "" {
		return AuthResult{}, ErrUsernameAlreadySet
	}
	updated, err := s.assignUsername(ctx, userID, username)
	if err != nil {
		return AuthResult{}, err
	}
	result, _, err := s.issueSession(ctx, updated, "")
	return result, err
}

func (s *AuthService) socialSession(ctx context.Context, userID, deviceLabel string) (SocialResult, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return SocialResult{}, ErrUserNotFound
		}
		return SocialResult{}, err
	}
	if user.Username == "" {
		pending, err := auth.SignPendingToken(s.jwtSecret, user.ID, PendingTokenTTL, s.currentTime())
		if err != nil {
			return SocialResult{}, err
		}
		return SocialResult{User: publicUser(user), Pending: true, PendingToken: pending}, nil
	}
	session, _, err := s.issueSession(ctx, user, deviceLabel)
	if err != nil {
		return SocialResult{}, err
	}
	return SocialResult{User: publicUser(user), Session: session}, nil
}

func (s *AuthService) verifySocialToken(ctx context.Context, provider, idToken, nonce string) (auth.SocialIdentity, error) {
	now := s.currentTime()
	var identity auth.SocialIdentity
	var err error
	switch provider {
	case auth.ProviderApple:
		if s.appleAudience == "" {
			return auth.SocialIdentity{}, ErrSocialMisconfigured
		}
		identity, err = auth.VerifyAppleIDToken(ctx, now, idToken, s.appleAudience, nonce)
	case auth.ProviderGoogle:
		if len(s.googleAudiences) == 0 {
			return auth.SocialIdentity{}, ErrSocialMisconfigured
		}
		identity, err = auth.VerifyGoogleIDToken(ctx, now, idToken, s.googleAudiences)
	default:
		return auth.SocialIdentity{}, ValidationError{Detail: "provider must be apple or google"}
	}
	if err != nil {
		// Keep provider verification failures behind service sentinels so
		// the transport never depends on the crypto layer.
		if errors.Is(err, auth.ErrSocialUnavailable) {
			return auth.SocialIdentity{}, ErrSocialUnavailable
		}
		return auth.SocialIdentity{}, ErrSocialTokenInvalid
	}
	return identity, nil
}

// ForgotPassword issues a reset challenge for a verified Email. Unknown,
// unverified, and rate-limited addresses all get the same neutral answer
// without a challenge, so accounts cannot be enumerated through this
// endpoint (at the cost of hiding "try again later" from legitimate users).
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repository.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil || !user.EmailVerified {
		if err != nil && !errors.Is(err, repository.ErrAuthNotFound) {
			return err
		}
		return nil
	}

	if pending, err := s.repository.GetLatestVerificationChallenge(ctx, user.ID, ChallengePurposeResetPassword); err == nil {
		if s.currentTime().Sub(pending.CreatedAt) < VerificationResendCooldown {
			return nil
		}
		_ = s.repository.ConsumeVerificationChallenge(ctx, pending.ID)
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return err
	}

	recent, err := s.repository.CountRecentVerificationChallenges(ctx, user.ID, ChallengePurposeResetPassword, s.currentTime().Add(-time.Hour))
	if err != nil {
		return err
	}
	if recent >= VerificationMaxChallengesPerHour {
		return nil
	}

	_, err = s.issuePurposeChallenge(ctx, user.ID, ChallengePurposeResetPassword)
	return err
}

// ResetPassword consumes a reset challenge, replaces the stored secret under
// the shared password policy, and revokes every pre-reset session.
func (s *AuthService) ResetPassword(ctx context.Context, email, code, password string) error {
	user, err := s.repository.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return ErrInvalidChallenge
		}
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	if err := s.checkChallenge(ctx, user.ID, ChallengePurposeResetPassword, code); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repository.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	return s.repository.RevokeAllUserSessions(ctx, user.ID)
}

// AddPassword sets the first password for a signed-in User that has none
// (typically a Social sign-in join), under the shared password policy. From
// then on both methods reach the same User.
func (s *AuthService) AddPassword(ctx context.Context, accessToken, password string) error {
	user, err := s.Profile(ctx, accessToken)
	if err != nil {
		return err
	}
	full, err := s.repository.GetUserByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if full.HasPassword {
		return ErrPasswordAlreadySet
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repository.UpdatePassword(ctx, user.ID, string(hash))
}

// ChangeUsername renames a signed-in User under the same format and
// uniqueness rules as sign-up.
func (s *AuthService) ChangeUsername(ctx context.Context, accessToken, username string) (PublicUser, error) {
	user, err := s.Profile(ctx, accessToken)
	if err != nil {
		return PublicUser{}, err
	}
	updated, err := s.assignUsername(ctx, user.ID, username)
	if err != nil {
		return PublicUser{}, err
	}
	return publicUser(updated), nil
}

// assignUsername validates, availability-checks, and commits a Username.
func (s *AuthService) assignUsername(ctx context.Context, userID, username string) (repository.AuthUser, error) {
	if err := validateUsername(username); err != nil {
		return repository.AuthUser{}, err
	}
	username = strings.TrimSpace(username)
	if existing, err := s.repository.GetUserByUsername(ctx, username); err == nil {
		if existing.ID != userID {
			return repository.AuthUser{}, ErrUsernameTaken
		}
		return existing, nil
	} else if !errors.Is(err, repository.ErrAuthNotFound) {
		return repository.AuthUser{}, err
	}
	updated, err := s.repository.UpdateUsername(ctx, userID, username)
	if err != nil {
		return repository.AuthUser{}, mapTaken(err)
	}
	return updated, nil
}

// issuePurposeChallenge creates a challenge like issueChallenge but for an
// explicit purpose, returning the raw code for the delivery channel.
func (s *AuthService) issuePurposeChallenge(ctx context.Context, userID, purpose string) (string, error) {
	code, hash, err := auth.NewVerificationCode()
	if err != nil {
		return "", err
	}
	_, err = s.repository.CreateVerificationChallenge(ctx, userID, purpose, hash, s.currentTime().Add(VerificationCodeTTL))
	if err != nil {
		return "", err
	}
	return code, nil
}

// Profile reads the current User from a presented access token. Pending
// grants are rejected: they authorize only setting the Username.
func (s *AuthService) Profile(ctx context.Context, accessToken string) (PublicUser, error) {
	userID, _, scope, err := auth.VerifyAccessToken(s.jwtSecret, accessToken, s.currentTime())
	if err != nil {
		return PublicUser{}, ErrInvalidSession
	}
	if scope != auth.ScopeFull {
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

// checkChallenge validates a code against the latest challenge for a
// purpose, consuming it on success, expiry, or lock. Wrong guesses count
// toward the lock; locked challenges stay until expiry or replacement.
func (s *AuthService) checkChallenge(ctx context.Context, userID, purpose, code string) error {
	challenge, err := s.repository.GetLatestVerificationChallenge(ctx, userID, purpose)
	if err != nil {
		if errors.Is(err, repository.ErrAuthNotFound) {
			return ErrInvalidChallenge
		}
		return err
	}

	if s.currentTime().After(challenge.ExpiresAt) {
		_ = s.repository.ConsumeVerificationChallenge(ctx, challenge.ID)
		return ErrInvalidChallenge
	}
	if challenge.Attempts >= VerificationMaxAttempts {
		return ErrChallengeLocked
	}
	if auth.HashVerificationCode(strings.TrimSpace(code)) != challenge.CodeHash {
		_ = s.repository.IncrementChallengeAttempts(ctx, challenge.ID)
		if challenge.Attempts+1 >= VerificationMaxAttempts {
			return ErrChallengeLocked
		}
		return ErrInvalidChallenge
	}

	return s.repository.ConsumeVerificationChallenge(ctx, challenge.ID)
}

func (s *AuthService) issueChallenge(ctx context.Context, userID string) (string, error) {
	return s.issuePurposeChallenge(ctx, userID, ChallengePurposeVerifyEmail)
}

func (s *AuthService) issueSession(ctx context.Context, user repository.AuthUser, deviceLabel string) (AuthResult, string, error) {
	now := s.currentTime()
	access, err := auth.SignAccessToken(s.jwtSecret, user.ID, user.Username, auth.ScopeFull, AccessTokenTTL, now)
	if err != nil {
		return AuthResult{}, "", err
	}
	refresh, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		return AuthResult{}, "", err
	}
	sessionID, err := s.repository.CreateSession(ctx, user.ID, refreshHash, now.Add(RefreshTokenTTL), deviceLabel)
	if err != nil {
		return AuthResult{}, "", err
	}
	return AuthResult{
		User:         publicUser(user),
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
	}, sessionID, nil
}

func publicUser(user repository.AuthUser) PublicUser {
	return PublicUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Verified:    user.EmailVerified,
		HasPassword: user.HasPassword,
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
