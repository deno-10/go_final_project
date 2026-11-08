package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordSession(t *testing.T) {
	auth, err := newAuth("test password")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/signin", strings.NewReader(`{"password":"test password"}`))
	response := httptest.NewRecorder()
	auth.signin(response, request)
	var login struct{ Token string }
	if err := json.Unmarshal(response.Body.Bytes(), &login); err != nil || login.Token == "" {
		t.Fatalf("signin failed: %s", response.Body.String())
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: auth.subject, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}).SignedString(auth.key)
	if err != nil {
		t.Fatal(err)
	}
	// Keep the signing key so this tests the password check independently of key rotation.
	changed := &authenticator{key: auth.key, subject: auth.passwordTag("new password")}
	parts := strings.Split(login.Token, ".")
	for _, tc := range []struct {
		name   string
		auth   *authenticator
		token  string
		status int
	}{
		{"valid", auth, login.Token, http.StatusNoContent},
		{"missing", auth, "", http.StatusUnauthorized},
		{"expired", auth, expired, http.StatusUnauthorized},
		{"bad signature", auth, strings.Join(parts[:2], ".") + ".AAAA", http.StatusUnauthorized},
		{"password changed", changed, login.Token, http.StatusUnauthorized},
		{"password disabled", &authenticator{}, "", http.StatusNoContent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
			if tc.token != "" {
				request.AddCookie(&http.Cookie{Name: "token", Value: tc.token})
			}
			response := httptest.NewRecorder()
			tc.auth.require(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})(response, request)
			if response.Code != tc.status {
				t.Errorf("got HTTP %d, want %d", response.Code, tc.status)
			}
		})
	}
}
