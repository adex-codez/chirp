-- name: AddContact :exec
INSERT INTO contacts (user_id, contact_id)
VALUES (sqlc.arg(user_id)::uuid, sqlc.arg(contact_id)::uuid)
ON CONFLICT (user_id, contact_id) DO NOTHING;

-- name: GetContactableUserByUsername :one
SELECT id::text AS id, COALESCE(username, '') AS username
FROM users
WHERE lower(username) = lower(sqlc.arg(username)) AND email_verified_at IS NOT NULL
LIMIT 1;

-- name: ListContacts :many
SELECT users.id::text AS id, users.username, users.profile_picture_url
FROM contacts
JOIN users ON users.id = contacts.contact_id
WHERE contacts.user_id = sqlc.arg(user_id)::uuid
ORDER BY contacts.created_at DESC, users.username ASC;
