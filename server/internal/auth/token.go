package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type accessClaims struct {
	Subject  string `json:"sub"`
	Username string `json:"username"`
	Expiry   int64  `json:"exp"`
	IssuedAt int64  `json:"iat"`
}

func SignAccessToken(secret, userID, username string, ttl time.Duration, now time.Time) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(accessClaims{
		Subject:  userID,
		Username: username,
		Expiry:   now.Add(ttl).Unix(),
		IssuedAt: now.Unix(),
	})
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	signature := hmacSHA256(secret, unsigned)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func VerifyAccessToken(secret, token string, now time.Time) (userID, username string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", ErrInvalidToken
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", ErrInvalidToken
	}
	var header map[string]string
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return "", "", ErrInvalidToken
	}
	if header["alg"] != "HS256" {
		return "", "", ErrInvalidToken
	}

	expected := hmacSHA256(secret, parts[0]+"."+parts[1])
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, expected) {
		return "", "", ErrInvalidToken
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", ErrInvalidToken
	}
	var claims accessClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return "", "", ErrInvalidToken
	}
	if claims.Subject == "" {
		return "", "", ErrInvalidToken
	}
	if now.Unix() >= claims.Expiry {
		return "", "", ErrExpiredToken
	}
	return claims.Subject, claims.Username, nil
}

func hmacSHA256(secret, message string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func NewRefreshToken() (token, hash string, err error) {
	raw, err := randomBytes(32)
	if err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, HashRefreshToken(token), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
