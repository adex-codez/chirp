package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/apicodegen"
	"backend/internal/requestid"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

func TestEveryCodeHasAFailure(t *testing.T) {
	codes, err := apicodegen.ParseCodes("codes.go")
	if err != nil {
		t.Fatalf("parse codes: %v", err)
	}

	for _, code := range codes {
		entry, ok := failures[code]
		if !ok {
			t.Errorf("code %q has no entry in the failures table", code)
			continue
		}
		if entry.status < 400 || entry.status > 599 {
			t.Errorf("code %q maps to status %d, want a 4xx or 5xx", code, entry.status)
		}
		if entry.message == "" {
			t.Errorf("code %q has no message", code)
		}
	}

	if len(failures) != len(codes) {
		t.Errorf("failures table has %d entries, codes.go defines %d", len(failures), len(codes))
	}
}

func TestFromMapsEveryDomainError(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"validation", service.ValidationError{Fields: map[string]string{"email": "bad"}}, http.StatusBadRequest, CodeValidationFailed},
		{"username taken", service.ErrUsernameTaken, http.StatusConflict, CodeUsernameTaken},
		{"email in use", fmt.Errorf("sign up: %w", service.ErrEmailTaken), http.StatusConflict, CodeEmailInUse},
		{"invalid sign in", service.ErrInvalidSignIn, http.StatusUnauthorized, CodeInvalidCredentials},
		{"invalid session", fmt.Errorf("wrap: %w", service.ErrInvalidSession), http.StatusUnauthorized, CodeInvalidSession},
		{"session revoked", service.ErrSessionRevoked, http.StatusUnauthorized, CodeSessionRevoked},
		{"email not verified", service.ErrEmailNotVerified, http.StatusForbidden, CodeEmailNotVerified},
		{"invalid challenge", service.ErrInvalidChallenge, http.StatusBadRequest, CodeInvalidCode},
		{"challenge locked", service.ErrChallengeLocked, http.StatusTooManyRequests, CodeChallengeLocked},
		{"no pending verify", service.ErrNoPendingVerify, http.StatusNotFound, CodeNoPendingVerify},
		{"already verified", service.ErrAlreadyVerified, http.StatusBadRequest, CodeAlreadyVerified},
		{"resend too soon", service.ErrResendTooSoon, http.StatusTooManyRequests, CodeResendTooSoon},
		{"resend limit", service.ErrResendLimit, http.StatusTooManyRequests, CodeResendLimit},
		{"user not found", service.ErrUserNotFound, http.StatusNotFound, CodeUserNotFound},
		{"social misconfigured", service.ErrSocialMisconfigured, http.StatusServiceUnavailable, CodeSocialMisconfigured},
		{"social no email", service.ErrSocialNoEmail, http.StatusBadRequest, CodeSocialNoEmail},
		{"social unverified", service.ErrSocialUnverified, http.StatusForbidden, CodeSocialUnverified},
		{"social conflict", service.ErrSocialConflict, http.StatusConflict, CodeSocialConflict},
		{"social token invalid", service.ErrSocialTokenInvalid, http.StatusUnauthorized, CodeSocialTokenInvalid},
		{"social unavailable", service.ErrSocialUnavailable, http.StatusBadGateway, CodeSocialUnavailable},
		{"username already set", service.ErrUsernameAlreadySet, http.StatusConflict, CodeUsernameAlreadySet},
		{"password already set", service.ErrPasswordAlreadySet, http.StatusConflict, CodePasswordAlreadySet},
		{"wrapped sentinel", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", service.ErrUserNotFound)), http.StatusNotFound, CodeUserNotFound},
		{"unknown error", errors.New("dial tcp: connection refused"), http.StatusInternalServerError, CodeInternalError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problem := resolve(From(tc.err))
			if problem.Code != tc.code {
				t.Errorf("code = %q, want %q", problem.Code, tc.code)
			}
			if problem.Status != tc.status {
				t.Errorf("status = %d, want %d", problem.Status, tc.status)
			}
		})
	}
}

func TestFromKeepsValidationFields(t *testing.T) {
	fields := map[string]string{"email": "email must be a valid address"}
	problem := resolve(From(service.ValidationError{Fields: fields}))
	if problem.Fields["email"] != fields["email"] {
		t.Fatalf("fields = %v, want %v", problem.Fields, fields)
	}
}

func TestUnknownErrorCollapsesToInternalError(t *testing.T) {
	problem := resolve(Problem{Code: "not_a_code", Fields: map[string]string{"email": "leak"}})
	if problem.Code != CodeInternalError {
		t.Fatalf("code = %q, want %q", problem.Code, CodeInternalError)
	}
	if problem.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", problem.Status, http.StatusInternalServerError)
	}
	if problem.Fields != nil {
		t.Fatalf("fields = %v, want nil so no detail leaks", problem.Fields)
	}
}

func TestFailWritesTheCodedFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/thing", func(c *gin.Context) {
		Fail(c, fmt.Errorf("sign up: %w", service.ErrEmailTaken))
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}

	var body Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != CodeEmailInUse {
		t.Errorf("code = %q, want %q", body.Code, CodeEmailInUse)
	}
	if body.Error != failures[CodeEmailInUse].message {
		t.Errorf("message = %q, want the table message", body.Error)
	}
	if body.Fields != nil {
		t.Errorf("fields = %v, want nil so no detail leaks", body.Fields)
	}
}

func TestFailCodeWritesTheTableMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/thing", func(c *gin.Context) {
		FailCode(c, CodeServiceUnavailable)
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}

	var body Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != CodeServiceUnavailable {
		t.Errorf("code = %q, want %q", body.Code, CodeServiceUnavailable)
	}
	if body.Error != failures[CodeServiceUnavailable].message {
		t.Errorf("message = %q, want the table message", body.Error)
	}
}

func TestValidationWritesFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/thing", func(c *gin.Context) {
		Validation(c, map[string]string{"email": "email must be a valid address"})
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/thing", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var body Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != CodeValidationFailed {
		t.Errorf("code = %q, want %q", body.Code, CodeValidationFailed)
	}
	if body.Fields["email"] != "email must be a valid address" {
		t.Errorf("fields = %v, want the per-field message", body.Fields)
	}
}

func TestFailLogs5xxWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logged bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	engine := gin.New()
	engine.GET("/thing", func(c *gin.Context) {
		Fail(c, errors.New("dial tcp: connection refused"))
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/thing", nil)
	request = request.WithContext(requestid.With(request.Context(), "req-123"))
	engine.ServeHTTP(recorder, request)

	line := logged.String()
	for _, want := range []string{
		`"request_id":"req-123"`,
		`"status":500`,
		`"code":"internal_error"`,
		"dial tcp: connection refused",
	} {
		if !bytes.Contains([]byte(line), []byte(want)) {
			t.Errorf("log line is missing %q: %s", want, line)
		}
	}
}
