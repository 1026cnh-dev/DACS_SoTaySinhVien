-- Dành riêng cho chạy lệnh trên nền tảng Supabase
CREATE TABLE categories (
    category_id SERIAL PRIMARY KEY,
    category_name VARCHAR(255) NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), 
    auth_provider VARCHAR(50) DEFAULT 'local', 
    role VARCHAR(50) DEFAULT 'student',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category_id INTEGER REFERENCES categories(category_id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO categories (category_name) VALUES 
('Học tập'), ('Học bổng'), ('Sự kiện'), ('Kỹ năng'), ('Việc làm'), ('Thông báo');
--====================================================================--
-- Tạo Bảng Chuyên Mục (Categories)
CREATE TABLE categories (
    category_id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_name VARCHAR(255) NOT NULL
);

-- Tạo Bảng Người Dùng (Users)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'student',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tạo Bảng Bài Viết (Posts)
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category_id INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE SET NULL
);

-- Thêm dữ liệu mẫu (Chuyên mục) để test
INSERT INTO categories (category_name) VALUES 
('Học tập'), ('Học bổng'), ('Sự kiện'), ('Kỹ năng'), ('Việc làm'), ('Thông báo');

-- Hàm Quản Trị Bài Viết (Posts)
INSERT INTO posts (title, content, category_id)
VALUES ('Thông báo học bổng học kỳ 1', 'Nội dung chi tiết của bài viết học bổng...', 2);

-- Hàm Lưu Tài Khoản Người Dùng (Users)
INSERT INTO users (full_name, email, password_hash, role)
VALUES ('Nguyễn Văn A', 'nguyenvana@example.com', 'chuoi_mat_khau_da_duoc_ma_hoa_bcrypt', 'student');
