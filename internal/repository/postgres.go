package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sotaysinhvien/internal/model"

	_ "github.com/lib/pq"
)

type pgDB struct{ *sql.DB }
type pgTx struct{ *sql.Tx }

func rebindPostgres(query string) string {
	var b strings.Builder
	n := 1
	inSingle := false
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' {
			if inSingle && i+1 < len(query) && query[i+1] == '\'' {
				b.WriteByte(ch)
				b.WriteByte(query[i+1])
				i++
				continue
			}
			inSingle = !inSingle
			b.WriteByte(ch)
			continue
		}
		if ch == '?' && !inSingle {
			b.WriteString("$")
			b.WriteString(strconv.Itoa(n))
			n++
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}
func pgArgs(args []any) []any {
	out := make([]any, len(args))
	for i, v := range args {
		switch x := v.(type) {
		case bool:
			if x {
				out[i] = 1
			} else {
				out[i] = 0
			}
		default:
			out[i] = v
		}
	}
	return out
}
func (d *pgDB) Exec(q string, args ...any) (sql.Result, error) {
	return d.DB.Exec(rebindPostgres(q), pgArgs(args)...)
}
func (d *pgDB) Query(q string, args ...any) (*sql.Rows, error) {
	return d.DB.Query(rebindPostgres(q), pgArgs(args)...)
}
func (d *pgDB) QueryRow(q string, args ...any) *sql.Row {
	return d.DB.QueryRow(rebindPostgres(q), pgArgs(args)...)
}
func (d *pgDB) Begin() (*pgTx, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &pgTx{tx}, nil
}
func (t *pgTx) Exec(q string, args ...any) (sql.Result, error) {
	return t.Tx.Exec(rebindPostgres(q), pgArgs(args)...)
}
func (t *pgTx) QueryRow(q string, args ...any) *sql.Row {
	return t.Tx.QueryRow(rebindPostgres(q), pgArgs(args)...)
}

type PostgresRepository struct{ db *pgDB }

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	// Keep enough warm connections for the parallel read path used by feed pages.
	// Supabase/Render latency is dominated by network RTT, so a small pool easily
	// becomes a bottleneck when independent reads are serialized.
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("kết nối Supabase PostgreSQL thất bại: %w", err)
	}
	r := &PostgresRepository{db: &pgDB{db}}
	if err := r.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := r.seed(); err != nil {
		db.Close()
		return nil, err
	}
	return r, nil
}
func (r *PostgresRepository) Close() error { return r.db.Close() }

