package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"sotaysinhvien/internal/model"
)

type UserLoader func(id int) (model.User, error)

type Manager struct {
	AdminCookieName string
	UserCookieName  string
	secret          []byte
	secureCookies   bool
	loadUser        UserLoader
}

func NewManager(loadUser UserLoader) *Manager {
	return &Manager{
		AdminCookieName: "admin_session",
		UserCookieName:  "user_session",
		secret:          loadOrCreateSecret(),
		secureCookies:   strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true") || strings.EqualFold(os.Getenv("APP_ENV"), "production") || os.Getenv("RENDER") != "",
		loadUser:        loadUser,
	}
}

func loadOrCreateSecret() []byte {
	if v := strings.TrimSpace(os.Getenv("SESSION_SECRET")); v != "" {
		return []byte(v)
	}
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") || os.Getenv("RENDER") != "" {
		panic("SESSION_SECRET là bắt buộc trong môi trường production")
	}
	const path = ".session_secret"
	if b, err := os.ReadFile(path); err == nil && len(b) >= 32 {
		return b
	}
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		panic("không thể tạo session secret: " + err.Error())
	}
	_ = os.WriteFile(path, b, 0600)
	return b
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
}

func (m *Manager) verify(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	payload := string(raw)
	expected := m.sign(payload)
	return payload, hmac.Equal([]byte(expected), []byte(token))
}

func (m *Manager) setCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: m.secureCookies,
	})
}

func (m *Manager) IsAdmin(r *http.Request) bool {
	c, err := r.Cookie(m.AdminCookieName)
	if err != nil {
		return false
	}
	payload, ok := m.verify(c.Value)
	if !ok {
		return false
	}
	parts := strings.Split(payload, "|")
	if len(parts) != 2 || parts[0] != "admin" {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil && time.Now().Unix() < exp
}

func (m *Manager) LoginAdmin(w http.ResponseWriter) {
	exp := time.Now().Add(12 * time.Hour).Unix()
	m.setCookie(w, m.AdminCookieName, m.sign(fmt.Sprintf("admin|%d", exp)), int((12 * time.Hour).Seconds()))
}

func (m *Manager) LoginUser(w http.ResponseWriter, u model.User) {
	exp := time.Now().Add(7 * 24 * time.Hour).Unix()
	payload := fmt.Sprintf("user|%d|%d", u.ID, exp)
	m.setCookie(w, m.UserCookieName, m.sign(payload), int((7 * 24 * time.Hour).Seconds()))
}

func (m *Manager) CurrentUser(r *http.Request) (model.User, bool) {
	if m.loadUser == nil {
		return model.User{}, false
	}
	c, err := r.Cookie(m.UserCookieName)
	if err != nil {
		return model.User{}, false
	}
	payload, ok := m.verify(c.Value)
	if !ok {
		return model.User{}, false
	}
	parts := strings.Split(payload, "|")
	if len(parts) != 3 || parts[0] != "user" {
		return model.User{}, false
	}
	id, err1 := strconv.Atoi(parts[1])
	exp, err2 := strconv.ParseInt(parts[2], 10, 64)
	if err1 != nil || err2 != nil || time.Now().Unix() >= exp {
		return model.User{}, false
	}
	u, err := m.loadUser(id)
	if err != nil {
		return model.User{}, false
	}
	if strings.EqualFold(strings.TrimSpace(u.AccountStatus), "locked") {
		locked := true
		if strings.TrimSpace(u.LockedUntil) != "" {
			if until, parseErr := time.Parse(time.RFC3339, u.LockedUntil); parseErr == nil && !time.Now().Before(until) {
				locked = false
			}
		}
		if locked {
			return model.User{}, false
		}
	}
	return u, true
}

func (m *Manager) RequireUser(r *http.Request) (model.User, error) {
	if u, ok := m.CurrentUser(r); ok {
		return u, nil
	}
	return model.User{}, errors.New("chưa đăng nhập")
}

func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	m.setCookie(w, m.AdminCookieName, "", -1)
	m.setCookie(w, m.UserCookieName, "", -1)
}
