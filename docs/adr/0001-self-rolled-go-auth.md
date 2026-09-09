# Self-rolled Go auth with server-verified social sign-in

We need username + email + password plus Apple/Google social sign-in with a forced post-social username pick. We decided to build auth in the existing Go backend (Gin, Postgres, sqlc, Goose) and verify Apple/Google ID tokens server-side, because it keeps the custom username-gate and linking rules in one place and avoids managed-provider lock-in and cost.
