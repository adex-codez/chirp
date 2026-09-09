# Self-rolled Go auth with server-verified social sign-in

We need username + email + password plus Apple/Google social sign-in with a forced post-social username pick. We decided to build auth in the existing Go backend (Gin, Postgres, sqlc, Goose) and verify Apple/Google ID tokens server-side, because it keeps the custom username-gate and linking rules in one place and avoids managed-provider lock-in and cost.

## Password policy

We need a single password rule shared by sign-up, reset, and add-password so tickets and code never drift. We decided passwords are 8-15 characters and must include an uppercase letter, a lowercase letter, a number, and a special character (any non-alphanumeric), because a short bounded length plus all four character classes is easy to state in tickets, easy to hint in mobile UI, and strong enough for this app without bcrypt-truncation surprises.

Enforced server-side in `AuthService.SignUp` (`server/internal/service/auth.go`) via `validatePassword`, returning field-specific `ValidationError`s (`"password must be 8-15 characters"` / `"password must include an uppercase letter, a lowercase letter, a number, and a special character"`). Reset and add-password flows enforce the identical rule. Go `regexp` (RE2) has no lookahead, so the implementation uses four single-class patterns (`[A-Z]`, `[a-z]`, `[0-9]`, `[^A-Za-z0-9]`) plus the length check instead of one `^(?=.*...)` expression.
