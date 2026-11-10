package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifyGoogleIDTokenWithNEOnlyJWKS(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	jwks := fmt.Sprintf(`{"keys":[{"kid":"google-test-key","kty":"RSA","n":%q,"e":%q}]}`,
		base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
		base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, jwks); err != nil {
			t.Errorf("serve JWKS: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	originalCache := googleKeyCache
	googleKeyCache = &keyCache{ttl: time.Hour, url: server.URL}
	t.Cleanup(func() { googleKeyCache = originalCache })

	now := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"google-test-key"}`))
	claims := fmt.Sprintf(`{"iss":"https://accounts.google.com","sub":"google-subject","aud":"test-web-client","exp":%d,"email":"user@example.test","email_verified":true}`, now.Add(time.Hour).Unix())
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString([]byte(claims))
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign Google ID token: %v", err)
	}
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	identity, err := VerifyGoogleIDToken(ctx, now, token, []string{"test-web-client"})
	if err != nil {
		t.Fatalf("verify Google ID token with n/e-only JWKS: %v", err)
	}
	want := SocialIdentity{
		Provider:      ProviderGoogle,
		Subject:       "google-subject",
		Email:         "user@example.test",
		EmailVerified: true,
	}
	if identity != want {
		t.Fatal("verified identity does not match the signed Google payload")
	}

	signature[0] ^= 1
	tamperedToken := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
	if _, err := VerifyGoogleIDToken(ctx, now, tamperedToken, []string{"test-web-client"}); !errors.Is(err, ErrSocialTokenInvalid) {
		t.Fatalf("tampered signature error = %v, want ErrSocialTokenInvalid", err)
	}
}