func (r *PostgresRepository) migrate() error {
	_, err := r.db.Exec(`
CREATE TABLE IF NOT EXISTS categories (
 id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL UNIQUE, slug TEXT NOT NULL UNIQUE,
 form_config TEXT NOT NULL DEFAULT '[]', show_company_card INTEGER NOT NULL DEFAULT 0,
 company_card_style TEXT NOT NULL DEFAULT 'full', sort_order INTEGER NOT NULL DEFAULT 0,
 audience_scope TEXT NOT NULL DEFAULT 'public', post_card_style TEXT NOT NULL DEFAULT 'normal',
 document_formats TEXT NOT NULL DEFAULT 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip'
);
CREATE TABLE IF NOT EXISTS posts (
 id BIGSERIAL PRIMARY KEY, title TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', content TEXT NOT NULL,
 category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT, meta_json TEXT NOT NULL DEFAULT '{}',
 author_id BIGINT NOT NULL DEFAULT 0, is_pinned INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS users (
 id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL UNIQUE, username TEXT UNIQUE,
 password_hash TEXT NOT NULL DEFAULT '', provider TEXT NOT NULL DEFAULT 'local', google_sub TEXT UNIQUE,
 avatar_url TEXT NOT NULL DEFAULT '', is_admin INTEGER NOT NULL DEFAULT 0, is_verified INTEGER NOT NULL DEFAULT 0,
 verification_type TEXT NOT NULL DEFAULT '', profile_role TEXT NOT NULL DEFAULT '', school TEXT NOT NULL DEFAULT '',
 major TEXT NOT NULL DEFAULT '', student_id TEXT NOT NULL DEFAULT '', phone TEXT NOT NULL DEFAULT '', phone_verified INTEGER NOT NULL DEFAULT 0,
 employer_company TEXT NOT NULL DEFAULT '', employer_tax_code TEXT NOT NULL DEFAULT '', employer_representative TEXT NOT NULL DEFAULT '',
 employer_website TEXT NOT NULL DEFAULT '', landlord_name TEXT NOT NULL DEFAULT '', landlord_address TEXT NOT NULL DEFAULT '',
 landlord_phone TEXT NOT NULL DEFAULT '', landlord_legal_info TEXT NOT NULL DEFAULT '', account_status TEXT NOT NULL DEFAULT 'active',
 locked_until TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS filter_values (
 id BIGSERIAL PRIMARY KEY, category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
 field_key TEXT NOT NULL, value TEXT NOT NULL, normalized_value TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
 usage_count INTEGER NOT NULL DEFAULT 1, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(category_id,field_key,normalized_value)
);
CREATE TABLE IF NOT EXISTS advertisements (
 id BIGSERIAL PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', image_url TEXT NOT NULL DEFAULT '',
 link_url TEXT NOT NULL DEFAULT '', position TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 1, sort_order INTEGER NOT NULL DEFAULT 10,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS drafts (
 id BIGSERIAL PRIMARY KEY, category_id BIGINT NOT NULL DEFAULT 0, title TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '',
 meta_json TEXT NOT NULL DEFAULT '{}', author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS post_votes (
 post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 value INTEGER NOT NULL CHECK(value IN(-1,1)), created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(post_id,user_id)
);
CREATE TABLE IF NOT EXISTS comments (
 id BIGSERIAL PRIMARY KEY, post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE, parent_id BIGINT NOT NULL DEFAULT 0,
 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, content TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS saved_posts (
 post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(post_id,user_id)
);
CREATE TABLE IF NOT EXISTS post_reports (
 id BIGSERIAL PRIMARY KEY, post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
 reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, reason TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending',
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(post_id,reporter_id)
);
CREATE TABLE IF NOT EXISTS verification_requests (
 id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, type TEXT NOT NULL,
 info TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending', created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS user_blocks (
 blocker_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, blocked_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(blocker_id,blocked_id)
);
CREATE INDEX IF NOT EXISTS idx_filter_values_lookup ON filter_values(category_id,field_key,status,usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_posts_category_created ON posts(category_id,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts(author_id,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_pinned_created ON posts(is_pinned,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id,created_at,id);
CREATE INDEX IF NOT EXISTS idx_ads_position_active_sort ON advertisements(position,active,sort_order,id);
CREATE INDEX IF NOT EXISTS idx_drafts_author_updated ON drafts(author_id,updated_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS idx_saved_user_created ON saved_posts(user_id,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_saved_user_post ON saved_posts(user_id,post_id);
CREATE INDEX IF NOT EXISTS idx_verification_user_type_status ON verification_requests(user_id,type,status,id DESC);
CREATE INDEX IF NOT EXISTS idx_reports_status_created ON post_reports(status,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_verification_user_id ON verification_requests(user_id,id DESC);
CREATE INDEX IF NOT EXISTS idx_verification_status_id ON verification_requests(status,id DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone) WHERE TRIM(phone) <> '';
`)
	if err != nil {
		return err
	}

	// CREATE TABLE IF NOT EXISTS does not add new columns to an existing
	// PostgreSQL database. Keep older Supabase databases compatible with the
	// current category/form builder and account model.
	_, err = r.db.Exec(`
ALTER TABLE categories ADD COLUMN IF NOT EXISTS form_config TEXT NOT NULL DEFAULT '[]';
ALTER TABLE categories ADD COLUMN IF NOT EXISTS show_company_card INTEGER NOT NULL DEFAULT 0;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS company_card_style TEXT NOT NULL DEFAULT 'full';
ALTER TABLE categories ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS audience_scope TEXT NOT NULL DEFAULT 'public';
ALTER TABLE categories ADD COLUMN IF NOT EXISTS post_card_style TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE categories ADD COLUMN IF NOT EXISTS document_formats TEXT NOT NULL DEFAULT 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip';

ALTER TABLE posts ADD COLUMN IF NOT EXISTS meta_json TEXT NOT NULL DEFAULT '{}';
ALTER TABLE posts ADD COLUMN IF NOT EXISTS author_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE posts ADD COLUMN IF NOT EXISTS is_pinned INTEGER NOT NULL DEFAULT 0;

ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_type TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_role TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS school TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS major TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS student_id TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS employer_company TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS employer_tax_code TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS employer_representative TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS employer_website TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS landlord_name TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS landlord_address TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS landlord_phone TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS landlord_legal_info TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TEXT NOT NULL DEFAULT '';
`)
	return err
}
func (r *PostgresRepository) seed() error {
	// One server round-trip instead of 10 sequential Supabase calls during cold start.
	_, err := r.db.Exec(`
INSERT INTO categories(name,slug) VALUES
 ('Học tập','hoc-tap'),
 ('Học bổng','hoc-bong'),
 ('Sự kiện','su-kien'),
 ('Kỹ năng','ky-nang'),
 ('Việc làm','viec-lam'),
 ('Thông báo','thong-bao'),
 ('Confession','confession')
ON CONFLICT DO NOTHING;
UPDATE categories SET post_card_style='document' WHERE slug='hoc-tap' AND COALESCE(post_card_style,'normal')='normal';
UPDATE categories SET show_company_card=1,company_card_style='full' WHERE slug='viec-lam';
UPDATE categories SET audience_scope='same_school' WHERE slug='confession';`)
	return err
}

