package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"sotaysinhvien/internal/model"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct{ db *sql.DB }

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Keep a single SQLite connection so per-connection PRAGMAs stay consistent
	// and local writes do not compete for the database lock.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;`); err != nil {
		db.Close()
		return nil, err
	}
	r := &SQLiteRepository{db: db}
	if err := r.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := r.seed(); err != nil {
		db.Close()
		return nil, err
	}
	_, _ = db.Exec(`PRAGMA optimize;`)
	return r, nil
}
func (r *SQLiteRepository) Close() error { return r.db.Close() }

func (r *SQLiteRepository) hasColumn(table, column string) bool {
	rows, err := r.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt any
		if rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk) == nil && name == column {
			return true
		}
	}
	return false
}

func (r *SQLiteRepository) migrate() error {
	if _, err := r.db.Exec(`
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS categories (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 name TEXT NOT NULL UNIQUE,
 slug TEXT NOT NULL UNIQUE,
 form_config TEXT NOT NULL DEFAULT '[]',
 show_company_card INTEGER NOT NULL DEFAULT 0,
 company_card_style TEXT NOT NULL DEFAULT 'full',
 sort_order INTEGER NOT NULL DEFAULT 0,
 audience_scope TEXT NOT NULL DEFAULT 'public',
 post_card_style TEXT NOT NULL DEFAULT 'normal',
 document_formats TEXT NOT NULL DEFAULT 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip'
);
CREATE TABLE IF NOT EXISTS posts (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 title TEXT NOT NULL,
 summary TEXT NOT NULL DEFAULT '',
 content TEXT NOT NULL,
 category_id INTEGER NOT NULL,
 meta_json TEXT NOT NULL DEFAULT '{}',
 author_id INTEGER NOT NULL DEFAULT 0,
 is_pinned INTEGER NOT NULL DEFAULT 0,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE RESTRICT
);
CREATE TABLE IF NOT EXISTS users (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 name TEXT NOT NULL,
 email TEXT NOT NULL UNIQUE,
 username TEXT UNIQUE,
 password_hash TEXT NOT NULL DEFAULT '',
 provider TEXT NOT NULL DEFAULT 'local',
 google_sub TEXT UNIQUE,
 avatar_url TEXT NOT NULL DEFAULT '',
 is_admin INTEGER NOT NULL DEFAULT 0,
 is_verified INTEGER NOT NULL DEFAULT 0,
 verification_type TEXT NOT NULL DEFAULT '',
 profile_role TEXT NOT NULL DEFAULT '',
 school TEXT NOT NULL DEFAULT '',
 major TEXT NOT NULL DEFAULT '',
 student_id TEXT NOT NULL DEFAULT '',
 phone TEXT NOT NULL DEFAULT '',
 phone_verified INTEGER NOT NULL DEFAULT 0,
 employer_company TEXT NOT NULL DEFAULT '',
 employer_tax_code TEXT NOT NULL DEFAULT '',
 employer_representative TEXT NOT NULL DEFAULT '',
 employer_website TEXT NOT NULL DEFAULT '',
 landlord_name TEXT NOT NULL DEFAULT '',
 landlord_address TEXT NOT NULL DEFAULT '',
 landlord_phone TEXT NOT NULL DEFAULT '',
 landlord_legal_info TEXT NOT NULL DEFAULT '',
 account_status TEXT NOT NULL DEFAULT 'active',
 locked_until TEXT NOT NULL DEFAULT '',
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS filter_values (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 category_id INTEGER NOT NULL,
 field_key TEXT NOT NULL,
 value TEXT NOT NULL,
 normalized_value TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'pending',
 usage_count INTEGER NOT NULL DEFAULT 1,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE(category_id, field_key, normalized_value),
 FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_filter_values_lookup ON filter_values(category_id, field_key, status, usage_count DESC);

CREATE TABLE IF NOT EXISTS advertisements (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 title TEXT NOT NULL,
 description TEXT NOT NULL DEFAULT '',
 image_url TEXT NOT NULL DEFAULT '',
 link_url TEXT NOT NULL DEFAULT '',
 position TEXT NOT NULL,
 active INTEGER NOT NULL DEFAULT 1,
 sort_order INTEGER NOT NULL DEFAULT 10,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS drafts (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 category_id INTEGER NOT NULL DEFAULT 0,
 title TEXT NOT NULL DEFAULT '',
 content TEXT NOT NULL DEFAULT '',
 meta_json TEXT NOT NULL DEFAULT '{}',
 author_id INTEGER NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS post_votes (
 post_id INTEGER NOT NULL,
 user_id INTEGER NOT NULL,
 value INTEGER NOT NULL CHECK(value IN(-1,1)),
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(post_id,user_id),
 FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS comments (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 post_id INTEGER NOT NULL,
 parent_id INTEGER NOT NULL DEFAULT 0,
 user_id INTEGER NOT NULL,
 content TEXT NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS saved_posts (
 post_id INTEGER NOT NULL,
 user_id INTEGER NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(post_id,user_id),
 FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS post_reports (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 post_id INTEGER NOT NULL,
 reporter_id INTEGER NOT NULL,
 reason TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'pending',
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE(post_id, reporter_id),
 FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
 FOREIGN KEY(reporter_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS verification_requests (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 user_id INTEGER NOT NULL,
 type TEXT NOT NULL,
 info TEXT NOT NULL DEFAULT '',
 status TEXT NOT NULL DEFAULT 'pending',
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS user_blocks (
 blocker_id INTEGER NOT NULL,
 blocked_id INTEGER NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY(blocker_id, blocked_id),
 FOREIGN KEY(blocker_id) REFERENCES users(id) ON DELETE CASCADE,
 FOREIGN KEY(blocked_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_category_created ON posts(category_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts(author_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_pinned_created ON posts(is_pinned, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_ads_position_active_sort ON advertisements(position, active, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_drafts_author_updated ON drafts(author_id, updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_saved_user_created ON saved_posts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_saved_user_post ON saved_posts(user_id, post_id);
CREATE INDEX IF NOT EXISTS idx_verification_user_type_status ON verification_requests(user_id, type, status, id DESC);
CREATE INDEX IF NOT EXISTS idx_reports_status_created ON post_reports(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_verification_user_id ON verification_requests(user_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_verification_status_id ON verification_requests(status, id DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone) WHERE TRIM(phone) <> '';`); err != nil {
		return err
	}

	if !r.hasColumn("categories", "sort_order") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
		_, _ = r.db.Exec(`UPDATE categories SET sort_order=id*10 WHERE sort_order=0`)
	}
	if !r.hasColumn("categories", "form_config") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN form_config TEXT NOT NULL DEFAULT '[]'`); err != nil {
			return err
		}
	}
	if !r.hasColumn("posts", "meta_json") {
		if _, err := r.db.Exec(`ALTER TABLE posts ADD COLUMN meta_json TEXT NOT NULL DEFAULT '{}'`); err != nil {
			return err
		}
	}
	if !r.hasColumn("posts", "author_id") {
		if _, err := r.db.Exec(`ALTER TABLE posts ADD COLUMN author_id INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	if !r.hasColumn("posts", "is_pinned") {
		if _, err := r.db.Exec(`ALTER TABLE posts ADD COLUMN is_pinned INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "avatar_url") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN avatar_url TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "is_admin") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "is_verified") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN is_verified INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "verification_type") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN verification_type TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "profile_role") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN profile_role TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "school") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN school TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "major") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN major TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "student_id") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN student_id TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "phone") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN phone TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "phone_verified") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN phone_verified INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	for _, migration := range []struct{ name, sql string }{
		{"employer_company", `ALTER TABLE users ADD COLUMN employer_company TEXT NOT NULL DEFAULT ''`},
		{"employer_tax_code", `ALTER TABLE users ADD COLUMN employer_tax_code TEXT NOT NULL DEFAULT ''`},
		{"employer_representative", `ALTER TABLE users ADD COLUMN employer_representative TEXT NOT NULL DEFAULT ''`},
		{"employer_website", `ALTER TABLE users ADD COLUMN employer_website TEXT NOT NULL DEFAULT ''`},
		{"landlord_name", `ALTER TABLE users ADD COLUMN landlord_name TEXT NOT NULL DEFAULT ''`},
		{"landlord_address", `ALTER TABLE users ADD COLUMN landlord_address TEXT NOT NULL DEFAULT ''`},
		{"landlord_phone", `ALTER TABLE users ADD COLUMN landlord_phone TEXT NOT NULL DEFAULT ''`},
		{"landlord_legal_info", `ALTER TABLE users ADD COLUMN landlord_legal_info TEXT NOT NULL DEFAULT ''`},
	} {
		if !r.hasColumn("users", migration.name) {
			if _, err := r.db.Exec(migration.sql); err != nil {
				return err
			}
		}
	}
	if !r.hasColumn("users", "account_status") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN account_status TEXT NOT NULL DEFAULT 'active'`); err != nil {
			return err
		}
	}
	if !r.hasColumn("users", "locked_until") {
		if _, err := r.db.Exec(`ALTER TABLE users ADD COLUMN locked_until TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	if !r.hasColumn("categories", "show_company_card") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN show_company_card INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
		_, _ = r.db.Exec(`UPDATE categories SET show_company_card=1 WHERE slug='viec-lam'`)
	}
	if !r.hasColumn("categories", "company_card_style") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN company_card_style TEXT NOT NULL DEFAULT 'full'`); err != nil {
			return err
		}
	}
	if !r.hasColumn("categories", "audience_scope") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN audience_scope TEXT NOT NULL DEFAULT 'public'`); err != nil {
			return err
		}
	}
	if !r.hasColumn("categories", "post_card_style") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN post_card_style TEXT NOT NULL DEFAULT 'normal'`); err != nil {
			return err
		}
		_, _ = r.db.Exec(`UPDATE categories SET post_card_style='document' WHERE slug='hoc-tap'`)
	}
	if !r.hasColumn("categories", "document_formats") {
		if _, err := r.db.Exec(`ALTER TABLE categories ADD COLUMN document_formats TEXT NOT NULL DEFAULT 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip'`); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLiteRepository) seed() error {
	// Không tự tạo chuyên mục hoặc bài viết từ source.
	// Toàn bộ nội dung mẫu được lưu trực tiếp trong file student_handbook.db.
	// Vì vậy, khi xóa file DB rồi chạy lại ứng dụng, migrate() chỉ tạo các bảng rỗng;
	// chuyên mục và bài viết cũng biến mất đúng theo cơ chế dữ liệu của SQLite.
	return nil
}

