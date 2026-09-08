-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id::text AS id, username, email, password_hash, email_verified_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id::text AS id, username, email, password_hash, email_verified_at, created_at, updated_at
FROM users
WHERE lower(email) = lower($1)
LIMIT 1;

-- name: GetUserByUsername :one
SELECT id::text AS id, username, email, password_hash, email_verified_at, created_at, updated_at
FROM users
WHERE lower(username) = lower($1)
LIMIT 1;

-- name: GetUserByID :one
SELECT id::text AS id, username, email, password_hash, email_verified_at, created_at, updated_at
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

-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_hash, expires_at, device_label)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text AS id, user_id::text AS user_id, refresh_hash, expires_at, revoked_at, created_at;
