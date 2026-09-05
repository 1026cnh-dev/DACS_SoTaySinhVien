-- Supabase PostgreSQL schema for Sổ tay sinh viên
-- The application also runs this migration automatically on startup.

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