func normalizeFilterValue(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func (r *SQLiteRepository) UpsertFilterValue(categoryID int, fieldKey, value string, approved bool) error {
	value = strings.TrimSpace(value)
	norm := normalizeFilterValue(value)
	if categoryID <= 0 || strings.TrimSpace(fieldKey) == "" || norm == "" {
		return nil
	}
	status := "pending"
	if approved {
		status = "approved"
	}
	_, err := r.db.Exec(`INSERT INTO filter_values(category_id,field_key,value,normalized_value,status,usage_count,updated_at)
VALUES(?,?,?,?,?,1,CURRENT_TIMESTAMP)
ON CONFLICT(category_id,field_key,normalized_value) DO UPDATE SET
 usage_count=filter_values.usage_count+1,
 value=CASE WHEN filter_values.status='approved' THEN filter_values.value ELSE excluded.value END,
 status=CASE WHEN filter_values.status='approved' THEN 'approved' ELSE excluded.status END,
 updated_at=CURRENT_TIMESTAMP`, categoryID, strings.TrimSpace(fieldKey), value, norm, status)
	return err
}

func (r *SQLiteRepository) ListFilterValues(categoryID int, fieldKey, status string) ([]model.FilterValue, error) {
	query := `SELECT fv.id,fv.category_id,c.name,fv.field_key,fv.value,fv.normalized_value,fv.status,fv.usage_count,fv.created_at
FROM filter_values fv JOIN categories c ON c.id=fv.category_id WHERE 1=1`
	args := []any{}
	if categoryID > 0 {
		query += ` AND fv.category_id=?`
		args = append(args, categoryID)
	}
	if strings.TrimSpace(fieldKey) != "" {
		query += ` AND fv.field_key=?`
		args = append(args, strings.TrimSpace(fieldKey))
	}
	if strings.TrimSpace(status) != "" {
		query += ` AND fv.status=?`
		args = append(args, strings.TrimSpace(status))
	}
	query += ` ORDER BY CASE fv.status WHEN 'pending' THEN 0 ELSE 1 END, fv.usage_count DESC, fv.value COLLATE NOCASE`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.FilterValue{}
	for rows.Next() {
		var v model.FilterValue
		if err := rows.Scan(&v.ID, &v.CategoryID, &v.CategoryName, &v.FieldKey, &v.Value, &v.Normalized, &v.Status, &v.UsageCount, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) ReviewFilterValue(id int, action, newValue string) error {
	if id <= 0 {
		return errors.New("id dữ liệu lọc không hợp lệ")
	}
	switch action {
	case "approve":
		res, err := r.db.Exec(`UPDATE filter_values SET status='approved',updated_at=CURRENT_TIMESTAMP WHERE id=?`, id)
		if err != nil {
			return err
		}
		return requireAffected(res, "dữ liệu bộ lọc")
	case "rename":
		newValue = strings.TrimSpace(newValue)
		if newValue == "" {
			return errors.New("giá trị mới không được để trống")
		}
		norm := normalizeFilterValue(newValue)
		var categoryID int
		var fieldKey string
		var usage int
		if err := r.db.QueryRow(`SELECT category_id,field_key,usage_count FROM filter_values WHERE id=?`, id).Scan(&categoryID, &fieldKey, &usage); err != nil {
			return err
		}
		tx, err := r.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var oldValue string
		if err = tx.QueryRow(`SELECT value FROM filter_values WHERE id=?`, id).Scan(&oldValue); err != nil {
			return err
		}
		jsonPath := "$." + fieldKey
		if _, err = tx.Exec(`UPDATE posts SET meta_json=json_set(meta_json, ?, ?) WHERE category_id=? AND lower(trim(COALESCE(json_extract(meta_json, ?),'')))=lower(trim(?))`, jsonPath, newValue, categoryID, jsonPath, oldValue); err != nil {
			return err
		}
		var targetID int
		err = tx.QueryRow(`SELECT id FROM filter_values WHERE category_id=? AND field_key=? AND normalized_value=? AND id<>?`, categoryID, fieldKey, norm, id).Scan(&targetID)
		if err == nil {
			if _, err = tx.Exec(`UPDATE filter_values SET usage_count=usage_count+?,status='approved',updated_at=CURRENT_TIMESTAMP WHERE id=?`, usage, targetID); err != nil {
				return err
			}
			if _, err = tx.Exec(`DELETE FROM filter_values WHERE id=?`, id); err != nil {
				return err
			}
			return tx.Commit()
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err = tx.Exec(`UPDATE filter_values SET value=?,normalized_value=?,status='approved',updated_at=CURRENT_TIMESTAMP WHERE id=?`, newValue, norm, id); err != nil {
			return err
		}
		return tx.Commit()
	case "delete":
		res, err := r.db.Exec(`DELETE FROM filter_values WHERE id=?`, id)
		if err != nil {
			return err
		}
		return requireAffected(res, "dữ liệu bộ lọc")
	default:
		return errors.New("thao tác dữ liệu lọc không hợp lệ")
	}
}

func (r *SQLiteRepository) ListCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`SELECT id,name,slug,COALESCE(form_config,'[]'),COALESCE(show_company_card,0),COALESCE(company_card_style,'full'),COALESCE(sort_order,0),COALESCE(audience_scope,'public'),COALESCE(post_card_style,'normal'),COALESCE(document_formats,'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip') FROM categories ORDER BY CASE WHEN sort_order=0 THEN id*10 ELSE sort_order END,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.FormConfig, &c.ShowCompanyCard, &c.CompanyCardStyle, &c.SortOrder, &c.AudienceScope, &c.PostCardStyle, &c.DocumentFormats); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func scanPostRows(rows *sql.Rows) ([]model.Post, error) {
	var out []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Summary, &p.Content, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.MetaJSON, &p.AuthorID, &p.IsPinned, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *SQLiteRepository) ListPosts(q, category string) ([]model.Post, error) {
	query := `SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE 1=1`
	args := []any{}
	if q != "" {
		query += ` AND (LOWER(p.title) LIKE ? OR LOWER(p.summary) LIKE ? OR LOWER(p.content) LIKE ?)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like, like)
	}
	if category != "" {
		query += ` AND c.slug=?`
		args = append(args, category)
	}
	query += ` ORDER BY p.created_at DESC,p.id DESC`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}
func (r *SQLiteRepository) ListPostsLimited(q, category string, limit int) ([]model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	query := `SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE 1=1`
	args := []any{}
	if q != "" {
		query += ` AND (LOWER(p.title) LIKE ? OR LOWER(p.summary) LIKE ? OR LOWER(p.content) LIKE ?)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like, like)
	}
	if category != "" {
		query += ` AND c.slug=?`
		args = append(args, category)
	}
	query += ` ORDER BY p.created_at DESC,p.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}
func (r *SQLiteRepository) ListTrendingPosts(limit int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at
		FROM posts p
		JOIN categories c ON c.id=p.category_id
		LEFT JOIN post_votes v ON v.post_id=p.id
		GROUP BY p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,p.meta_json,p.author_id,p.created_at
		ORDER BY COALESCE(SUM(v.value),0) DESC, COALESCE(SUM(CASE WHEN v.value=1 THEN 1 ELSE 0 END),0) DESC, p.created_at DESC, p.id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func (r *SQLiteRepository) ListPostsByAuthor(authorID int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at
		FROM posts p JOIN categories c ON c.id=p.category_id
		WHERE p.author_id=? ORDER BY p.created_at DESC,p.id DESC`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Summary, &p.Content, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.MetaJSON, &p.AuthorID, &p.IsPinned, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) ListPinnedPosts(category string, limit int) ([]model.Post, error) {
	query := `SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE COALESCE(p.is_pinned,0)=1`
	args := []any{}
	if category != "" {
		query += ` AND c.slug=?`
		args = append(args, category)
	}
	query += ` ORDER BY p.created_at DESC,p.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}
func (r *SQLiteRepository) ListTodayPosts(category string, limit int) ([]model.Post, error) {
	query := `SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE p.created_at >= datetime('now','localtime','start of day','utc') AND p.created_at < datetime('now','localtime','start of day','+1 day','utc')`
	args := []any{}
	if category != "" {
		query += ` AND c.slug=?`
		args = append(args, category)
	}
	query += ` ORDER BY p.created_at DESC,p.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}
func (r *SQLiteRepository) SetPostPinned(id int, pinned bool) error {
	res, err := r.db.Exec(`UPDATE posts SET is_pinned=? WHERE id=?`, pinned, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "bài viết")
}
func (r *SQLiteRepository) GetPost(id int) (model.Post, error) {
	var p model.Post
	err := r.db.QueryRow(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE p.id=?`, id).Scan(&p.ID, &p.Title, &p.Summary, &p.Content, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.MetaJSON, &p.AuthorID, &p.IsPinned, &p.CreatedAt)
	return p, err
}
func (r *SQLiteRepository) SavePost(p model.Post) error {
	if p.ID > 0 {
		res, err := r.db.Exec(`UPDATE posts SET title=?,summary=?,content=?,category_id=?,meta_json=?,author_id=?,is_pinned=? WHERE id=?`, p.Title, p.Summary, p.Content, p.CategoryID, p.MetaJSON, p.AuthorID, p.IsPinned, p.ID)
		if err != nil {
			return err
		}
		return requireAffected(res, "bài viết")
	}
	_, err := r.db.Exec(`INSERT INTO posts(title,summary,content,category_id,meta_json,author_id,is_pinned) VALUES(?,?,?,?,?,?,?)`, p.Title, p.Summary, p.Content, p.CategoryID, p.MetaJSON, p.AuthorID, p.IsPinned)
	return err
}
func (r *SQLiteRepository) DeletePost(id int) error {
	res, err := r.db.Exec(`DELETE FROM posts WHERE id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "bài viết")
}
func (r *SQLiteRepository) SaveCategory(c model.Category) error {
	_, err := r.db.Exec(`INSERT INTO categories(name,slug,form_config,show_company_card,company_card_style,sort_order,audience_scope,post_card_style,document_formats) VALUES(?,?,?,?,?,COALESCE(NULLIF(?,0),(SELECT COALESCE(MAX(sort_order),0)+10 FROM categories)),?,?,?)`, c.Name, c.Slug, c.FormConfig, c.ShowCompanyCard, c.CompanyCardStyle, c.SortOrder, c.AudienceScope, c.PostCardStyle, c.DocumentFormats)
	return err
}
func (r *SQLiteRepository) UpdateCategoryConfig(c model.Category) error {
	res, err := r.db.Exec(`UPDATE categories SET form_config=?, show_company_card=?, company_card_style=?, audience_scope=?, post_card_style=?, document_formats=? WHERE id=?`, c.FormConfig, c.ShowCompanyCard, c.CompanyCardStyle, c.AudienceScope, c.PostCardStyle, c.DocumentFormats, c.ID)
	if err != nil {
		return err
	}
	return requireAffected(res, "chuyên mục")
}
func (r *SQLiteRepository) UpdateCategoryMeta(c model.Category) error {
	res, err := r.db.Exec(`UPDATE categories SET name=?, audience_scope=?, post_card_style=?, show_company_card=?, company_card_style=?, document_formats=? WHERE id=?`, c.Name, c.AudienceScope, c.PostCardStyle, c.ShowCompanyCard, c.CompanyCardStyle, c.DocumentFormats, c.ID)
	if err != nil {
		return err
	}
	return requireAffected(res, "chuyên mục")
}
func (r *SQLiteRepository) ReorderCategories(ids []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if id <= 0 {
			continue
		}
		res, err := tx.Exec(`UPDATE categories SET sort_order=? WHERE id=?`, (i+1)*10, id)
		if err != nil {
			return err
		}
		if err := requireAffected(res, "chuyên mục"); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *SQLiteRepository) DeleteCategory(id int) error {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE category_id=?`, id).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return errors.New("category in use")
	}
	res, err := r.db.Exec(`DELETE FROM categories WHERE id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "chuyên mục")
}
func (r *SQLiteRepository) VotePost(postID, userID, value int) error {
	_, err := r.db.Exec(`INSERT INTO post_votes(post_id,user_id,value) VALUES(?,?,?) ON CONFLICT(post_id,user_id) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, postID, userID, value)
	return err
}
func (r *SQLiteRepository) GetPostStats(postID int) (model.PostStats, error) {
	var st model.PostStats
	err := r.db.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN value=1 THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN value=-1 THEN 1 ELSE 0 END),0),
		COALESCE(SUM(value),0),
		(SELECT COUNT(*) FROM comments WHERE post_id=?)
		FROM post_votes WHERE post_id=?`, postID, postID).Scan(&st.Upvotes, &st.Downvotes, &st.Score, &st.Comments)
	return st, err
}
func (r *SQLiteRepository) GetPostStatsBatch(postIDs []int) (map[int]model.PostStats, error) {
	result := make(map[int]model.PostStats, len(postIDs))
	if len(postIDs) == 0 {
		return result, nil
	}
	marks := strings.TrimRight(strings.Repeat("?,", len(postIDs)), ",")
	args := make([]any, 0, len(postIDs)*2)
	for _, id := range postIDs {
		args = append(args, id)
	}
	for _, id := range postIDs {
		args = append(args, id)
	}
	query := fmt.Sprintf(`SELECT post_id,
		COALESCE(SUM(upvotes),0),COALESCE(SUM(downvotes),0),COALESCE(SUM(score),0),COALESCE(SUM(comments),0)
		FROM (
			SELECT post_id, CASE WHEN value=1 THEN 1 ELSE 0 END upvotes,
			CASE WHEN value=-1 THEN 1 ELSE 0 END downvotes, value score, 0 comments
			FROM post_votes WHERE post_id IN (%s)
			UNION ALL
			SELECT post_id,0,0,0,1 FROM comments WHERE post_id IN (%s)
		) stats GROUP BY post_id`, marks, marks)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var st model.PostStats
		if err := rows.Scan(&id, &st.Upvotes, &st.Downvotes, &st.Score, &st.Comments); err != nil {
			return nil, err
		}
		result[id] = st
	}
	return result, rows.Err()
}
func (r *SQLiteRepository) GetUserPostVote(postID, userID int) (int, error) {
	var v int
	err := r.db.QueryRow(`SELECT COALESCE((SELECT value FROM post_votes WHERE post_id=? AND user_id=?),0)`, postID, userID).Scan(&v)
	return v, err
}
func (r *SQLiteRepository) CreateComment(c model.Comment) error {
	_, err := r.db.Exec(`INSERT INTO comments(post_id,parent_id,user_id,content) VALUES(?,?,?,?)`, c.PostID, c.ParentID, c.UserID, c.Content)
	return err
}
func (r *SQLiteRepository) ListComments(postID int) ([]model.Comment, error) {
	rows, err := r.db.Query(`SELECT c.id,c.post_id,c.parent_id,c.user_id,u.name,c.content,c.created_at FROM comments c JOIN users u ON u.id=c.user_id WHERE c.post_id=? ORDER BY c.created_at ASC,c.id ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.UserID, &c.UserName, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) ListAds(position string, activeOnly bool) ([]model.Advertisement, error) {
	query := `SELECT id,title,description,image_url,link_url,position,active,sort_order,created_at FROM advertisements WHERE 1=1`
	args := []any{}
	if position != "" {
		query += ` AND position=?`
		args = append(args, position)
	}
	if activeOnly {
		query += ` AND active=1`
	}
	query += ` ORDER BY sort_order ASC,id DESC`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Advertisement
	for rows.Next() {
		var a model.Advertisement
		var active int
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.LinkURL, &a.Position, &active, &a.SortOrder, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Active = active == 1
		out = append(out, a)
	}
	return out, rows.Err()
}
func (r *SQLiteRepository) GetAd(id int) (model.Advertisement, error) {
	var a model.Advertisement
	var active int
	err := r.db.QueryRow(`SELECT id,title,description,image_url,link_url,position,active,sort_order,created_at FROM advertisements WHERE id=?`, id).Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.LinkURL, &a.Position, &active, &a.SortOrder, &a.CreatedAt)
	a.Active = active == 1
	return a, err
}
func (r *SQLiteRepository) SaveAd(a model.Advertisement) error {
	active := 0
	if a.Active {
		active = 1
	}
	if a.ID > 0 {
		res, err := r.db.Exec(`UPDATE advertisements SET title=?,description=?,image_url=?,link_url=?,position=?,active=?,sort_order=? WHERE id=?`, a.Title, a.Description, a.ImageURL, a.LinkURL, a.Position, active, a.SortOrder, a.ID)
		if err != nil {
			return err
		}
		return requireAffected(res, "quảng cáo")
	}
	_, err := r.db.Exec(`INSERT INTO advertisements(title,description,image_url,link_url,position,active,sort_order) VALUES(?,?,?,?,?,?,?)`, a.Title, a.Description, a.ImageURL, a.LinkURL, a.Position, active, a.SortOrder)
	return err
}
func (r *SQLiteRepository) ReorderAds(ids []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		if id <= 0 {
			continue
		}
		res, err := tx.Exec(`UPDATE advertisements SET sort_order=? WHERE id=?`, (i+1)*10, id)
		if err != nil {
			return err
		}
		if err := requireAffected(res, "quảng cáo"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *SQLiteRepository) DeleteAd(id int) error {
	res, err := r.db.Exec(`DELETE FROM advertisements WHERE id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "quảng cáo")
}
func (r *SQLiteRepository) SaveDraft(d model.Draft) (int, error) {
	if d.ID > 0 {
		res, err := r.db.Exec(`UPDATE drafts SET category_id=?,title=?,content=?,meta_json=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND author_id=?`, d.CategoryID, d.Title, d.Content, d.MetaJSON, d.ID, d.AuthorID)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return 0, errors.New("draft not found")
		}
		return d.ID, nil
	}
	res, err := r.db.Exec(`INSERT INTO drafts(category_id,title,content,meta_json,author_id) VALUES(?,?,?,?,?)`, d.CategoryID, d.Title, d.Content, d.MetaJSON, d.AuthorID)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}
func (r *SQLiteRepository) GetDraft(id, authorID int) (model.Draft, error) {
	var d model.Draft
	err := r.db.QueryRow(`SELECT d.id,d.category_id,COALESCE(c.name,''),d.title,d.content,d.meta_json,d.author_id,d.created_at,d.updated_at FROM drafts d LEFT JOIN categories c ON c.id=d.category_id WHERE d.id=? AND d.author_id=?`, id, authorID).Scan(&d.ID, &d.CategoryID, &d.CategoryName, &d.Title, &d.Content, &d.MetaJSON, &d.AuthorID, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}
func (r *SQLiteRepository) ListDrafts(authorID int) ([]model.Draft, error) {
	rows, err := r.db.Query(`SELECT d.id,d.category_id,COALESCE(c.name,''),d.title,d.content,d.meta_json,d.author_id,d.created_at,d.updated_at FROM drafts d LEFT JOIN categories c ON c.id=d.category_id WHERE d.author_id=? ORDER BY d.updated_at DESC,d.id DESC`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Draft
	for rows.Next() {
		var d model.Draft
		if err := rows.Scan(&d.ID, &d.CategoryID, &d.CategoryName, &d.Title, &d.Content, &d.MetaJSON, &d.AuthorID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (r *SQLiteRepository) DeleteDraft(id, authorID int) error {
	res, err := r.db.Exec(`DELETE FROM drafts WHERE id=? AND author_id=?`, id, authorID)
	if err != nil {
		return err
	}
	return requireAffected(res, "bản nháp")
}
func (r *SQLiteRepository) IsPostSaved(postID, userID int) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM saved_posts WHERE post_id=? AND user_id=?`, postID, userID).Scan(&n)
	return n > 0, err
}
func (r *SQLiteRepository) ListSavedPostIDs(userID int, postIDs []int) (map[int]bool, error) {
	result := make(map[int]bool, len(postIDs))
	if userID <= 0 || len(postIDs) == 0 {
		return result, nil
	}
	marks := strings.TrimRight(strings.Repeat("?,", len(postIDs)), ",")
	args := make([]any, 0, len(postIDs)+1)
	args = append(args, userID)
	for _, id := range postIDs {
		args = append(args, id)
	}
	rows, err := r.db.Query(`SELECT post_id FROM saved_posts WHERE user_id=? AND post_id IN (`+marks+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, rows.Err()
}
func (r *SQLiteRepository) SavePostForUser(postID, userID int) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO saved_posts(post_id,user_id) VALUES(?,?)`, postID, userID)
	return err
}
func (r *SQLiteRepository) UnsavePostForUser(postID, userID int) error {
	_, err := r.db.Exec(`DELETE FROM saved_posts WHERE post_id=? AND user_id=?`, postID, userID)
	return err
}
func (r *SQLiteRepository) ListSavedPosts(userID int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM saved_posts s JOIN posts p ON p.id=s.post_id JOIN categories c ON c.id=p.category_id WHERE s.user_id=? ORDER BY s.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func (r *SQLiteRepository) CreatePostReport(report model.PostReport) error {
	_, err := r.db.Exec(`INSERT INTO post_reports(post_id,reporter_id,reason,status) VALUES(?,?,?,'pending') ON CONFLICT(post_id,reporter_id) DO UPDATE SET reason=excluded.reason,status='pending',created_at=CURRENT_TIMESTAMP`, report.PostID, report.ReporterID, report.Reason)
	return err
}

func (r *SQLiteRepository) ListPostReports(status string) ([]model.PostReport, error) {
	query := `SELECT pr.id,pr.post_id,p.title,COALESCE(p.author_id,0),COALESCE(au.name,'Không rõ'),pr.reporter_id,COALESCE(ru.name,'Thành viên'),pr.reason,pr.status,pr.created_at FROM post_reports pr JOIN posts p ON p.id=pr.post_id LEFT JOIN users au ON au.id=p.author_id LEFT JOIN users ru ON ru.id=pr.reporter_id`
	args := []any{}
	if strings.TrimSpace(status) != "" && status != "all" {
		query += ` WHERE pr.status=?`
		args = append(args, status)
	}
	query += ` ORDER BY CASE pr.status WHEN 'pending' THEN 0 ELSE 1 END, pr.created_at DESC`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PostReport
	for rows.Next() {
		var item model.PostReport
		if err := rows.Scan(&item.ID, &item.PostID, &item.PostTitle, &item.PostAuthorID, &item.PostAuthor, &item.ReporterID, &item.ReporterName, &item.Reason, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) UpdatePostReportStatus(id int, status string) error {
	res, err := r.db.Exec(`UPDATE post_reports SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "báo cáo")
}
func (r *SQLiteRepository) BlockUser(blockerID, blockedID int) error {
	if blockerID == blockedID {
		return errors.New("không thể tự chặn chính mình")
	}
	_, err := r.db.Exec(`INSERT OR IGNORE INTO user_blocks(blocker_id,blocked_id) VALUES(?,?)`, blockerID, blockedID)
	return err
}

