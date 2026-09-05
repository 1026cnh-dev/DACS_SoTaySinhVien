package main

import (
	"compress/gzip"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"sotaysinhvien/internal/config"
	"sotaysinhvien/internal/handler"
	"sotaysinhvien/internal/repository"
	"sotaysinhvien/internal/service"
	"sotaysinhvien/internal/session"
)

//go:embed web/*/*.html web/*/*.css web/*/*.js web/icons/*.svg web/icons/*.webp web/icons/*.txt
var webFS embed.FS

func main() {
	runtimeConfig, err := config.Load()
	if err != nil {
		log.Fatal("config: ", err)
	}
	if err := validateProductionEnv(); err != nil {
		log.Fatal(err)
	}

	var repo repository.ContentRepository
	var closeRepo func() error
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		pgRepo, openErr := repository.NewPostgresRepository(databaseURL)
		if openErr != nil {
			log.Fatal(openErr)
		}
		repo = pgRepo
		closeRepo = pgRepo.Close
		log.Println("database: PostgreSQL / Supabase")
	} else {
		dbPath := strings.TrimSpace(os.Getenv("SQLITE_PATH"))
		if dbPath == "" {
			dbPath = "student_handbook.db"
		}
		sqliteRepo, openErr := repository.NewSQLiteRepository(dbPath)
		if openErr != nil {
			log.Fatal(openErr)
		}
		repo = sqliteRepo
		closeRepo = sqliteRepo.Close
		if absPath, absErr := filepath.Abs(dbPath); absErr == nil {
			log.Printf("database: SQLite local (%s)", absPath)
		} else {
			log.Printf("database: SQLite local (%s)", dbPath)
		}
	}
	defer closeRepo()

	funcs := template.FuncMap{
		"b64":    func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) },
		"dateVN": func(t time.Time) string { return t.Format("02/01/2006") },
		"audienceHas": func(raw, want string) bool {
			want = strings.ToLower(strings.TrimSpace(want))
			for _, part := range strings.Split(raw, ",") {
				if strings.ToLower(strings.TrimSpace(part)) == want {
					return true
				}
			}
			return false
		},
		"add": func(a, b int) int { return a + b },
		"isPDF": func(path string) bool {
			path = strings.ToLower(strings.TrimSpace(strings.Split(path, "?")[0]))
			return strings.HasSuffix(path, ".pdf")
		},
		"excerpt": func(s string) string {
			r := []rune(s)
			if len(r) > 145 {
				return string(r[:145]) + "…"
			}
			return s
		},
	}
	tmpl := template.Must(template.New("").Funcs(funcs).ParseFS(webFS, "web/*/*.html"))
	content := service.NewContentService(repo)
	auth := service.NewAuthService(repo)
	if _, err := auth.EnsureAdminUser(); err != nil {
		log.Printf("không thể chuẩn bị tài khoản quản trị: %v", err)
	}
	sessions := session.NewManager(auth.GetUserByID)
	h := handler.New(content, auth, sessions, tmpl)

	mux := http.NewServeMux()
	static, _ := fs.Sub(webFS, "web")
	staticHandler := http.StripPrefix("/web/", http.FileServer(http.FS(static)))
	mux.Handle("/web/", cacheStatic(staticHandler))
	uploadDir := strings.TrimSpace(os.Getenv("UPLOAD_DIR"))
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	_ = os.MkdirAll(uploadDir, 0755)
	uploadHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir)))
	mux.Handle("/uploads/", cacheUploads(uploadHandler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	h.RegisterRoutes(mux)

	var app http.Handler = mux
	app = securityHeaders(app)
	app = gzipMiddleware(app)
	if os.Getenv("DACS_ACCESS_LOG") == "1" {
		app = logging(app)
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	host := "127.0.0.1"
	if strings.EqualFold(runtimeConfig.Environment, "production") || strings.TrimSpace(os.Getenv("RENDER")) != "" {
		host = "0.0.0.0"
	}
	addr := host + ":" + port
	server := &http.Server{
		Addr:              addr,
		Handler:           app,
		ReadHeaderTimeout: 8 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       75 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	displayURL := "http://localhost:" + port
	if host != "127.0.0.1" {
		displayURL = "http://" + addr
	}
	fmt.Printf("DACS đang chạy tại %s · môi trường=%s · config=%s\n", displayURL, runtimeConfig.Environment, runtimeConfig.ConfigFile)
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-errCh:
		log.Fatal(err)
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}
}

func validateProductionEnv() error {
	production := strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") || strings.TrimSpace(os.Getenv("RENDER")) != ""
	if !production {
		return nil
	}
	required := []string{"DATABASE_URL", "SESSION_SECRET", "ADMIN_USER", "ADMIN_PASSWORD"}
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return fmt.Errorf("production thiếu biến môi trường %s", key)
		}
	}
	if len(os.Getenv("SESSION_SECRET")) < 32 {
		return fmt.Errorf("SESSION_SECRET production phải có ít nhất 32 ký tự")
	}
	if os.Getenv("ADMIN_PASSWORD") == "admin" {
		return fmt.Errorf("ADMIN_PASSWORD production không được dùng mật khẩu mặc định admin")
	}
	return nil
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(r.URL.Path, "/web/icons/") {
			w.Header().Set("Cache-Control", "public, max-age=604800")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		next.ServeHTTP(w, r)
	})
}

func cacheUploads(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Uploaded assets use generated/hash-like file names. A replacement gets a new
		// URL, so the browser/CDN can safely cache the old URL for a long time.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}
func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	w.Header().Del("Content-Length")
	if w.Header().Get("Content-Type") == "" && len(p) > 0 {
		w.Header().Set("Content-Type", http.DetectContentType(p))
	}
	return w.writer.Write(p)
}

var gzipPool = sync.Pool{New: func() any {
	gz, _ := gzip.NewWriterLevel(io.Discard, 5)
	return gz
}}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		alreadyCompressed := ext == ".webp" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".zip" || ext == ".pdf" || ext == ".woff2"
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || strings.HasPrefix(r.URL.Path, "/uploads/") || alreadyCompressed {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			_ = gz.Close()
			gz.Reset(io.Discard)
			gzipPool.Put(gz)
		}()
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, writer: gz}, r)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
