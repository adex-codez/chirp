-- +goose Up
ALTER TABLE users ALTER COLUMN username DROP NOT NULL;
DROP INDEX users_username_unique;
CREATE UNIQUE INDEX users_username_unique ON users (lower(username)) WHERE username IS NOT NULL;

CREATE TABLE linked_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_sub TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT linked_identities_provider_check CHECK (provider IN ('apple', 'google'))
);
CREATE UNIQUE INDEX linked_identities_provider_sub_unique ON linked_identities (provider, provider_sub);
CREATE UNIQUE INDEX linked_identities_user_provider_unique ON linked_identities (user_id, provider);
CREATE INDEX linked_identities_user_idx ON linked_identities (user_id);

-- +goose Down
DROP TABLE IF EXISTS linked_identities;
DELETE FROM users WHERE username IS NULL;
DROP INDEX IF EXISTS users_username_unique;
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
CREATE UNIQUE INDEX users_username_unique ON users (lower(username));
