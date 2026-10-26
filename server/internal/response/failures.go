package response

import (
	"errors"
	"net/http"

	"backend/internal/service"
)

type failure struct {
	status  int
	message string
}

var failures = map[string]failure{
	CodeValidationFailed:    {http.StatusBadRequest, "validation failed"},
	CodeUsernameTaken:       {http.StatusConflict, "username is already taken"},
	CodeEmailInUse:          {http.StatusConflict, "email is already in use"},
	CodeInvalidCredentials:  {http.StatusUnauthorized, "invalid email or password"},
	CodeInvalidSession:      {http.StatusUnauthorized, "invalid or expired session"},
	CodeSessionRevoked:      {http.StatusUnauthorized, "session was revoked, sign in again"},
	CodeEmailNotVerified:    {http.StatusForbidden, "email is not verified, check your inbox for the code"},
	CodeInvalidCode:         {http.StatusBadRequest, "invalid or expired code"},
	CodeChallengeLocked:     {http.StatusTooManyRequests, "too many attempts, request a new code"},
	CodeNoPendingVerify:     {http.StatusNotFound, "no pending verification for this email"},
	CodeAlreadyVerified:     {http.StatusBadRequest, "email is already verified"},
	CodeResendTooSoon:       {http.StatusTooManyRequests, "verification code was just sent, wait a minute"},
	CodeResendLimit:         {http.StatusTooManyRequests, "too many codes requested, try again later"},
	CodeUserNotFound:        {http.StatusNotFound, "user not found"},
	CodeSocialMisconfigured: {http.StatusServiceUnavailable, "social sign-in is not configured"},
	CodeSocialNoEmail:       {http.StatusBadRequest, "no Email came with this Social sign-in"},
	CodeSocialUnverified:    {http.StatusForbidden, "this Email is not verified by the provider"},
	CodeSocialConflict:      {http.StatusConflict, "a User with this Email already exists, sign in with your password first"},
	CodeSocialTokenInvalid:  {http.StatusUnauthorized, "invalid social token"},
	CodeSocialUnavailable:   {http.StatusBadGateway, "could not verify social token"},
	CodeUsernameAlreadySet:  {http.StatusConflict, "username is already set"},
	CodePasswordAlreadySet:  {http.StatusConflict, "password is already set, use reset instead"},
	CodeServiceUnavailable:  {http.StatusServiceUnavailable, "service unavailable"},
	CodeInternalError:       {http.StatusInternalServerError, "something went wrong"},
}

// Problem is a failed request rendered for the client.
type Problem struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
}

var sentinelCodes = []struct {
	err  error
	code string
}{
	{service.ErrUsernameTaken, CodeUsernameTaken},
	{service.ErrEmailTaken, CodeEmailInUse},
	{service.ErrInvalidSignIn, CodeInvalidCredentials},
	{service.ErrInvalidSession, CodeInvalidSession},
	{service.ErrSessionRevoked, CodeSessionRevoked},
	{service.ErrEmailNotVerified, CodeEmailNotVerified},
	{service.ErrInvalidChallenge, CodeInvalidCode},
	{service.ErrChallengeLocked, CodeChallengeLocked},
	{service.ErrNoPendingVerify, CodeNoPendingVerify},
	{service.ErrAlreadyVerified, CodeAlreadyVerified},
	{service.ErrResendTooSoon, CodeResendTooSoon},
	{service.ErrResendLimit, CodeResendLimit},
	{service.ErrUserNotFound, CodeUserNotFound},
	{service.ErrSocialMisconfigured, CodeSocialMisconfigured},
	{service.ErrSocialNoEmail, CodeSocialNoEmail},
	{service.ErrSocialUnverified, CodeSocialUnverified},
	{service.ErrSocialConflict, CodeSocialConflict},
	{service.ErrSocialTokenInvalid, CodeSocialTokenInvalid},
	{service.ErrSocialUnavailable, CodeSocialUnavailable},
	{service.ErrUsernameAlreadySet, CodeUsernameAlreadySet},
	{service.ErrPasswordAlreadySet, CodePasswordAlreadySet},
}

// From maps a domain error to the failure the client sees.
func From(err error) Problem {
	var validation service.ValidationError
	if errors.As(err, &validation) {
		return Problem{Code: CodeValidationFailed, Fields: validation.Fields}
	}

	for _, sentinel := range sentinelCodes {
		if errors.Is(err, sentinel.err) {
			return Problem{Code: sentinel.code}
		}
	}

	return Problem{Code: CodeInternalError}
}

func resolve(problem Problem) Problem {
	entry, known := failures[problem.Code]
	if !known {
		entry = failures[CodeInternalError]
		problem.Code = CodeInternalError
		problem.Fields = nil
	}

	problem.Status = entry.status
	problem.Message = entry.message

	return problem
}
