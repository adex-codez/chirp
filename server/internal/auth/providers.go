package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// ProviderApple and ProviderGoogle name the supported Social sign-in
	// providers. They match the linked_identities provider constraint.
	ProviderApple  = "apple"
	ProviderGoogle = "google"

	appleKeysURL  = "https://appleid.apple.com/auth/keys"
	appleIssuer   = "https://appleid.apple.com"
	googleKeysURL = "https://www.googleapis.com/oauth2/v3/certs"
)

var (
	// ErrSocialTokenInvalid marks a presented social token that fails
	// signature, issuer, audience, expiry, or nonce checks.
	ErrSocialTokenInvalid = errors.New("invalid social token")
	// ErrSocialUnavailable marks a verification failure on our side, such as
	// unreachable provider keys.
	ErrSocialUnavailable = errors.New("could not verify social token")
)

// SocialIdentity is the verified identity a provider asserts: who (Subject),
// where to reach them (Email), and whether the provider vouches for the
// address (EmailVerified).
type SocialIdentity struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
}

var socialHTTPClient = &http.Client{Timeout: 10 * time.Second}

// keyCache holds provider signing keys with a TTL, refreshing on expiry and
// once more on an unknown key ID (rotation gap).
type keyCache struct {
	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
	ttl     time.Duration
	url     string
}

func (c *keyCache) keyFor(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.keys == nil || time.Now().After(c.expires) {
		if err := c.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	if err := c.refreshLocked(ctx); err != nil {
		return nil, err
	}
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("unknown signing key: %w", ErrSocialTokenInvalid)
}

func (c *keyCache) refreshLocked(ctx context.Context) error {
	keys, err := fetchRSAKeys(ctx, c.url)
	if err != nil {
		return err
	}
	c.keys = keys
	c.expires = time.Now().Add(c.ttl)
	return nil
}

var (
	appleKeyCache  = &keyCache{ttl: time.Hour, url: appleKeysURL}
	googleKeyCache = &keyCache{ttl: time.Hour, url: googleKeysURL}
)

// VerifyAppleIDToken verifies a Sign in with Apple identity token: RS256
// signature against Apple's keys, issuer, audience, expiry, and — when nonce
// is non-empty — the nonce claim against the sha256 of the raw nonce.
func VerifyAppleIDToken(ctx context.Context, now time.Time, token, audience, nonce string) (SocialIdentity, error) {
	claims, err := verifyRS256(ctx, appleKeyCache, token, now)
	if err != nil {
		return SocialIdentity{}, err
	}
	if claims.Issuer != appleIssuer {
		return SocialIdentity{}, fmt.Errorf("unexpected issuer %q: %w", claims.Issuer, ErrSocialTokenInvalid)
	}
	if audience == "" || !claims.Audiences.contains(audience) {
		return SocialIdentity{}, fmt.Errorf("unexpected audience: %w", ErrSocialTokenInvalid)
	}
	if nonce != "" && claims.Nonce != sha256Hex(nonce) {
		return SocialIdentity{}, fmt.Errorf("nonce mismatch: %w", ErrSocialTokenInvalid)
	}
	// Apple only asserts addresses it vouches for, including relay ones. An
	// explicitly negative claim is honored; an absent one means vouched.
	verified := claims.Email != "" &&
		(claims.EmailVerified == nil || bool(*claims.EmailVerified))
	return SocialIdentity{
		Provider:      ProviderApple,
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: verified,
	}, nil
}

// VerifyGoogleIDToken verifies a Google ID token: RS256 signature against
// Google's keys, issuer, audience membership, and expiry.
func VerifyGoogleIDToken(ctx context.Context, now time.Time, token string, audiences []string) (SocialIdentity, error) {
	claims, err := verifyRS256(ctx, googleKeyCache, token, now)
	if err != nil {
		return SocialIdentity{}, err
	}
	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return SocialIdentity{}, fmt.Errorf("unexpected issuer %q: %w", claims.Issuer, ErrSocialTokenInvalid)
	}
	matched := false
	for _, audience := range audiences {
		if audience != "" && claims.Audiences.contains(audience) {
			matched = true
			break
		}
	}
	if !matched {
		return SocialIdentity{}, fmt.Errorf("unexpected audience: %w", ErrSocialTokenInvalid)
	}
	return SocialIdentity{
		Provider:      ProviderGoogle,
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: claims.Email != "" && claims.EmailVerified != nil && bool(*claims.EmailVerified),
	}, nil
}

