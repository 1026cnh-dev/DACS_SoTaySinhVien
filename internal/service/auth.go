package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sotaysinhvien/internal/model"
	"sotaysinhvien/internal/reference"
	"sotaysinhvien/internal/repository"
)

const defaultGoogleClientID = "545680002150-rs51scvphjuas8ume163im9fng3j1qld.apps.googleusercontent.com"

type cachedUser struct {
	user    model.User
	expires time.Time
}

type AuthService struct {
	repo           repository.ContentRepository
	adminUsername  string
	adminPassword  string
	googleClientID string
	httpClient     *http.Client
	userCacheMu    sync.RWMutex
	userCache      map[int]cachedUser
}

const userCacheTTL = 20 * time.Second

func NewAuthService(repo repository.ContentRepository) *AuthService {
	return &AuthService{
		repo:           repo,
		adminUsername:  getenv("ADMIN_USER", "admin"),
		adminPassword:  getenv("ADMIN_PASSWORD", "admin"),
		googleClientID: getenv("GOOGLE_CLIENT_ID", defaultGoogleClientID),
		httpClient:     &http.Client{Timeout: 8 * time.Second},
		userCache:      map[int]cachedUser{},
	}
}

func (s *AuthService) GoogleClientID() string { return s.googleClientID }

func accountLockActive(u model.User) bool {
	if !strings.EqualFold(strings.TrimSpace(u.AccountStatus), "locked") {
		return false
	}
	if strings.TrimSpace(u.LockedUntil) == "" {
		return true
	}
	until, err := time.Parse(time.RFC3339, u.LockedUntil)
	return err != nil || time.Now().Before(until)
}

func (s *AuthService) SearchUsers(query string, limit int) ([]model.User, error) {
	return s.repo.SearchUsers(query, limit)
}
func (s *AuthService) SetUserAdmin(id int, value bool) error {
	err := s.repo.SetUserAdmin(id, value)
	if err == nil {
		s.invalidateUser(id)
	}
	return err
}
func (s *AuthService) SetUserAccountStatus(id int, status, lockedUntil string) error {
	err := s.repo.SetUserAccountStatus(id, status, lockedUntil)
	if err == nil {
		s.invalidateUser(id)
	}
	return err
}
func (s *AuthService) DeleteUser(id int) error {
	err := s.repo.DeleteUser(id)
	if err == nil {
		s.invalidateUser(id)
	}
	return err
}
func (s *AuthService) CountAdmins() (int, error) { return s.repo.CountAdmins() }

func (s *AuthService) EnsureAdminUser() (model.User, error) {
	hash, err := hashPassword(s.adminPassword)
	if err != nil {
		return model.User{}, err
	}
	if u, err := s.repo.FindUserByLogin(s.adminUsername); err == nil {
		if err := s.repo.UpdateUserPasswordHash(u.ID, hash); err != nil {
			return model.User{}, err
		}
		if err := s.repo.SetUserAdmin(u.ID, true); err != nil {
			return model.User{}, err
		}
		return s.repo.FindUserByID(u.ID)
	}
	u, err := s.repo.CreateUser(model.User{Name: "Quản trị viên", Email: "admin@local.invalid", Username: s.adminUsername, PasswordHash: hash, Provider: "local", IsAdmin: true})
	if err != nil {
		return model.User{}, err
	}
	if err := s.repo.SetUserAdmin(u.ID, true); err != nil {
		return model.User{}, err
	}
	return s.repo.FindUserByID(u.ID)
}

func (s *AuthService) invalidateUser(id int) {
	if id <= 0 {
		return
	}
	s.userCacheMu.Lock()
	delete(s.userCache, id)
	s.userCacheMu.Unlock()
}

func (s *AuthService) cacheUser(u model.User) {
	if u.ID <= 0 {
		return
	}
	s.userCacheMu.Lock()
	s.userCache[u.ID] = cachedUser{user: u, expires: time.Now().Add(userCacheTTL)}
	s.userCacheMu.Unlock()
}

