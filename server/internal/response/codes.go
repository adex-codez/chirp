package response

// The machine-readable failure codes the API returns. Clients branch on these,
// so the values are part of the wire contract: renaming one is a breaking
// change, and adding one obliges the client to handle it.
//
// This file is the only place a code is written down. `server/cmd/apicodes`
// generates `chirp-mobile/src/lib/api-codes.ts` from it, so the two cannot
// drift. Every exported string constant whose name starts with Code is a code.
const (
	CodeValidationFailed    = "validation_failed"
	CodeUsernameTaken       = "username_taken"
	CodeEmailInUse          = "email_in_use"
	CodeInvalidCredentials  = "invalid_credentials"
	CodeInvalidSession      = "invalid_session"
	CodeSessionRevoked      = "session_revoked"
	CodeEmailNotVerified    = "email_not_verified"
	CodeInvalidCode         = "invalid_code"
	CodeChallengeLocked     = "challenge_locked"
	CodeNoPendingVerify     = "no_pending_verification"
	CodeAlreadyVerified     = "already_verified"
	CodeResendTooSoon       = "resend_too_soon"
	CodeResendLimit         = "resend_limit"
	CodeUserNotFound        = "user_not_found"
	CodeSocialMisconfigured = "social_not_configured"
	CodeSocialNoEmail       = "provider_no_email"
	CodeSocialUnverified    = "provider_email_unverified"
	CodeSocialConflict      = "social_conflict"
	CodeSocialTokenInvalid  = "invalid_provider_token"
	CodeSocialUnavailable   = "provider_unavailable"
	CodeUsernameAlreadySet  = "username_already_set"
	CodePasswordAlreadySet  = "password_already_set"
	CodeServiceUnavailable  = "service_unavailable"
	CodeInternalError       = "internal_error"
)
