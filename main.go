package main

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

//go:embed templates
var templateFiles embed.FS

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type Post struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CategoryID int    `json:"category_id"`
}
type User struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

var db *sql.DB

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func initDB() {
	var err error
	_, errCheck := os.Stat("database.db")
	dbExists := !os.IsNotExist(errCheck)

	db, err = sql.Open("sqlite", "database.db")
	if err != nil {
		log.Fatal(err)
	}

	// Đảm bảo bảng users luôn tồn tại với cấu trúc chuẩn để tránh lỗi khi đăng ký
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		auth_provider TEXT DEFAULT 'local',
		role TEXT DEFAULT 'student'
	);`
	_, err = db.Exec(createUsersTableSQL)
	if err != nil {
		log.Fatal("Lỗi tạo bảng users: ", err)
	}

	if !dbExists {
		fmt.Println("Phát hiện lần chạy đầu tiên. Đang khởi tạo Database từ web.sql (nếu có)...")
		sqlBytes, err := os.ReadFile("web.sql")
		if err == nil {
			_, err = db.Exec(string(sqlBytes))
			if err != nil {
				fmt.Println("Cảnh báo khi chạy web.sql: ", err)
			} else {
				fmt.Println("Đã nạp dữ liệu từ web.sql thành công!")
			}
		} else {
			fmt.Println("Không tìm thấy file web.sql, hệ thống sử dụng cấu trúc mặc định.")
		}

		// Tạo tài khoản Admin mẫu nếu chưa có
		var count int
		db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if count == 0 {
			adminPass := hashPassword("123456")
			_, err = db.Exec("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, ?, ?, ?)", "Quản Trị Viên", "admin@example.com", adminPass, "local", "admin")
			if err == nil {
				fmt.Println("Đã tạo tài khoản Admin mặc định (admin@example.com / 123456) thành công!")
			}
		}
	} else {
		fmt.Println("Đã kết nối với Database hiện tại thành công!")
	}
}

func main() {
	initDB()

	// 1. API Đăng nhập thủ công
	http.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var creds User
			json.NewDecoder(r.Body).Decode(&creds)
			hashed := hashPassword(creds.Password)

			var user User
			err := db.QueryRow("SELECT id, full_name, email, role FROM users WHERE email = ? AND password_hash = ? AND auth_provider = 'local'", creds.Email, hashed).Scan(&user.ID, &user.FullName, &user.Email, &user.Role)
			if err != nil {
				http.Error(w, `{"error": "Sai email hoặc mật khẩu"}`, http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(user)
		}
	})

	// 2. API Đăng ký thủ công (Đã bỏ chặn email trường, hỗ trợ mọi email)
	http.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var newUser User
			json.NewDecoder(r.Body).Decode(&newUser)

			if newUser.Email == "" || newUser.Password == "" || newUser.FullName == "" {
				http.Error(w, `{"error": "Vui lòng nhập đầy đủ thông tin"}`, http.StatusBadRequest)
				return
			}

			hashed := hashPassword(newUser.Password)
			_, err := db.Exec("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, ?, 'local', ?)", newUser.FullName, newUser.Email, hashed, "student")
			if err != nil {
				http.Error(w, `{"error": "Email này đã được sử dụng hoặc có lỗi xảy ra"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"message": "Đăng ký thành công!"})
		}
	})

	// 3. API Đăng nhập bằng Google
	http.HandleFunc("/api/auth/google", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var reqData struct {
				Credential string `json:"credential"`
			}
			json.NewDecoder(r.Body).Decode(&reqData)

			resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + reqData.Credential)
			if err != nil || resp.StatusCode != 200 {
				http.Error(w, `{"error": "Token Google không hợp lệ"}`, http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			var googleData struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			}
			json.NewDecoder(resp.Body).Decode(&googleData)

			var user User
			err = db.QueryRow("SELECT id, full_name, email, role FROM users WHERE email = ?", googleData.Email).Scan(&user.ID, &user.FullName, &user.Email, &user.Role)

			if err == sql.ErrNoRows {
				res, err := db.Exec("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, '', 'google', 'student')", googleData.Name, googleData.Email)
				if err != nil {
					http.Error(w, `{"error": "Lỗi tạo tài khoản từ Google"}`, http.StatusInternalServerError)
					return
				}
				id, _ := res.LastInsertId()
				user = User{ID: int(id), FullName: googleData.Name, Email: googleData.Email, Role: "student"}
			} else if err != nil {
				http.Error(w, `{"error": "Lỗi cơ sở dữ liệu"}`, http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(user)
		}
	})

	// 4. API Danh mục
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			rows, _ := db.Query("SELECT category_id, category_name FROM categories")
			defer rows.Close()
			var cats []Category
			for rows.Next() {
				var c Category
				rows.Scan(&c.ID, &c.Name)
				cats = append(cats, c)
			}
			json.NewEncoder(w).Encode(cats)
		}
	})

	// 5. API Bài Viết (CRUD)
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			idQuery := r.URL.Query().Get("id")
			if idQuery != "" {
				row := db.QueryRow("SELECT id, title, content, category_id FROM posts WHERE id = ?", idQuery)
				var p Post
				if err := row.Scan(&p.ID, &p.Title, &p.Content, &p.CategoryID); err != nil {
					http.Error(w, `{"error": "Không tìm thấy bài viết"}`, http.StatusNotFound)
					return
				}
				json.NewEncoder(w).Encode(p)
				return
			}

			categoryQuery := r.URL.Query().Get("category_id")
			var rows *sql.Rows
			if categoryQuery != "" {
				catID, _ := strconv.Atoi(categoryQuery)
				rows, _ = db.Query("SELECT id, title, content, category_id FROM posts WHERE category_id = ? ORDER BY id DESC", catID)
			} else {
				rows, _ = db.Query("SELECT id, title, content, category_id FROM posts ORDER BY id DESC")
			}
			defer rows.Close()

			var posts []Post
			for rows.Next() {
				var p Post
				rows.Scan(&p.ID, &p.Title, &p.Content, &p.CategoryID)
				posts = append(posts, p)
			}
			if posts == nil {
				posts = []Post{}
			}
			json.NewEncoder(w).Encode(posts)
			return
		}

		if r.Method == http.MethodPost {
			var newPost Post
			json.NewDecoder(r.Body).Decode(&newPost)
			_, err := db.Exec("INSERT INTO posts (title, content, category_id) VALUES (?, ?, ?)", newPost.Title, newPost.Content, newPost.CategoryID)
			if err != nil {
				http.Error(w, `{"error": "Lỗi lưu dữ liệu"}`, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"message": "Đăng bài thành công!"})
			return
		}
	})

	// ==========================================
	// CẤU HÌNH ĐIỀU HƯỚNG VÀ PHỤC VỤ FILE TĨNH
	// ==========================================
	frontend, err := fs.Sub(templateFiles, "templates")
	if err != nil {
		log.Fatal("Lỗi khi nạp thư mục templates: ", err)
	}

	fileServer := http.FileServer(http.FS(frontend))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/student_handbook.html", http.StatusFound)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	fmt.Println("Server đang chạy! Vui lòng truy cập: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