type tokenClaims struct {
	Subject       string    `json:"sub"`
	Email         string    `json:"email"`
	EmailVerified *flexBool `json:"email_verified"`
	Issuer        string    `json:"iss"`
	Audiences     audience  `json:"aud"`
	Expiry        int64     `json:"exp"`
	Nonce         string    `json:"nonce"`
}

// flexBool accepts the claim as a boolean or a "true"/"false" string: Apple
// encodes email_verified as a string, Google as a boolean.
type flexBool bool

func (b *flexBool) UnmarshalJSON(raw []byte) error {
	var value bool
	if err := json.Unmarshal(raw, &value); err == nil {
		*b = flexBool(value)
		return nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return err
	}
	*b = flexBool(text == "true" || text == "1")
	return nil
}

// audience accepts the claim as either a string or a string array.
type audience []string

func (a *audience) UnmarshalJSON(raw []byte) error {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		*a = []string{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a audience) contains(value string) bool {
	for _, item := range a {
		if item == value {
			return true
		}
	}
	return false
}

func verifyRS256(ctx context.Context, cache *keyCache, token string, now time.Time) (tokenClaims, error) {
	header, claims, signingInput, signature, err := splitToken(token)
	if err != nil {
		return tokenClaims{}, err
	}
	if header.Algorithm != "RS256" {
		return tokenClaims{}, fmt.Errorf("unexpected algorithm %q: %w", header.Algorithm, ErrSocialTokenInvalid)
	}
	key, err := cache.keyFor(ctx, header.KeyID)
	if err != nil {
		return tokenClaims{}, err
	}
	digest := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return tokenClaims{}, fmt.Errorf("bad signature: %w", ErrSocialTokenInvalid)
	}
	if claims.Subject == "" {
		return tokenClaims{}, fmt.Errorf("missing subject: %w", ErrSocialTokenInvalid)
	}
	if now.Unix() >= claims.Expiry {
		return tokenClaims{}, fmt.Errorf("expired token: %w", ErrSocialTokenInvalid)
	}
	return claims, nil
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

func splitToken(token string) (tokenHeader, tokenClaims, string, []byte, error) {
	var header tokenHeader
	var claims tokenClaims
	parts := splitDots(token)
	if len(parts) != 3 {
		return header, claims, "", nil, fmt.Errorf("malformed token: %w", ErrSocialTokenInvalid)
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return header, claims, "", nil, fmt.Errorf("bad header: %w", ErrSocialTokenInvalid)
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return header, claims, "", nil, fmt.Errorf("bad header: %w", ErrSocialTokenInvalid)
	}
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return header, claims, "", nil, fmt.Errorf("bad claims: %w", ErrSocialTokenInvalid)
	}
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return header, claims, "", nil, fmt.Errorf("bad claims: %w", ErrSocialTokenInvalid)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return header, claims, "", nil, fmt.Errorf("bad signature: %w", ErrSocialTokenInvalid)
	}
	return header, claims, parts[0] + "." + parts[1], signature, nil
}

type jwksDocument struct {
	Keys []struct {
		KeyID   string `json:"kid"`
		KeyType string `json:"kty"`
		N       string `json:"n"`
		E       string `json:"e"`
	} `json:"keys"`
}

func fetchRSAKeys(ctx context.Context, url string) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create signing-key request for %s: %w: %w", url, ErrSocialUnavailable, err)
	}
	response, err := socialHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch signing keys from %s: %w: %w", url, ErrSocialUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch signing keys from %s: HTTP %d: %w", url, response.StatusCode, ErrSocialUnavailable)
	}
	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode signing keys from %s: %w: %w", url, ErrSocialUnavailable, err)
	}
	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, entry := range document.Keys {
		if entry.KeyType != "RSA" || entry.KeyID == "" {
			continue
		}
		modulus, err := base64URLToInt(entry.N)
		if err != nil || modulus.Sign() <= 0 {
			continue
		}
		exponent, err := base64URLToInt(entry.E)
		if err != nil || !exponent.IsInt64() {
			continue
		}
		e := exponent.Int64()
		if e < 3 || e > 1<<31-1 || e%2 == 0 {
			continue
		}
		keys[entry.KeyID] = &rsa.PublicKey{N: modulus, E: int(e)}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("fetch signing keys from %s: no usable RSA keys: %w", url, ErrSocialUnavailable)
	}
	return keys, nil
}

func base64URLToInt(raw string) (*big.Int, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(bytes), nil
}

func sha256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func splitDots(token string) []string {
	return strings.Split(token, ".")
}