func (r *SQLiteRepository) IsUserBlocked(blockerID, blockedID int) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM user_blocks WHERE blocker_id=? AND blocked_id=?`, blockerID, blockedID).Scan(&n)
	return n > 0, err
}

func (r *SQLiteRepository) ListBlockedUserIDs(blockerID int) ([]int, error) {
	rows, err := r.db.Query(`SELECT blocked_id FROM user_blocks WHERE blocker_id=?`, blockerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) CreateUser(u model.User) (model.User, error) {
	var username any
	if strings.TrimSpace(u.Username) != "" {
		username = strings.TrimSpace(u.Username)
	}
	provider := strings.TrimSpace(u.Provider)
	if provider == "" {
		provider = "local"
	}
	res, err := r.db.Exec(`INSERT INTO users(name,email,username,password_hash,provider,is_admin,phone,phone_verified) VALUES(?,?,?,?,?,?,?,?)`, u.Name, strings.ToLower(u.Email), username, u.PasswordHash, provider, u.IsAdmin, strings.TrimSpace(u.Phone), u.PhoneVerified)
	if err != nil {
		return model.User{}, err
	}
	id, _ := res.LastInsertId()
	u.ID = int(id)
	u.Email = strings.ToLower(u.Email)
	u.Phone = strings.TrimSpace(u.Phone)
	u.Provider = provider
	return u, nil
}

func (r *SQLiteRepository) FindUserByLogin(login string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE LOWER(email)=LOWER(?) OR LOWER(username)=LOWER(?) OR phone=? LIMIT 1`, login, login, login).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *SQLiteRepository) FindUserByEmail(email string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE LOWER(email)=LOWER(?) LIMIT 1`, email).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *SQLiteRepository) FindUserByPhone(phone string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE phone=? LIMIT 1`, phone).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *SQLiteRepository) FindUserByID(id int) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE id=? LIMIT 1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *SQLiteRepository) FindUsersByIDs(ids []int) (map[int]model.User, error) {
	out := map[int]model.User{}
	if len(ids) == 0 {
		return out, nil
	}
	seen := map[int]bool{}
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			args = append(args, id)
		}
	}
	if len(args) == 0 {
		return out, nil
	}
	marks := strings.TrimRight(strings.Repeat("?,", len(args)), ",")
	rows, err := r.db.Query("SELECT "+userSelectColumns+" FROM users WHERE id IN ("+marks+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRows(rows)
}