func (s *AuthService) GetUserByID(id int) (model.User, error) {
	if id <= 0 {
		return model.User{}, sql.ErrNoRows
	}
	now := time.Now()
	s.userCacheMu.RLock()
	if item, ok := s.userCache[id]; ok && now.Before(item.expires) {
		s.userCacheMu.RUnlock()
		return item.user, nil
	}
	s.userCacheMu.RUnlock()
	u, err := s.repo.FindUserByID(id)
	if err == nil {
		s.cacheUser(u)
	}
	return u, err
}

func (s *AuthService) GetUsersByIDs(ids []int) (map[int]model.User, error) {
	result := make(map[int]model.User, len(ids))
	now := time.Now()
	missing := make([]int, 0, len(ids))
	seen := map[int]bool{}
	s.userCacheMu.RLock()
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		if item, ok := s.userCache[id]; ok && now.Before(item.expires) {
			result[id] = item.user
		} else {
			missing = append(missing, id)
		}
	}
	s.userCacheMu.RUnlock()
	if len(missing) == 0 {
		return result, nil
	}
	loaded, err := s.repo.FindUsersByIDs(missing)
	if err != nil {
		return result, err
	}
	for id, u := range loaded {
		result[id] = u
		s.cacheUser(u)
	}
	return result, nil
}

func (s *AuthService) UpdateProfile(id int, name, username, avatarURL, profileRole, school, major, studentID, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo string) (model.User, error) {
	name = strings.TrimSpace(name)
	username = strings.TrimSpace(username)
	avatarURL = strings.TrimSpace(avatarURL)
	profileRole = strings.TrimSpace(profileRole)
	school = strings.TrimSpace(school)
	major = strings.TrimSpace(major)
	studentID = strings.TrimSpace(studentID)
	if profileRole == "student" {
		if school != "" {
			canonicalSchool, ok := reference.CanonicalSchool(school)
			if !ok {
				return model.User{}, errors.New("vui lòng chọn Trường / Đại học từ danh sách gợi ý để dữ liệu cùng trường không bị phân mảnh")
			}
			school = canonicalSchool
		}
		major = reference.CanonicalMajor(major)
	}
	phone = strings.TrimSpace(phone)
	if phone != "" {
		if normalizedPhone, phoneErr := normalizeVNPhone(phone); phoneErr == nil {
			phone = normalizedPhone
		} else {
			return model.User{}, phoneErr
		}
	}
	employerCompany = strings.TrimSpace(employerCompany)
	employerTaxCode = strings.TrimSpace(employerTaxCode)
	employerRepresentative = strings.TrimSpace(employerRepresentative)
	employerWebsite = strings.TrimSpace(employerWebsite)
	landlordName = strings.TrimSpace(landlordName)
	landlordAddress = strings.TrimSpace(landlordAddress)
	landlordPhone = strings.TrimSpace(landlordPhone)
	landlordLegalInfo = strings.TrimSpace(landlordLegalInfo)
	if id <= 0 {
		return model.User{}, errors.New("tài khoản không hợp lệ")
	}
	if name == "" {
		return model.User{}, errors.New("vui lòng nhập họ và tên")
	}
	if username != "" && len([]rune(username)) < 3 {
		return model.User{}, errors.New("tên đăng nhập phải có ít nhất 3 ký tự")
	}
	switch profileRole {
	case "", "student", "employer", "landlord":
	default:
		return model.User{}, errors.New("loại hồ sơ xác thực không hợp lệ")
	}
	validatePhone := func(value string) bool {
		if value == "" {
			return true
		}
		digits := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, value)
		return len(digits) >= 9 && len(digits) <= 12
	}
	if !validatePhone(landlordPhone) {
		return model.User{}, errors.New("số điện thoại liên hệ không hợp lệ")
	}
	if err := s.repo.UpdateUserProfile(id, name, username, avatarURL, profileRole, school, major, studentID, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			if strings.Contains(strings.ToLower(err.Error()), "phone") {
				return model.User{}, errors.New("số điện thoại đã được sử dụng bởi tài khoản khác")
			}
			return model.User{}, errors.New("tên đăng nhập đã được sử dụng")
		}
		return model.User{}, err
	}
	s.invalidateUser(id)
	updated, err := s.repo.FindUserByID(id)
	if err != nil {
		return model.User{}, err
	}

	// Tích xanh là trạng thái uy tín độc lập với vai trò tài khoản.
	// Hệ thống tự cấp khi hồ sơ cơ bản và bộ thông tin của vai trò đã chọn được điền đầy đủ.
	completeCore := strings.TrimSpace(updated.Name) != "" && strings.TrimSpace(updated.Email) != "" && strings.TrimSpace(updated.Phone) != ""
	roleComplete := false
	switch updated.ProfileRole {
	case "student":
		roleComplete = isEducationEmailAddress(updated.Email) && updated.School != "" && updated.Major != "" && updated.StudentID != ""
	case "employer":
		roleComplete = updated.EmployerCompany != "" && updated.EmployerTaxCode != "" && updated.EmployerRepresentative != "" && updated.EmployerWebsite != ""
	case "landlord":
		roleComplete = updated.LandlordName != "" && updated.LandlordAddress != "" && updated.LandlordPhone != "" && updated.LandlordLegalInfo != ""
	}
	if !updated.IsVerified && completeCore && roleComplete {
		if err := s.repo.SetUserTrustVerified(id, true); err == nil {
			updated.IsVerified = true
		}
	}
	s.cacheUser(updated)
	return updated, nil
}

