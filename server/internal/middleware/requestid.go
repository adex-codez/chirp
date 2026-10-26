package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

// requestIDCounter keeps generated IDs unique if the system entropy source
// ever fails.
var requestIDCounter atomic.Uint64

// RequestID assigns every request a correlation ID, attaches it to the request
// context, and echoes it in the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestid.Header)
		if id == "" {
			id = newRequestID()
		}

		c.Header(requestid.Header, id)
		c.Request = c.Request.WithContext(requestid.With(c.Request.Context(), id))

		c.Next()
	}
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%x-%x", time.Now().UnixNano(), requestIDCounter.Add(1))
	}

	return hex.EncodeToString(buf[:])
}
