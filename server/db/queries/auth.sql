-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES (NULLIF($1::text, ''), $2, $3)
RETURNING id::text AS id, COALESCE(username, '') AS username, email, password_hash, email_verified_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id::text AS id, COALESCE(username, '') AS username, email, password_hash, email_verified_at, created_at, updated_at
FROM users
WHERE lower(email) = lower($1)
LIMIT 1;

-- name: GetUserByUsername :one
SELECT id::text AS id, COALESCE(username, '') AS username, email, password_hash, email_verified_at, created_at, updated_at
FROM users
WHERE lower(username) = lower($1)
LIMIT 1;

-- name: GetUserByID :one
SELECT id::text AS id, COALESCE(username, '') AS username, email, password_hash, email_verified_at, created_at, updated_at
FROM users
WHERE id = $1::uuid;

-- name: MarkUserVerified :exec
UPDATE users
SET email_verified_at = now(), updated_at = now()
WHERE id = $1::uuid AND email_verified_at IS NULL;

-- name: CreateVerificationChallenge :one
INSERT INTO verification_challenges (user_id, purpose, code_hash, expires_at)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text AS id, user_id::text AS user_id, purpose, code_hash, expires_at, attempts, consumed_at, created_at;

-- name: GetLatestVerificationChallenge :one
SELECT id::text AS id, user_id::text AS user_id, purpose, code_hash, expires_at, attempts, consumed_at, created_at
FROM verification_challenges
WHERE user_id = $1::uuid AND purpose = $2 AND consumed_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: ConsumeVerificationChallenge :exec
UPDATE verification_challenges
SET consumed_at = now()
WHERE id = $1::uuid;

-- name: IncrementChallengeAttempts :exec
UPDATE verification_challenges
SET attempts = attempts + 1
WHERE id = $1::uuid;

-- name: CountRecentVerificationChallenges :one
SELECT COUNT(*)
FROM verification_challenges
WHERE user_id = $1::uuid AND purpose = $2 AND created_at > $3;

-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_hash, expires_at, device_label)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text AS id, user_id::text AS user_id, refresh_hash, expires_at, revoked_at, created_at;

-- name: GetIdentity :one
SELECT id::text AS id, user_id::text AS user_id, provider, provider_sub, email
FROM linked_identities
WHERE provider = $1 AND provider_sub = $2
LIMIT 1;

-- name: CreateIdentity :one
INSERT INTO linked_identities (user_id, provider, provider_sub, email)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text AS id, user_id::text AS user_id, provider, provider_sub, email;

-- name: UpdateUsername :one
UPDATE users
SET username = $2, updated_at = now()
WHERE id = $1::uuid
RETURNING id::text AS id, COALESCE(username, '') AS username, email, password_hash, email_verified_at, created_at, updated_at;

-- name: GetSessionByRefreshHash :one
SELECT id::text AS id, user_id::text AS user_id, refresh_hash, expires_at, revoked_at, COALESCE(replaced_by::text, '') AS replaced_by, created_at
FROM sessions
WHERE refresh_hash = $1
LIMIT 1;

-- name: ReplaceSession :exec
UPDATE sessions
SET replaced_by = $2::uuid, revoked_at = now()
WHERE id = $1::uuid;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = now()
WHERE id = $1::uuid AND revoked_at IS NULL;

-- name: RevokeAllUserSessions :exec
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1::uuid AND revoked_at IS NULL;
