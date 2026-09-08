# Short JWT access plus rotating opaque refresh

We need 15-minute access with seamless renewal and lockout on theft. We decided on short-lived JWT access tokens plus opaque single-use rotating refresh tokens (30 days, hashed in Postgres, reuse revokes all sessions), because rotation with reuse detection gives theft signal that static long-lived tokens do not, at the cost of server-side session storage.

## Consequences

Mobile keeps both tokens in SecureStore and refreshes proactively; logout revokes one session, logout-all revokes every session.
