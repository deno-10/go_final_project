package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authenticator struct {
	key     []byte
	subject string
}

func newAuth(password string) (*authenticator, error) {
	a := &authenticator{}
	if password == "" {
		return a, nil
	}
	a.key = make([]byte, sha256.Size)
	if _, err := rand.Read(a.key); err != nil {
		return nil, err
	}
	a.subject = a.passwordTag(password)
	return a, nil
}

// Use a keyed digest because the JWT payload is readable by its holder.
func (a *authenticator) passwordTag(password string) string {
	mac := hmac.New(sha256.New, a.key)
	mac.Write([]byte("password:" + password))
	return hex.EncodeToString(mac.Sum(nil))
}

func (a *authenticator) signin(w http.ResponseWriter, r *http.Request) {
	if a.key == nil {
		writeError(w, http.StatusForbidden, "Вход по паролю отключён")
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if !hmac.Equal([]byte(a.passwordTag(input.Password)), []byte(a.subject)) {
		writeError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}
	claims := jwt.RegisteredClaims{
		Subject:   a.subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось создать токен")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (a *authenticator) require(next http.HandlerFunc) http.HandlerFunc {
	if a.key == nil {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err == nil {
			_, err = jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{},
				func(*jwt.Token) (any, error) { return a.key, nil },
				jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithSubject(a.subject))
		}
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Требуется вход по паролю")
			return
		}
		next(w, r)
	}
}