func isEducationEmailAddress(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	return strings.HasSuffix(domain, ".edu.vn") || strings.HasSuffix(domain, ".edu")
}

func normalizeVNPhone(value string) (string, error) {
	raw := strings.TrimSpace(value)
	var digits strings.Builder
	for _, ch := range raw {
		if ch >= '0' && ch <= '9' {
			digits.WriteRune(ch)
		}
	}
	phone := digits.String()
	if strings.HasPrefix(phone, "84") && len(phone) >= 11 {
		phone = "0" + phone[2:]
	}
	if len(phone) != 10 || !strings.HasPrefix(phone, "0") {
		return "", errors.New("Số điện thoại phải gồm 10 chữ số, ví dụ 0912345678")
	}
	validPrefix := phone[1] == '3' || phone[1] == '5' || phone[1] == '7' || phone[1] == '8' || phone[1] == '9'
	if !validPrefix {
		return "", errors.New("Số điện thoại di động Việt Nam không hợp lệ")
	}
	return phone, nil
}

func (s *AuthService) PhoneAccountStatus(value string) (bool, model.User, error) {
	phone, err := normalizeVNPhone(value)
	if err != nil {
		return false, model.User{}, err
	}
	u, err := s.repo.FindUserByPhone(phone)
	if errors.Is(err, sql.ErrNoRows) {
		return false, model.User{}, nil
	}
	if err != nil {
		return false, model.User{}, err
	}
	return true, u, nil
}

func (s *AuthService) IsAdmin(login, password string) bool {
	return login == s.adminUsername && password == s.adminPassword
}

func (s *AuthService) Register(phoneValue, password string) (model.User, error) {
	phone, err := normalizeVNPhone(phoneValue)
	if err != nil {
		return model.User{}, err
	}
	if len(password) < 6 {
		return model.User{}, errors.New("Mật khẩu phải có ít nhất 6 ký tự")
	}
	if _, err := s.repo.FindUserByPhone(phone); err == nil {
		return model.User{}, errors.New("Số điện thoại này đã có tài khoản. Hãy chuyển sang Đăng nhập.")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return model.User{}, err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.User{}, err
	}
	// Email kỹ thuật chỉ dùng để giữ tương thích với cấu trúc dữ liệu cũ và không hiển thị cho tài khoản số điện thoại.
	internalEmail := "phone." + phone + "@local.invalid"
	return s.repo.CreateUser(model.User{Name: "Thành viên", Email: internalEmail, Username: "", PasswordHash: hash, Provider: "phone", Phone: phone})
}

