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
	"strings"

	_ "github.com/lib/pq"      // Driver cho PostgreSQL (Supabase)
	_ "modernc.org/sqlite"     // Driver cho SQLite (Local)
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
var isPostgres bool // Cờ nhận diện môi trường

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// Hàm thông minh: Tự động đổi dấu "?" (của SQLite) sang "$1, $2" nếu đang chạy Postgres
func adaptQuery(query string) string {
	if !isPostgres {
		return query
	}
	parts := strings.Split(query, "?")
	var result strings.Builder
	for i := 0; i < len(parts)-1; i++ {
		result.WriteString(parts[i])
		result.WriteString("$" + strconv.Itoa(i+1))
	}
	result.WriteString(parts[len(parts)-1])
	return result.String()
}

func initDB() {
	dbUrl := os.Getenv("DATABASE_URL")
	var err error

	// NẾU CÓ BIẾN MÔI TRƯỜNG DATABASE_URL -> CHẠY SUPABASE
	if dbUrl != "" {
		isPostgres = true
		db, err = sql.Open("postgres", dbUrl)
		if err != nil {
			log.Fatal("Lỗi kết nối Supabase: ", err)
		}
		fmt.Println("🚀 Đã kết nối với Database Supabase (PostgreSQL) thành công!")
	} else {
		// NẾU KHÔNG CÓ -> CHẠY SQLITE LOCAL
		isPostgres = false
		_, errCheck := os.Stat("database.db")
		dbExists := !os.IsNotExist(errCheck)

		db, err = sql.Open("sqlite", "database.db")
		if err != nil {
			log.Fatal(err)
		}

		if !dbExists {
			fmt.Println("Đang khởi tạo SQLite từ web.sql...")
			sqlBytes, err := os.ReadFile("web.sql")
			if err != nil {
				log.Fatal("Lỗi: Không tìm thấy file web.sql! ", err)
			}
			_, err = db.Exec(string(sqlBytes))
			if err != nil {
				log.Fatal("Lỗi khi chạy lệnh SQL: ", err)
			}

			adminPass := hashPassword("123456")
			db.Exec("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, ?, ?, ?)", "Quản Trị Viên", "admin@example.com", adminPass, "local", "admin")
			fmt.Println("Đã nạp dữ liệu từ web.sql và tạo Admin thành công!")
		} else {
			fmt.Println("💻 Đã kết nối với SQLite Local thành công!")
		}
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
			err := db.QueryRow(adaptQuery("SELECT id, full_name, email, role FROM users WHERE email = ? AND password_hash = ? AND auth_provider = 'local'"), creds.Email, hashed).Scan(&user.ID, &user.FullName, &user.Email, &user.Role)
			if err != nil {
				http.Error(w, `{"error": "Sai email hoặc mật khẩu"}`, http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(user)
		}
	})

	// 2. API Đăng ký thủ công
	http.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var newUser User
			json.NewDecoder(r.Body).Decode(&newUser)

			if !strings.HasSuffix(newUser.Email, "@dlu.edu.vn") {
				http.Error(w, `{"error": "Chỉ chấp nhận email của trường Đại học Đà Lạt (@dlu.edu.vn)"}`, http.StatusBadRequest)
				return
			}

			hashed := hashPassword(newUser.Password)
			_, err := db.Exec(adaptQuery("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, ?, 'local', ?)"), newUser.FullName, newUser.Email, hashed, "student")
			if err != nil {
				http.Error(w, `{"error": "Email này đã tồn tại hoặc có lỗi xảy ra"}`, http.StatusBadRequest)
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

			if !strings.HasSuffix(googleData.Email, "@dlu.edu.vn") {
				http.Error(w, `{"error": "Vui lòng sử dụng email trường (@dlu.edu.vn) để đăng nhập"}`, http.StatusBadRequest)
				return
			}

			var user User
			err = db.QueryRow(adaptQuery("SELECT id, full_name, email, role FROM users WHERE email = ?"), googleData.Email).Scan(&user.ID, &user.FullName, &user.Email, &user.Role)

			if err == sql.ErrNoRows {
				// Cú pháp lấy ID sau khi INSERT giữa Postgres và SQLite khác nhau
				var insertedID int
				if isPostgres {
					err = db.QueryRow(adaptQuery("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, '', 'google', 'student') RETURNING id"), googleData.Name, googleData.Email).Scan(&insertedID)
				} else {
					res, errExec := db.Exec(adaptQuery("INSERT INTO users (full_name, email, password_hash, auth_provider, role) VALUES (?, ?, '', 'google', 'student')"), googleData.Name, googleData.Email)
					if errExec == nil {
						id, _ := res.LastInsertId()
						insertedID = int(id)
					}
					err = errExec
				}

				if err != nil {
					http.Error(w, `{"error": "Lỗi tạo tài khoản từ Google"}`, http.StatusInternalServerError)
					return
				}
				user = User{ID: insertedID, FullName: googleData.Name, Email: googleData.Email, Role: "student"}
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
				row := db.QueryRow(adaptQuery("SELECT id, title, content, category_id FROM posts WHERE id = ?"), idQuery)
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
				rows, _ = db.Query(adaptQuery("SELECT id, title, content, category_id FROM posts WHERE category_id = ? ORDER BY id DESC"), catID)
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
			_, err := db.Exec(adaptQuery("INSERT INTO posts (title, content, category_id) VALUES (?, ?, ?)"), newPost.Title, newPost.Content, newPost.CategoryID)
			if err != nil {
				http.Error(w, `{"error": "Lỗi lưu dữ liệu"}`, http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"message": "Đăng bài thành công!"})
			return
		}
	})

	// HỖ TRỢ CỔNG ĐỘNG (DYNAMIC PORT) CHO RENDER VÀ LOCAL
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Cổng mặc định khi chạy Local
	}

	frontend, _ := fs.Sub(templateFiles, "templates")
	http.Handle("/", http.FileServer(http.FS(frontend)))
	
	fmt.Println("🌐 Server đang lắng nghe tại cổng :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