func (r *PostgresRepository) UpsertFilterValue(categoryID int, fieldKey, value string, approved bool) error {
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

func (r *PostgresRepository) ListFilterValues(categoryID int, fieldKey, status string) ([]model.FilterValue, error) {
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
	query += ` ORDER BY CASE fv.status WHEN 'pending' THEN 0 ELSE 1 END, fv.usage_count DESC, fv.value`
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

func (r *PostgresRepository) ReviewFilterValue(id int, action, newValue string) error {
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
		if _, err = tx.Exec(`UPDATE posts SET meta_json=jsonb_set(COALESCE(NULLIF(meta_json,''),'{}')::jsonb, ARRAY[?], to_jsonb(?::text), true)::text WHERE category_id=? AND lower(trim(COALESCE(COALESCE(NULLIF(meta_json,''),'{}')::jsonb ->> ?,'')))=lower(trim(?))`, fieldKey, newValue, categoryID, fieldKey, oldValue); err != nil {
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

func (r *PostgresRepository) ListCategories() ([]model.Category, error) {
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
func (r *PostgresRepository) ListPosts(q, category string) ([]model.Post, error) {
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
func (r *PostgresRepository) ListPostsLimited(q, category string, limit int) ([]model.Post, error) {
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
func (r *PostgresRepository) ListTrendingPosts(limit int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at
		FROM posts p
		JOIN categories c ON c.id=p.category_id
		LEFT JOIN post_votes v ON v.post_id=p.id
		GROUP BY p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,p.meta_json,p.author_id,p.is_pinned,p.created_at
		ORDER BY COALESCE(SUM(v.value),0) DESC, COALESCE(SUM(CASE WHEN v.value=1 THEN 1 ELSE 0 END),0) DESC, p.created_at DESC, p.id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func (r *PostgresRepository) ListPostsByAuthor(authorID int) ([]model.Post, error) {
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

func (r *PostgresRepository) ListPinnedPosts(category string, limit int) ([]model.Post, error) {
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
func (r *PostgresRepository) ListTodayPosts(category string, limit int) ([]model.Post, error) {
	query := `SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE p.created_at >= (date_trunc('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh') AT TIME ZONE 'Asia/Ho_Chi_Minh') AND p.created_at < ((date_trunc('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh') + interval '1 day') AT TIME ZONE 'Asia/Ho_Chi_Minh')`
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
func (r *PostgresRepository) SetPostPinned(id int, pinned bool) error {
	res, err := r.db.Exec(`UPDATE posts SET is_pinned=? WHERE id=?`, pinned, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "bài viết")
}
func (r *PostgresRepository) GetPost(id int) (model.Post, error) {
	var p model.Post
	err := r.db.QueryRow(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM posts p JOIN categories c ON c.id=p.category_id WHERE p.id=?`, id).Scan(&p.ID, &p.Title, &p.Summary, &p.Content, &p.CategoryID, &p.CategoryName, &p.CategorySlug, &p.MetaJSON, &p.AuthorID, &p.IsPinned, &p.CreatedAt)
	return p, err
}
func (r *PostgresRepository) SavePost(p model.Post) error {
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
func (r *PostgresRepository) DeletePost(id int) error {
	res, err := r.db.Exec(`DELETE FROM posts WHERE id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "bài viết")
}
func (r *PostgresRepository) SaveCategory(c model.Category) error {
	_, err := r.db.Exec(`INSERT INTO categories(name,slug,form_config,show_company_card,company_card_style,sort_order,audience_scope,post_card_style,document_formats) VALUES(?,?,?,?,?,COALESCE(NULLIF(?,0),(SELECT COALESCE(MAX(sort_order),0)+10 FROM categories)),?,?,?)`, c.Name, c.Slug, c.FormConfig, c.ShowCompanyCard, c.CompanyCardStyle, c.SortOrder, c.AudienceScope, c.PostCardStyle, c.DocumentFormats)
	return err
}
func (r *PostgresRepository) UpdateCategoryConfig(c model.Category) error {
	res, err := r.db.Exec(`UPDATE categories SET form_config=?, show_company_card=?, company_card_style=?, audience_scope=?, post_card_style=?, document_formats=? WHERE id=?`, c.FormConfig, c.ShowCompanyCard, c.CompanyCardStyle, c.AudienceScope, c.PostCardStyle, c.DocumentFormats, c.ID)
	if err != nil {
		return err
	}
	return requireAffected(res, "chuyên mục")
}
func (r *PostgresRepository) UpdateCategoryMeta(c model.Category) error {
	res, err := r.db.Exec(`UPDATE categories SET name=?, audience_scope=?, post_card_style=?, show_company_card=?, company_card_style=?, document_formats=? WHERE id=?`, c.Name, c.AudienceScope, c.PostCardStyle, c.ShowCompanyCard, c.CompanyCardStyle, c.DocumentFormats, c.ID)
	if err != nil {
		return err
	}
	return requireAffected(res, "chuyên mục")
}
func (r *PostgresRepository) ReorderCategories(ids []int) error {
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
func (r *PostgresRepository) DeleteCategory(id int) error {
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
func (r *PostgresRepository) VotePost(postID, userID, value int) error {
	_, err := r.db.Exec(`INSERT INTO post_votes(post_id,user_id,value) VALUES(?,?,?) ON CONFLICT(post_id,user_id) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, postID, userID, value)
	return err
}
func (r *PostgresRepository) GetPostStats(postID int) (model.PostStats, error) {
	var st model.PostStats
	err := r.db.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN value=1 THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN value=-1 THEN 1 ELSE 0 END),0),
		COALESCE(SUM(value),0),
		(SELECT COUNT(*) FROM comments WHERE post_id=?)
		FROM post_votes WHERE post_id=?`, postID, postID).Scan(&st.Upvotes, &st.Downvotes, &st.Score, &st.Comments)
	return st, err
}
func (r *PostgresRepository) GetPostStatsBatch(postIDs []int) (map[int]model.PostStats, error) {
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
func (r *PostgresRepository) GetUserPostVote(postID, userID int) (int, error) {
	var v int
	err := r.db.QueryRow(`SELECT COALESCE((SELECT value FROM post_votes WHERE post_id=? AND user_id=?),0)`, postID, userID).Scan(&v)
	return v, err
}
func (r *PostgresRepository) CreateComment(c model.Comment) error {
	_, err := r.db.Exec(`INSERT INTO comments(post_id,parent_id,user_id,content) VALUES(?,?,?,?)`, c.PostID, c.ParentID, c.UserID, c.Content)
	return err
}
func (r *PostgresRepository) ListComments(postID int) ([]model.Comment, error) {
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

func (r *PostgresRepository) ListAds(position string, activeOnly bool) ([]model.Advertisement, error) {
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
func (r *PostgresRepository) GetAd(id int) (model.Advertisement, error) {
	var a model.Advertisement
	var active int
	err := r.db.QueryRow(`SELECT id,title,description,image_url,link_url,position,active,sort_order,created_at FROM advertisements WHERE id=?`, id).Scan(&a.ID, &a.Title, &a.Description, &a.ImageURL, &a.LinkURL, &a.Position, &active, &a.SortOrder, &a.CreatedAt)
	a.Active = active == 1
	return a, err
}
func (r *PostgresRepository) SaveAd(a model.Advertisement) error {
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
func (r *PostgresRepository) ReorderAds(ids []int) error {
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

func (r *PostgresRepository) DeleteAd(id int) error {
	res, err := r.db.Exec(`DELETE FROM advertisements WHERE id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "quảng cáo")
}
func (r *PostgresRepository) SaveDraft(d model.Draft) (int, error) {
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
	var id int
	err := r.db.QueryRow(`INSERT INTO drafts(category_id,title,content,meta_json,author_id) VALUES(?,?,?,?,?) RETURNING id`, d.CategoryID, d.Title, d.Content, d.MetaJSON, d.AuthorID).Scan(&id)
	return id, err
}
func (r *PostgresRepository) GetDraft(id, authorID int) (model.Draft, error) {
	var d model.Draft
	err := r.db.QueryRow(`SELECT d.id,d.category_id,COALESCE(c.name,''),d.title,d.content,d.meta_json,d.author_id,d.created_at,d.updated_at FROM drafts d LEFT JOIN categories c ON c.id=d.category_id WHERE d.id=? AND d.author_id=?`, id, authorID).Scan(&d.ID, &d.CategoryID, &d.CategoryName, &d.Title, &d.Content, &d.MetaJSON, &d.AuthorID, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}
func (r *PostgresRepository) ListDrafts(authorID int) ([]model.Draft, error) {
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
func (r *PostgresRepository) DeleteDraft(id, authorID int) error {
	res, err := r.db.Exec(`DELETE FROM drafts WHERE id=? AND author_id=?`, id, authorID)
	if err != nil {
		return err
	}
	return requireAffected(res, "bản nháp")
}
func (r *PostgresRepository) IsPostSaved(postID, userID int) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM saved_posts WHERE post_id=? AND user_id=?`, postID, userID).Scan(&n)
	return n > 0, err
}
func (r *PostgresRepository) ListSavedPostIDs(userID int, postIDs []int) (map[int]bool, error) {
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
func (r *PostgresRepository) SavePostForUser(postID, userID int) error {
	_, err := r.db.Exec(`INSERT INTO saved_posts(post_id,user_id) VALUES(?,?) ON CONFLICT DO NOTHING`, postID, userID)
	return err
}
func (r *PostgresRepository) UnsavePostForUser(postID, userID int) error {
	_, err := r.db.Exec(`DELETE FROM saved_posts WHERE post_id=? AND user_id=?`, postID, userID)
	return err
}
func (r *PostgresRepository) ListSavedPosts(userID int) ([]model.Post, error) {
	rows, err := r.db.Query(`SELECT p.id,p.title,p.summary,p.content,p.category_id,c.name,c.slug,COALESCE(p.meta_json,'{}'),COALESCE(p.author_id,0),COALESCE(p.is_pinned,0),p.created_at FROM saved_posts s JOIN posts p ON p.id=s.post_id JOIN categories c ON c.id=p.category_id WHERE s.user_id=? ORDER BY s.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostRows(rows)
}

func (r *PostgresRepository) CreatePostReport(report model.PostReport) error {
	_, err := r.db.Exec(`INSERT INTO post_reports(post_id,reporter_id,reason,status) VALUES(?,?,?,'pending') ON CONFLICT(post_id,reporter_id) DO UPDATE SET reason=excluded.reason,status='pending',created_at=CURRENT_TIMESTAMP`, report.PostID, report.ReporterID, report.Reason)
	return err
}

func (r *PostgresRepository) ListPostReports(status string) ([]model.PostReport, error) {
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

func (r *PostgresRepository) UpdatePostReportStatus(id int, status string) error {
	res, err := r.db.Exec(`UPDATE post_reports SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "báo cáo")
}
func (r *PostgresRepository) BlockUser(blockerID, blockedID int) error {
	if blockerID == blockedID {
		return errors.New("không thể tự chặn chính mình")
	}
	_, err := r.db.Exec(`INSERT INTO user_blocks(blocker_id,blocked_id) VALUES(?,?) ON CONFLICT DO NOTHING`, blockerID, blockedID)
	return err
}

func (r *PostgresRepository) IsUserBlocked(blockerID, blockedID int) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM user_blocks WHERE blocker_id=? AND blocked_id=?`, blockerID, blockedID).Scan(&n)
	return n > 0, err
}

func (r *PostgresRepository) ListBlockedUserIDs(blockerID int) ([]int, error) {
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

func (r *PostgresRepository) CreateUser(u model.User) (model.User, error) {
	var username any
	if strings.TrimSpace(u.Username) != "" {
		username = strings.TrimSpace(u.Username)
	}
	provider := strings.TrimSpace(u.Provider)
	if provider == "" {
		provider = "local"
	}
	var id int
	err := r.db.QueryRow(`INSERT INTO users(name,email,username,password_hash,provider,is_admin,phone,phone_verified) VALUES(?,?,?,?,?,?,?,?) RETURNING id`, u.Name, strings.ToLower(u.Email), username, u.PasswordHash, provider, u.IsAdmin, strings.TrimSpace(u.Phone), u.PhoneVerified).Scan(&id)
	if err != nil {
		return model.User{}, err
	}
	u.ID = id
	u.Email = strings.ToLower(u.Email)
	u.Phone = strings.TrimSpace(u.Phone)
	u.Provider = provider
	return u, nil
}

func (r *PostgresRepository) FindUserByLogin(login string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE LOWER(email)=LOWER(?) OR LOWER(username)=LOWER(?) OR phone=? LIMIT 1`, login, login, login).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) FindUserByEmail(email string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE LOWER(email)=LOWER(?) LIMIT 1`, email).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) FindUserByPhone(phone string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE phone=? LIMIT 1`, phone).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) FindUserByID(id int) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(`SELECT id,name,email,COALESCE(username,''),password_hash,provider,COALESCE(google_sub,''),COALESCE(avatar_url,''),COALESCE(is_admin,0),COALESCE(is_verified,0),COALESCE(verification_type,''),COALESCE(profile_role,''),COALESCE(school,''),COALESCE(major,''),COALESCE(student_id,''),COALESCE(phone,''),COALESCE(phone_verified,0),COALESCE(employer_company,''),COALESCE(employer_tax_code,''),COALESCE(employer_representative,''),COALESCE(employer_website,''),COALESCE(landlord_name,''),COALESCE(landlord_address,''),COALESCE(landlord_phone,''),COALESCE(landlord_legal_info,''),COALESCE(account_status,'active'),COALESCE(locked_until,''),created_at FROM users WHERE id=? LIMIT 1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Username, &u.PasswordHash, &u.Provider, &u.GoogleSub, &u.AvatarURL, &u.IsAdmin, &u.IsVerified, &u.VerificationType, &u.ProfileRole, &u.School, &u.Major, &u.StudentID, &u.Phone, &u.PhoneVerified, &u.EmployerCompany, &u.EmployerTaxCode, &u.EmployerRepresentative, &u.EmployerWebsite, &u.LandlordName, &u.LandlordAddress, &u.LandlordPhone, &u.LandlordLegalInfo, &u.AccountStatus, &u.LockedUntil, &u.CreatedAt)
	return u, err
}

func (r *PostgresRepository) FindUsersByIDs(ids []int) (map[int]model.User, error) {
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

func (r *PostgresRepository) UpdateUserPasswordHash(id int, hash string) error {
	res, err := r.db.Exec(`UPDATE users SET password_hash=? WHERE id=?`, hash, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *PostgresRepository) UpdateUserProfile(id int, name, username, avatarURL, profileRole, school, major, studentID, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo string) error {
	res, err := r.db.Exec(`UPDATE users SET name=?, username=NULLIF(?,''), avatar_url=?, profile_role=?, school=?, major=?, student_id=?, phone_verified=CASE WHEN phone=? THEN phone_verified ELSE 0 END, phone=?, employer_company=?, employer_tax_code=?, employer_representative=?, employer_website=?, landlord_name=?, landlord_address=?, landlord_phone=?, landlord_legal_info=? WHERE id=?`, name, username, avatarURL, profileRole, school, major, studentID, phone, phone, employerCompany, employerTaxCode, employerRepresentative, employerWebsite, landlordName, landlordAddress, landlordPhone, landlordLegalInfo, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *PostgresRepository) SetUserVerification(id int, verified bool, verificationType string) error {
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
func (r *PostgresRepository) SetUserTrustVerified(id int, verified bool) error {
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
func (r *PostgresRepository) SetUserRoleVerification(id int, verificationType string) error {
	res, err := r.db.Exec(`UPDATE users SET verification_type=? WHERE id=?`, verificationType, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *PostgresRepository) SetUserAdmin(id int, isAdmin bool) error {
	res, err := r.db.Exec(`UPDATE users SET is_admin=? WHERE id=?`, isAdmin, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "tài khoản")
}
func (r *PostgresRepository) SearchUsers(query string, limit int) ([]model.User, error) {
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

func (r *PostgresRepository) SetUserAccountStatus(id int, status, lockedUntil string) error {
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
func (r *PostgresRepository) DeleteUser(id int) error {
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
func (r *PostgresRepository) CountAdmins() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin=1`).Scan(&n)
	return n, err
}

func (r *PostgresRepository) UpsertGoogleUser(name, email, googleSub string) (model.User, error) {
	email = strings.ToLower(email)
	_, err := r.db.Exec(`INSERT INTO users(name,email,provider,google_sub) VALUES(?,?,'google',?) ON CONFLICT(email) DO UPDATE SET name=excluded.name, google_sub=excluded.google_sub`, name, email, googleSub)
	if err != nil {
		return model.User{}, err
	}
	return r.FindUserByEmail(email)
}

func (r *PostgresRepository) CreateVerificationRequest(req model.VerificationRequest) error {
	_, err := r.db.Exec(`INSERT INTO verification_requests(user_id,type,info,status) VALUES(?,?,?,'pending')`, req.UserID, req.Type, req.Info)
	return err
}

func (r *PostgresRepository) GetLatestVerificationRequest(userID int) (model.VerificationRequest, error) {
	var v model.VerificationRequest
	err := r.db.QueryRow(`SELECT vr.id,vr.user_id,u.name,u.email,vr.type,vr.info,vr.status,vr.created_at FROM verification_requests vr JOIN users u ON u.id=vr.user_id WHERE vr.user_id=? ORDER BY vr.id DESC LIMIT 1`, userID).Scan(&v.ID, &v.UserID, &v.UserName, &v.UserEmail, &v.Type, &v.Info, &v.Status, &v.CreatedAt)
	return v, err
}

func (r *PostgresRepository) HasPendingVerificationRequest(userID int, requestType string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(1) FROM verification_requests WHERE user_id=? AND type=? AND status='pending'`, userID, requestType).Scan(&count)
	return count > 0, err
}

func (r *PostgresRepository) ListVerificationRequests(status string) ([]model.VerificationRequest, error) {
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

func (r *PostgresRepository) ResolveVerificationRequest(id int, status string) error {
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