func (s *AuthService) Login(login, password string) (model.User, error) {
	login = strings.TrimSpace(login)
	if phone, phoneErr := normalizeVNPhone(login); phoneErr == nil {
		login = phone
	}
	u, err := s.repo.FindUserByLogin(login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, errors.New("Sai tài khoản hoặc mật khẩu")
		}
		return model.User{}, err
	}
	if u.PasswordHash == "" || !verifyPassword(password, u.PasswordHash) {
		return model.User{}, errors.New("Sai tài khoản hoặc mật khẩu")
	}
	if accountLockActive(u) {
		return model.User{}, errors.New("Tài khoản đang bị tạm khóa. Vui lòng liên hệ quản trị viên.")
	}
	if isLegacyPasswordHash(u.PasswordHash) {
		if upgraded, hashErr := hashPassword(password); hashErr == nil {
			_ = s.repo.UpdateUserPasswordHash(u.ID, upgraded)
			u.PasswordHash = upgraded
		}
	}
	return u, nil
}

type googleTokenInfo struct {
	Audience      string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Subject       string `json:"sub"`
	Issuer        string `json:"iss"`
	ExpiresAt     string `json:"exp"`
	Error         string `json:"error_description"`
}

func (s *AuthService) LoginWithGoogle(credential string) (model.User, error) {
	credential = strings.TrimSpace(credential)
	if credential == "" {
		return model.User{}, errors.New("Thiếu Google credential")
	}
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(credential)
	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		return model.User{}, fmt.Errorf("không thể xác minh Google: %w", err)
	}
	defer resp.Body.Close()
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return model.User{}, errors.New("phản hồi Google không hợp lệ")
	}
	if resp.StatusCode != http.StatusOK || info.Error != "" {
		return model.User{}, errors.New("Google credential không hợp lệ")
	}
	if info.Audience != s.googleClientID {
		return model.User{}, errors.New("Google Client ID không khớp")
	}
	if info.Issuer != "accounts.google.com" && info.Issuer != "https://accounts.google.com" {
		return model.User{}, errors.New("Google issuer không hợp lệ")
	}
	if info.Email == "" || info.Subject == "" {
		return model.User{}, errors.New("Google không trả về email hợp lệ")
	}
	if info.EmailVerified != "true" && info.EmailVerified != "True" {
		return model.User{}, errors.New("Email Google chưa được xác minh")
	}
	name := strings.TrimSpace(info.Name)
	if name == "" {
		name = strings.Split(info.Email, "@")[0]
	}
	u, err := s.repo.UpsertGoogleUser(name, info.Email, info.Subject)
	if err != nil {
		return model.User{}, err
	}
	if accountLockActive(u) {
		return model.User{}, errors.New("Tài khoản đang bị tạm khóa. Vui lòng liên hệ quản trị viên.")
	}
	return u, nil
}

const pbkdf2Iterations = 210000

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := 32
	blocks := (keyLen + hashLen - 1) / hashLen
	out := make([]byte, 0, blocks*hashLen)
	for block := 1; block <= blocks; block++ {
		msg := make([]byte, len(salt)+4)
		copy(msg, salt)
		msg[len(salt)] = byte(block >> 24)
		msg[len(salt)+1] = byte(block >> 16)
		msg[len(salt)+2] = byte(block >> 8)
		msg[len(salt)+3] = byte(block)
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(msg)
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := pbkdf2SHA256([]byte(password), salt, pbkdf2Iterations, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", pbkdf2Iterations,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(dk)), nil
}

func verifyPassword(password, encoded string) bool {
	if strings.HasPrefix(encoded, "pbkdf2_sha256$") {
		parts := strings.Split(encoded, "$")
		if len(parts) != 4 {
			return false
		}
		iterations, err := strconv.Atoi(parts[1])
		if err != nil || iterations < 100000 {
			return false
		}
		salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
		expected, err2 := base64.RawStdEncoding.DecodeString(parts[3])
		if err1 != nil || err2 != nil || len(expected) == 0 {
			return false
		}
		actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
		return subtle.ConstantTimeCompare(expected, actual) == 1
	}
	// Hỗ trợ hash SHA-256 cũ để người dùng hiện tại vẫn đăng nhập được.
	parts := strings.Split(encoded, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err1 := hex.DecodeString(parts[0])
	expected, err2 := hex.DecodeString(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	h := sha256.Sum256(append(salt, []byte(password)...))
	return len(expected) == len(h) && subtle.ConstantTimeCompare(expected, h[:]) == 1
}

func isLegacyPasswordHash(encoded string) bool { return !strings.HasPrefix(encoded, "pbkdf2_sha256$") }
func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
