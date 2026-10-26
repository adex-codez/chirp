/**
 * User-facing wording for API failures. The server's own `error` field is
 * never shown.
 */

import type { ApiCode } from "./api-codes";

export const CODE_MESSAGES: Record<ApiCode, string> = {
  validation_failed: "Some of the details are not right. Check them and try again.",
  username_taken: "That username is already taken.",
  email_in_use: "An account already uses that email. Sign in instead.",
  invalid_credentials: "Email or password is incorrect.",
  invalid_session: "Your session has ended. Sign in again.",
  session_revoked: "Your session was ended somewhere else. Sign in again.",
  email_not_verified: "Confirm your email address first. Check your inbox for the code.",
  invalid_code: "That code is invalid or has expired.",
  challenge_locked: "Too many attempts. Request a new code.",
  no_pending_verification: "There is no pending verification for this email.",
  already_verified: "That email is already verified.",
  resend_too_soon: "Wait a minute before requesting another code.",
  resend_limit: "Too many codes requested. Try again later.",
  user_not_found: "We could not find that user.",
  social_not_configured: "That sign-in method is not available right now.",
  provider_no_email: "That provider did not share an email address.",
  provider_email_unverified: "That provider has not verified this email address.",
  social_conflict: "An account already uses this email. Sign in with your password first.",
  invalid_provider_token: "That sign-in could not be verified. Try again.",
  provider_unavailable: "We could not verify that sign-in. Try again.",
  username_already_set: "This account already has a username.",
  password_already_set: "This account already has a password. Reset it instead.",
  service_unavailable: "Chirp is temporarily unavailable. Try again.",
  internal_error: "Something went wrong on our end. Try again.",
};

/** Fallbacks for a 4xx whose code we do not recognise. */
export const STATUS_MESSAGES: Record<number, string> = {
  400: "That request was not valid.",
  401: "Sign in again to continue.",
  403: "You do not have permission to do that.",
  404: "We could not find what you were looking for.",
  405: "That action is not available.",
  408: "The request timed out. Try again.",
  409: "That conflicts with something that already exists.",
  413: "That request is too large.",
  415: "That format is not supported.",
  422: "Check the details and try again.",
  429: "Too many requests. Wait before trying again.",
};

/** Fallbacks for a 5xx. The server's own wording is never shown here. */
export const SERVER_STATUS_MESSAGES: Record<number, string> = {
  500: "Something went wrong on our end. Try again.",
  502: "Chirp is temporarily unavailable. Try again.",
  503: "Chirp is temporarily unavailable. Try again.",
  504: "The request took too long. Try again.",
};

export const DEFAULT_ERROR_MESSAGE = "Something went wrong. Try again.";

/**
 * Picks the message for a failed response.
 *
 * A 5xx never uses the server's wording, even when the code is one we know:
 * the server is in a state we cannot reason about from the client.
 */
export function messageFor(status: number, code?: ApiCode): string {
  if (status >= 500) {
    return SERVER_STATUS_MESSAGES[status] ?? DEFAULT_ERROR_MESSAGE;
  }
  if (code) {
    return CODE_MESSAGES[code];
  }
  return STATUS_MESSAGES[status] ?? DEFAULT_ERROR_MESSAGE;
}
