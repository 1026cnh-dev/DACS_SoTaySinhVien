# Mạng xã hội cho sinh viên

Dự án này là một ứng dụng web dạng nền tảng chia sẻ bài viết, được xây dựng với kiến trúc backend sử dụng ngôn ngữ Go (Golang) kết hợp với cơ sở dữ liệu PostgreSQL (thông qua Supabase). 

Dự án bao gồm đầy đủ các tính năng cơ bản của một diễn đàn / blog cá nhân cho phép người dùng đăng nhập, tạo bài viết, lưu nháp, quản lý hồ sơ và tương tác với các nội dung.

## Công nghệ sử dụng

*   **Backend:** Go (Golang)
*   **Database:** PostgreSQL (quản lý bởi Supabase)
*   **Frontend:** HTML, CSS, JavaScript thuần (Vanilla JS)
*   **Containerization:** Docker
*   **Routing/API:** Go `net/http` tiêu chuẩn.

## Cấu trúc thư mục dự án

Dự án tuân theo cấu trúc phân chia module chuẩn của Go để dễ dàng bảo trì và mở rộng:

```text
├── app/                  
├── config/               # File cấu hình môi trường (.env)
│   ├── local.env         # Cấu hình chạy local
│   ├── production.env    # Cấu hình chạy production
│   └── environment       # Script load cấu hình
├── database/             # Các file script SQL (Schema)
│   ├── schema.sql        # Lược đồ database chung
│   └── supabase_schema.sql # Các functions/trigger đặc thù của Supabase
├── internal/             # Code logic chính của Backend (Golang)
│   ├── config/           # Khởi tạo config từ biến môi trường
│   ├── handler/          # HTTP Handlers (Controller) xử lý request từ client
│   ├── model/            # Cấu trúc dữ liệu (Structs) tương ứng với Database
│   ├── reference/        # Các file code mẫu/tham khảo
│   ├── repository/       # Tương tác với CSDL (Postgres, SQLite)
│   ├── service/          # Logic nghiệp vụ (Business logic) như Auth, Content
│   └── session/          # Xử lý phiên người dùng (Cookie/Session)
├── uploads/              # Thư mục chứa file, hình ảnh người dùng upload
├── web/                  # Giao diện người dùng (Frontend)
│   ├── admin/            # Giao diện quản trị hệ thống
│   ├── drafts/           # Giao diện quản lý bản nháp
│   ├── home/             # Giao diện trang chủ (Danh sách bài viết)
│   ├── icons/            # Asset SVG icons
│   ├── login/            # Giao diện Đăng nhập / Đăng ký
│   ├── post/             # Giao diện chi tiết bài viết
│   ├── profile/          # Giao diện hồ sơ cá nhân
│   ├── shared/           # Các component dùng chung (Header, Navigation...)
│   └── submit/           # Giao diện soạn thảo/đăng bài mới
├── Dockerfile            # Cấu hình build Docker image
├── go.mod / go.sum       # Quản lý dependencies của Go
├── main.go               # Điểm entry point khởi chạy server Go
├── render.yaml           # Cấu hình deploy lên nền tảng Render
└── run-local.sh / .bat   # Script khởi chạy dự án môi trường Dev
```

## Các chức năng chính (Core Features)

1.  **Quản lý người dùng (Authentication & Profile):**
    *   Đăng nhập, đăng ký tài khoản.
    *   Quản lý phiên đăng nhập (Session/Cookies).
    *   Xem và cập nhật thông tin hồ sơ cá nhân (`web/profile`).
2.  **Quản lý bài viết (Content Management):**
    *   Hiển thị danh sách tất cả bài viết trên trang chủ (`web/home`).
    *   Tạo bài viết mới (`web/submit`).
    *   Lưu bài viết dưới dạng bản nháp để chỉnh sửa sau (`web/drafts`).
    *   Xem chi tiết một bài viết (`web/post`).
3.  **Tương tác hệ thống:**
    *   (Dựa theo icons và template) Hỗ trợ tính năng bình luận (`message-circle.svg`), chia sẻ (`share-2.svg`), bookmark (`bookmark.svg`).
4.  **Quản trị viên (Admin Panel):**
    *   Giao diện dành riêng cho Admin (`web/admin`) để quản lý nội dung và người dùng toàn hệ thống.

## Hướng dẫn cài đặt và chạy dự án (Local)

1. **Yêu cầu hệ thống:**
   * Cài đặt sẵn [Go](https://go.dev/dl/).
   * Tài khoản [Supabase](https://supabase.com/) (để lấy URL và Key cho PostgreSQL).

2. **Cài đặt:**
   ```bash
   # Clone dự án (nếu có repo git) hoặc giải nén mã nguồn
   
   # Tải các thư viện Go (Dependencies)
   go mod tidy
   ```

3. **Cấu hình môi trường:**
   * Copy file `.env.example` thành `.env` (hoặc cấu hình trực tiếp vào `config/local.env`).
   * Điền các thông tin kết nối Database Supabase và Session Secret.

4. **Chạy ứng dụng:**
   * **Cách 1:** Chạy trực tiếp bằng Go
     ```bash
     go run main.go
     ```
   * **Cách 2:** Chạy bằng Script (Linux/Mac)
     ```bash
     ./run-local.sh
     ```
   * **Cách 3:** Chạy bằng Script (Windows)
     ```cmd
     run-local.bat
     ```

## Docker (Dành cho Production)

Dự án hỗ trợ chạy trong container. Để build và chạy bằng Docker:
```bash
docker build -t forum-app .
docker run -p 8080:8080 --env-file ./config/production.env forum-app
```
