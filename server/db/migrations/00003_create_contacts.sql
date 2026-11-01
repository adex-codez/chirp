-- +goose Up
CREATE TABLE contacts (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, contact_id),
    CHECK (user_id <> contact_id)
);
CREATE INDEX contacts_user_created_idx ON contacts (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS contacts;
