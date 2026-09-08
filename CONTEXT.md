# Chirp

A social mobile app where every participant has a verified identity and a unique public handle.

## Language

**User**:
A person with an account who can sign in and use the app.
_Avoid_: Account, profile, customer

**Username**:
The unique public handle of a User, picked by the User.
_Avoid_: Handle, display name, nickname

**Email**:
The unique contact address of a User, used for sign-in and verification.
_Avoid_: Email address, login, contact

**Password sign-in**:
Signing in with an Email and a secret chosen by the User.
_Avoid_: Credentials, email login

**Social sign-in**:
Signing in with an Apple or Google identity instead of a password.
_Avoid_: OAuth, SSO, third-party login

**Verified email**:
An Email confirmed to belong to the User via a verification step or a trusted social provider.
_Avoid_: Confirmed email, validated email
