package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// NewVerificationCode returns a human-typable 6-digit code and its hash.
// Only the hash is stored; the code travels to the User through the
// verification channel (or the dev-only response field until a mail
// sender exists).
func NewVerificationCode() (code, hash string, err error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", "", err
	}
	code = fmt.Sprintf("%06d", n.Int64()+100000)
	return code, HashVerificationCode(code), nil
}

func HashVerificationCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func randomBytes(n int) ([]byte, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	return raw, nil
}