func (r *SQLiteRepository) UpdateUserPasswordHash(id int, hash string) error {
	res, err := r.db.Exec(`UPDATE users SET password_hash=? WHERE id=?`, hash, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) UpdateUserProfile(id int, name, username, avatarURL, profileRole, school, major, studentID, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo string) error {
	res, err := r.db.Exec(`UPDATE users SET name=?, username=NULLIF(?,''), avatar_url=?, profile_role=?, school=?, major=?, student_id=?, phone_verified=CASE WHEN phone=? THEN phone_verified ELSE 0 END, phone=?, employer_company=?, employer_tax_code=?, employer_representative=?, employer_website=?, landlord_name=?, landlord_address=?, landlord_phone=?, landlord_legal_info=? WHERE id=?`, name, username, avatarURL, profileRole, school, major, studentID, phone, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) SetUserVerification(id int, verified bool, verificationType string) error {
	value := 0
	if verified {
		value = 1
	}
	res, err := r.db.Exec(`UPDATE users SET is_verified=?, verification_type=? WHERE id=?`, value, verificationType, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) SetUserTrustVerified(id int, verified bool) error {
	value := 0
	if verified {
		value = 1
	}
	res, err := r.db.Exec(`UPDATE users SET is_verified=? WHERE id=?`, value, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) SetUserRoleVerification(id int, verificationType string) error {
	res, err := r.db.Exec(`UPDATE users SET verification_type=? WHERE id=?`, verificationType, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) SetUserAdmin(id int, isAdmin bool) error {
	res, err := r.db.Exec(`UPDATE users SET is_admin=? WHERE id=?`, isAdmin, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) SearchUsers(query string, limit int) ([]model.User, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	like := "%" + strings.ToLower(query) + "%"
	var rows *sql.Rows
	var err error
	base := `SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users`
	if query == "" {
		rows, err = r.db.Query(base+` ORDER BY id DESC LIMIT ?`, limit)
	} else if id, convErr := strconv.Atoi(query); convErr == nil {
		rows, err = r.db.Query(base+` WHERE id=? OR LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(COALESCE(username,'')) LIKE ? OR phone LIKE ? ORDER BY CASE WHEN id=? THEN 0 ELSE 1 END,id DESC LIMIT ?`, id, like, like, like, like, id, limit)
	} else {
		rows, err = r.db.Query(base+` WHERE LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(COALESCE(username,'')) LIKE ? OR phone LIKE ? ORDER BY id DESC LIMIT ?`, like, like, like, like, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) SetUserAccountStatus(id int, status, lockedUntil string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	res, err := r.db.Exec(`UPDATE users SET account_status=?,locked_until=? WHERE id=?`, status, strings.TrimSpace(lockedUntil), id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *SQLiteRepository) DeleteUser(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM posts WHERE author_id=?`, id); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil {
		return err
	}
	if err := requireAffected(res, "tài khoản"); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *SQLiteRepository) CountAdmins() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin=1`).Scan(&n)
	return n, err
}

func (r *SQLiteRepository) UpsertGoogleUser(name, email, googleSub string) (model.User, error) {
	email = strings.ToLower(email)
	_, err := r.db.Exec(`INSERT INTO users(name,email,provider,google_sub) VALUES(?,?,'google',?) ON CONFLICT(email) DO UPDATE SET name=excluded.name, google_sub=excluded.google_sub`, name, email, googleSub)
	if err != nil {
		return model.User{}, err
	}
	return r.FindUserByEmail(email)
}

func (r *SQLiteRepository) CreateVerificationRequest(req model.VerificationRequest) error {
	_, err := r.db.Exec(`INSERT INTO verification_requests(user_id,type,info,status) VALUES(?,?,?,'pending')`, req.UserID, req.Type, req.Info)
	return err
}

func (r *SQLiteRepository) GetLatestVerificationRequest(userID int) (model.VerificationRequest, error) {
	var v model.VerificationRequest
	err := r.db.QueryRow(`SELECT vr.id,vr.user_id,u.name,u.email,vr.type,vr.info,vr.status,vr.created_at FROM verification_requests vr JOIN users u ON u.id=vr.user_id WHERE vr.user_id=? ORDER BY vr.id DESC LIMIT 1`, userID).Scan(&v.ID, &v.UserID, &v.UserName, &v.UserEmail, &v.Type, &v.Info, &v.Status, &v.CreatedAt)
	return v, err
}

func (r *SQLiteRepository) HasPendingVerificationRequest(userID int, requestType string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM verification_requests WHERE user_id=? AND type=? AND status='pending'`, userID, requestType).Scan(&count)
	return count > 0, err
}

func (r *SQLiteRepository) ListVerificationRequests(status string) ([]model.VerificationRequest, error) {
	q := `SELECT vr.id,vr.user_id,u.name,u.email,vr.type,vr.info,vr.status,vr.created_at FROM verification_requests vr JOIN users u ON u.id=vr.user_id`
	var rows *sql.Rows
	var err error
	if status != "" && status != "all" {
		rows, err = r.db.Query(q+` WHERE vr.status=? ORDER BY vr.id DESC`, status)
	} else {
		rows, err = r.db.Query(q + ` ORDER BY vr.id DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.VerificationRequest
	for rows.Next() {
		var v model.VerificationRequest
		if err := rows.Scan(&v.ID, &v.UserID, &v.UserName, &v.UserEmail, &v.Type, &v.Info, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *SQLiteRepository) ResolveVerificationRequest(id int, status string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int
	var typ string
	var currentStatus string
	if err := tx.QueryRow(`SELECT user_id,type,status FROM verification_requests WHERE id=?`, id).Scan(&userID, &typ, &currentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("yêu cầu xác thực không tồn tại")
		}
		return err
	}
	if currentStatus != "pending" {
		return errors.New("yêu cầu xác thực này đã được xử lý")
	}

	res, err := tx.Exec(`UPDATE verification_requests SET status=? WHERE id=? AND status='pending'`, status, id)
	if err != nil {
		return err
	}
	if err := requireAffected(res, "yêu cầu xác thực"); err != nil {
		return err
	}

	if status == "approved" {
		if typ == "phone" {
			res, err = tx.Exec(`UPDATE users SET phone_verified=1 WHERE id=?`, userID)
		} else {
			res, err = tx.Exec(`UPDATE users SET verification_type=? WHERE id=?`, typ, userID)
		}
		if err != nil {
			return err
		}
		if err := requireAffected(res, "tài khoản"); err != nil {
			return err
		}
	}
	if status == "rejected" && typ != "phone" {
		res, err = tx.Exec(`UPDATE users SET verification_type='' WHERE id=?`, userID)
		if err != nil {
			return err
		}
		if err := requireAffected(res, "tài khoản"); err != nil {
			return err
		}
	}
	return tx.Commit()
}
