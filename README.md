> Gói phát hành hiện tại: **V1.7_DACS_Dong_Bo_Giao_Dien_Dang_Bai_Da_Luu**

# Sổ tay sinh viên — DACS V1.24.36


## Cập nhật V1.7_DACS — Đồng bộ giao diện Đăng bài / Bài viết đã lưu

- Thiết kế lại **trang Đăng bài viết** theo cùng ngôn ngữ thị giác với Trang chủ: xanh thông Đà Lạt, vàng ấm, nền kem, thẻ trắng ngà, viền xanh nhẹ và bóng đổ mềm.
- Đồng bộ lại nhóm nút **Trở lại / Xem trước / Bài đã lưu**, khu chọn chuyên mục, các khối trường dữ liệu, editor, nút Lưu riêng/Đăng bài và popup xem trước.
- Thiết kế lại **trang Bài viết đã lưu** với phần tiêu đề xanh thông, điểm nhấn vàng, thẻ bài đã lưu/bản nháp cùng phong cách với card của website và trạng thái rỗng đồng bộ.
- Tối ưu responsive cho hai trang trên điện thoại; không thay đổi cấu trúc form, logic đăng bài, lưu bài, bản nháp, database, phân quyền hoặc các module khác.
- Chỉ cập nhật cache version của `submit.css` và `drafts.css`; JavaScript/backend giữ nguyên.


## Cập nhật V1.6_DACS — Autocomplete trường gọn

- Thay gợi ý `datalist` của ô **Trường / Đại học** bằng dropdown riêng để mỗi trường chỉ hiển thị một dòng tên trường, không còn các dòng trùng do mã/alias.
- Mã và alias như `DLU`, `ĐH Đà Lạt`, `VNU`, `ĐHQG...` chỉ dùng làm từ khóa tìm kiếm; khi chọn vẫn lưu tên chuẩn để bộ lọc **Cùng trường** không bị phân mảnh.
- Khi tên trường thành viên có hậu tố `- Đại học Quốc gia ...`, dropdown rút gọn phần hiển thị thành `- ĐHQG ...`; dữ liệu lưu trong hồ sơ vẫn giữ tên chuẩn đầy đủ.
- Xóa các dòng giao diện: hướng dẫn chọn trường, hướng dẫn ngành phổ biến và trạng thái `✓ Đã chuẩn hóa · ...`.
- Không thay đổi danh mục trường/ngành, database, quyền, bộ lọc hay các module khác.

## Cập nhật V1.5_DACS — Gợi ý và chuẩn hóa Trường / Ngành học

- Hồ sơ Sinh viên tải sẵn danh mục **201 trường/đại học/học viện/cơ sở đào tạo** trên toàn quốc và **105 ngành học phổ biến**; dữ liệu được đóng gói trong ứng dụng nên chạy local/Render không phụ thuộc API Internet.
- Ô **Trường / Đại học** hỗ trợ gợi ý theo tên chuẩn, tên thường dùng và mã/viết tắt (ví dụ `DLU`, `ĐH Đà Lạt`). Khi người dùng chọn hoặc lưu, backend quy về một tên chuẩn duy nhất, ví dụ **Trường Đại học Đà Lạt**.
- Backend từ chối lưu tên trường không nhận diện được đối với hồ sơ Sinh viên để tránh một trường bị tách thành nhiều giá trị khác nhau.
- Quy tắc **Cùng trường** không còn so sánh chuỗi thô; hệ thống chuẩn hóa tên/mã/alias trước khi so sánh và vẫn có lớp tương thích cho dữ liệu cũ chưa nằm trong danh mục.
- Ô **Ngành học** có gợi ý và chuẩn hóa các ngành đã biết; ngành mới/chuyên ngành đặc thù vẫn được phép nhập tự do để không khóa người dùng.
- Thêm API nội bộ `GET /api/education-options` để giao diện dùng chung danh mục trường/ngành. Không thay đổi schema database.
- Nguồn đối chiếu danh mục: Cổng thông tin tuyển sinh Bộ GD&ĐT và các danh sách cơ sở giáo dục của Cục Quản lý Chất lượng (cập nhật năm 2026).
- Kiểm thử riêng module chuẩn hóa: `DLU`, `ĐH Đà Lạt`, `Đại học Đà Lạt` đều quy về cùng một trường; kiểm tra không trùng tên/mã trong danh mục.


## Cập nhật V1.3_DACS — Xóa footer + gom tài liệu Markdown

- Xóa toàn bộ footer hiện có khỏi trang chủ, trang chi tiết bài viết và trang đăng nhập/đăng ký; các trang còn lại vốn không có footer riêng.
- Dọn toàn bộ CSS và responsive chỉ phục vụ các footer đã xóa, không để lại selector footer thừa.
- Gom toàn bộ tài liệu `.md` của gói phát hành vào duy nhất file `README.md`.
- Không thay đổi dữ liệu, chức năng bài viết, chuyên mục, form, tài khoản, phân quyền hoặc cơ sở dữ liệu.


**Đề tài:** Tìm hiểu ngôn ngữ Golang và tính năng Embed, xây dựng website “Sổ tay sinh viên”.

## Công nghệ
- Go 1.22 (`net/http`, `html/template`, Go Embed).
- HTML, CSS, JavaScript thuần.
- SQLite khi chạy local; PostgreSQL khi triển khai production.
- Render và Supabase có thể dùng cho bản online.

## Cấu trúc chính
```text
main.go                 Khởi động server, template và Go Embed
internal/               config, handler, model, repository, service, session
web/                    giao diện HTML/CSS/JS
database/               schema CSDL
config/                 cấu hình local/production
student_handbook.db     SQLite mẫu local
uploads/                tệp tải lên local
```

## Chạy local
Windows: `run-local.bat`  
macOS/Linux: `./run-local.sh`

Mở: **http://localhost:8080**

DB mẫu có **7 chuyên mục × 5 bài = 35 bài viết**.



## Cập nhật V1.24.27 — Category Module Studio
- Bấm trực tiếp **Tên chuyên mục**, **Hiển thị cho**, **Thẻ bài viết** và **Thẻ công ty** để chỉnh nhanh ngay trên bảng.
- Nút **Thiết kế form** chỉ quản lý các trường người dùng điền khi đăng bài, không trộn với cấu hình hiển thị chuyên mục.
- Form Module Studio cho phép đổi tên hiển thị từng field, bật/tắt, bắt buộc, kéo-thả thứ tự, bộ lọc, gợi ý dữ liệu và cho phép giá trị mới.
- Giữ nguyên mã kỹ thuật của field và cấu trúc CSDL để bảo toàn dữ liệu cũ.

## Cập nhật V1.24.24 — nhóm trường dạng tab
- Thu gọn các nút nhóm trường thành tab nhỏ.
- Mặc định mở nhóm **Nội dung chính**; chọn nhóm nào thì danh sách bên dưới chỉ hiện trường của nhóm đó.
- Giữ **Đang dùng** và **Tất cả** để kiểm tra nhanh toàn bộ cấu hình.
- Nút ↑/↓ sắp xếp trong nhóm đang hiển thị; kéo-thả và sửa tên trường vẫn giữ nguyên.
- Không thay đổi mã field, cấu trúc CSDL hay backend lưu cấu hình.

## Cập nhật V1.24.22 — sửa lỗi lưu dữ liệu
- Sửa nguyên nhân chính khiến **Lưu cấu hình chuyên mục báo thành công nhưng dữ liệu không đổi**: Form Builder gửi `FormData` dạng multipart nhưng backend trước đó chỉ gọi `ParseForm()`.
- Form Builder hiện gửi dữ liệu URL-encoded; backend vẫn hỗ trợ cả URL-encoded và multipart để tránh lỗi tương tự về sau.
- Backend không trả về thành công nếu thiếu payload Form Builder và đọc lại chuyên mục từ CSDL sau khi ghi để xác nhận.
- Các thao tác UPDATE/DELETE quan trọng kiểm tra `RowsAffected`, tránh thông báo “đã lưu/đã xóa” khi ID không tồn tại.
- Tên hiển thị trường đã chỉnh trong quản trị được áp dụng thật trên form Đăng bài / Sửa bài.
- Gia cố cập nhật bài viết, quảng cáo, hồ sơ, quyền tài khoản, xác thực, kiểm duyệt và sắp xếp.
- Bổ sung migration cho các cột chuyên mục mới khi dùng DB SQLite/PostgreSQL cũ.
- Cache `admin.js` và `submit.js` tăng lên `1.24.21`.


## Bố cục thẻ V1.24.18
- Bỏ cơ chế ép các hàng/card bằng nhau.
- Khu bài chính dùng CSS Grid masonry tính theo chiều cao thật của từng thẻ.
- Một thẻ có ảnh cao có thể nằm cạnh hai hoặc nhiều thẻ không ảnh thấp hơn.
- Ảnh của thẻ thông thường giữ tỷ lệ vuông 1:1.
- Nội dung thẻ lấy trực tiếp từ nội dung bài, chỉ hiển thị 2 dòng và tự `...`.
- Kiểu thẻ `horizontal` và kiểu `document` giữ layout riêng do quản trị cấu hình.
- Khi mở riêng một chuyên mục, cột Ghim/Tài trợ/Xu hướng được ẩn để tránh lẫn nội dung.

## Tài liệu
- `BAO_CAO_DO_AN.md`
- `HUONG_DAN_CAU_HINH_VA_SU_DUNG.md`
- `DATABASE_LOCAL.md`


## Cập nhật V1.24.22
- Sửa bộ lọc/phân nhóm trường trong Form Builder.
- Đưa thẻ xem trước form lên đầu cửa sổ cấu hình và cập nhật trực tiếp.
- Nới rộng không gian quản trị, tăng khoảng cách và vùng kéo-thả; giữ nguyên cấu trúc dữ liệu và chức năng.


## Cập nhật V1.24.23
- Xóa thẻ xem trước trực tiếp trong popup tạo/sửa chuyên mục.
- Giữ nút **Xem trước form**.
- Popup xem trước dùng lớp nổi cao nhất (`z-index: 1000000`) nên không bị popup tạo/sửa chuyên mục hoặc lớp nền che.
- Nhấn Esc đóng popup xem trước trước, sau đó mới đến popup bên dưới.


## V1.24.27
- Form Builder dùng thẻ module độc lập, 2 cột, không chồng lấn.
- Mặc định mở các module đang dùng; nhóm module ưu tiên theo từng chuyên mục/slug.
- Giữ nguyên mã kỹ thuật, cấu trúc CSDL và chức năng lưu cấu hình.

## Cập nhật V1.24.35
- Chuyên mục luôn hiển thị; nếu hồ sơ chưa đủ điều kiện thì khu vực bài viết hiện yêu cầu cập nhật hồ sơ thay vì tự chuyển trang.
- Form Builder trong tạo/sửa chuyên mục cho phép thêm và xóa trường tùy chỉnh, chọn kiểu dữ liệu, bắt buộc, sắp xếp và cấu hình lọc.
- Dữ liệu trường tùy chỉnh được lưu trong `meta_json`, dùng được khi đăng, lưu nháp, sửa và xem chi tiết bài viết.

---

# TÀI LIỆU GỘP: `HUONG_DAN_CAU_HINH_VA_SU_DUNG.md`

# HƯỚNG DẪN CẤU HÌNH VÀ SỬ DỤNG

## 1. Chạy local

Yêu cầu: Go 1.22+.

### Windows

```bat
run-local.bat
```

### macOS/Linux

```bash
./run-local.sh
```

Hoặc:

```bash
go run .
```

Mở trình duyệt tại:

```text
http://localhost:8080
```

## 2. Cấu hình local

File chính:

```text
config/local.env
```

Các biến cần chú ý:

```env
APP_ENV=local
PORT=8080
SQLITE_PATH=student_handbook.db
COOKIE_SECURE=false
SESSION_SECRET=chuoi-bi-mat-rieng
```

Nếu `DATABASE_URL` để trống, hệ thống sử dụng SQLite.

## 3. Database mẫu

File:

```text
student_handbook.db
```

Dữ liệu mẫu:

- 7 chuyên mục.
- 35 bài viết.
- 5 bài/chuyên mục.

Nếu xóa file DB và chạy lại, chương trình tạo schema rỗng. Chuyên mục và bài viết mẫu không tự sinh lại.

## 4. Chạy production

Production ưu tiên PostgreSQL qua:

```env
APP_ENV=production
DATABASE_URL=postgresql://...
SESSION_SECRET=...
COOKIE_SECURE=true
```

Trên Render nên cấu hình secret trực tiếp trong Environment Variables.

Nếu dùng Supabase Storage, khai báo thêm thông tin dự án/khóa dịch vụ theo cấu hình hiện có của source. Không commit service key lên GitHub.

## 5. Sử dụng website

### Người dùng

1. Đăng ký hoặc đăng nhập.
2. Chọn chuyên mục hoặc tìm kiếm bài.
3. Mở bài viết để xem chi tiết.
4. Có thể bình chọn, bình luận, lưu hoặc báo cáo khi đã đăng nhập.
5. Chọn **Đăng bài** để tạo nội dung mới.
6. Quản lý bài/nháp trong khu vực tài khoản.

### Quản trị

Trang admin dùng để:

- quản lý bài viết;
- chuyên mục và form động;
- người dùng;
- quảng cáo;
- báo cáo nội dung.

Trong form chuyên mục có thể kéo-thả hoặc dùng nút `↑` / `↓` để sắp xếp trường.

Quảng cáo có thể chỉ nhập ảnh; tiêu đề và nội dung là tùy chọn.

## 6. Lưu ý trước khi triển khai

- Đổi `SESSION_SECRET` thành chuỗi mạnh.
- Không công khai mật khẩu admin hoặc khóa dịch vụ.
- Kiểm tra `DATABASE_URL`.
- Kiểm tra Google Login nếu sử dụng.
- Kiểm tra upload file/ảnh production.
- Kiểm tra giao diện responsive.
- Sao lưu CSDL trước khi sửa schema hoặc xóa dữ liệu.

---

# TÀI LIỆU GỘP: `DATABASE_LOCAL.md`

# DATABASE LOCAL

Bản nội bộ sử dụng SQLite:

```text
student_handbook.db
```

Địa chỉ chạy:

```text
http://localhost:8080
```

## Dữ liệu mẫu

DB đóng gói sẵn có:

- 7 chuyên mục: Học tập, Học bổng, Sự kiện, Kỹ năng, Việc làm, Thông báo, Confession.
- 5 bài mẫu cho mỗi chuyên mục.
- Tổng cộng 35 bài viết.

## Reset dữ liệu

Chuyên mục/bài viết không được seed lại từ code.

Nếu xóa:

```text
student_handbook.db
```

rồi chạy lại ứng dụng, hệ thống chỉ tạo schema rỗng. Dữ liệu cũ không tự phục hồi.

Muốn quay lại dữ liệu mẫu, chép lại file `student_handbook.db` từ gói phát hành.

## Cấu hình

`config/local.env`:

```env
APP_ENV=local
PORT=8080
SQLITE_PATH=student_handbook.db
```

---

# TÀI LIỆU GỘP: `BAO_CAO_DO_AN.md`

# BÁO CÁO TÓM TẮT ĐỒ ÁN CƠ SỞ

## 1. Thông tin đề tài

**Tên đề tài:** Tìm hiểu ngôn ngữ Golang và tính năng Embed, xây dựng website "Sổ tay sinh viên".

**Phiên bản hiện tại:** DACS V1.24.26  
**Ngày cập nhật:** 01/09/2026

## 2. Mục tiêu

Đồ án nghiên cứu cách xây dựng website bằng Go, đặc biệt là tính năng `embed`, đồng thời phát triển website Sổ tay sinh viên để tập trung thông tin học tập và đời sống sinh viên.

Các mục tiêu chính:

1. Tìm hiểu Go, HTTP server và cách tổ chức project web.
2. Áp dụng Go Embed để đóng gói HTML, CSS, JavaScript và tài nguyên giao diện.
3. Xây dựng hệ thống bài viết theo chuyên mục, tìm kiếm và lọc nội dung.
4. Xây dựng tài khoản người dùng và các chức năng tương tác cộng đồng.
5. Xây dựng trang quản trị cơ bản.
6. Hỗ trợ SQLite cho local và PostgreSQL cho production.

## 3. Công nghệ sử dụng

| Thành phần | Công nghệ |
|---|---|
| Backend | Go 1.22, `net/http` |
| Template | `html/template` |
| Đóng gói giao diện | Go Embed (`//go:embed`) |
| Frontend | HTML, CSS, JavaScript |
| Local database | SQLite |
| Production database | PostgreSQL |
| Driver PostgreSQL | `github.com/lib/pq` |
| Driver SQLite | `modernc.org/sqlite` |
| Triển khai | Render |
| Dịch vụ dữ liệu tùy chọn | Supabase PostgreSQL/Storage |

## 4. Kiến trúc hệ thống

```text
Trình duyệt
    ↓ HTTP/HTTPS
Go HTTP Handler
    ↓
Service
    ↓
Repository
    ├── SQLite (local)
    └── PostgreSQL (production)
```

Frontend được đặt trong thư mục `web/` và được nhúng vào chương trình bằng Go Embed. Cách tổ chức này giúp file thực thi mang theo phần lớn tài nguyên giao diện, phù hợp với mục tiêu nghiên cứu của đề tài.

## 5. Cấu trúc mã nguồn

```text
main.go
internal/
├── config/
├── handler/
├── model/
├── repository/
├── service/
└── session/
web/
database/
config/
student_handbook.db
uploads/
```

Ý nghĩa:

- `main.go`: điểm khởi động, khai báo Embed và server.
- `handler`: tiếp nhận request và trả response.
- `service`: xử lý nghiệp vụ.
- `repository`: thao tác CSDL.
- `model`: định nghĩa dữ liệu.
- `session`: quản lý phiên đăng nhập.
- `web`: giao diện người dùng và quản trị.
- `database`: schema cho CSDL.

## 6. Chức năng đã xây dựng

### 6.1. Người dùng

- Đăng ký, đăng nhập, đăng xuất.
- Hồ sơ cá nhân và thông tin tài khoản.
- Đăng bài, lưu nháp, sửa và xóa bài thuộc quyền của mình.
- Xem bài viết chi tiết.
- Tìm kiếm và lọc bài theo chuyên mục.
- Bình chọn bài viết.
- Bình luận và trả lời bình luận.
- Lưu bài viết quan tâm.
- Báo cáo nội dung vi phạm.

### 6.2. Chuyên mục

Hệ thống hỗ trợ các chuyên mục mẫu:

- Học tập.
- Học bổng.
- Sự kiện.
- Kỹ năng.
- Việc làm.
- Thông báo.
- Confession.

Trang quản trị cho phép tạo/sửa chuyên mục và cấu hình form động. Các trường có thể kéo-thả hoặc dùng nút lên/xuống để thay đổi thứ tự.

### 6.3. Quảng cáo

- Quản lý quảng cáo từ trang admin.
- Tiêu đề và nội dung không bắt buộc.
- Có thể tạo quảng cáo chỉ bằng hình ảnh.
- Không cho lưu một quảng cáo hoàn toàn trống.
- Khi hiển thị, hình ảnh nằm sau tiêu đề/nội dung.

### 6.4. Quản trị

- Quản lý bài viết.
- Quản lý chuyên mục.
- Quản lý người dùng.
- Quản lý quảng cáo.
- Xử lý báo cáo và nội dung cần kiểm duyệt.

## 7. Cơ sở dữ liệu local

Bản chạy nội bộ sử dụng:

```text
student_handbook.db
```

DB mẫu hiện có:

- 7 chuyên mục.
- 5 bài mẫu cho mỗi chuyên mục.
- Tổng cộng 35 bài viết mẫu.

Chuyên mục và bài viết không còn được tự seed từ code. Nếu xóa `student_handbook.db`, chương trình tạo lại schema rỗng, vì vậy dữ liệu chuyên mục và bài viết cũ sẽ không tự xuất hiện lại.

## 8. Cấu hình môi trường

### Local

- URL: `http://localhost:8080`
- SQLite: `student_handbook.db`
- Cấu hình: `config/local.env`

### Production

- Dùng PostgreSQL qua `DATABASE_URL`.
- Có thể triển khai trên Render.
- Có thể sử dụng Supabase cho PostgreSQL/Storage.
- Secret như `SESSION_SECRET` và khóa dịch vụ phải đặt ở biến môi trường, không commit công khai.

## 9. Một số cải tiến kỹ thuật hiện tại

- Session không dùng map RAM theo kiểu cũ.
- Downvote không tự động xóa bài viết.
- Đã loại bỏ chức năng Tra cứu DLU và các phụ thuộc liên quan.
- Có kiểm tra độ dài và định dạng dữ liệu khi tạo/sửa bài.
- Giới hạn số bài lấy ở trang chủ để giảm tải truy vấn.
- Dùng gzip/cache cho tài nguyên phù hợp.
- Menu dấu ba chấm trang chi tiết hiển thị theo viewport để không bị cột bên cạnh che.

## 10. Tiến độ đến 01/09/2026

| Hạng mục | Trạng thái |
|---|---|
| Tìm hiểu Go và HTTP | Đã thực hiện |
| Go Embed | Đã tích hợp |
| Giao diện chính | Đã xây dựng |
| SQLite local | Đã tích hợp |
| PostgreSQL production | Đã tích hợp |
| Tài khoản | Đã xây dựng |
| Bài viết/chuyên mục | Đã xây dựng |
| Tìm kiếm/lọc | Đã xây dựng |
| Bình chọn/bình luận/lưu bài | Đã xây dựng |
| Trang quản trị | Đã xây dựng |
| Quảng cáo | Đã xây dựng và cải tiến |
| Form chuyên mục động | Đã xây dựng và cải tiến |
| Kiểm thử giao diện | Đang tiếp tục |
| Kiểm thử production | Cần tiếp tục |
| Báo cáo đồ án | Đang hoàn thiện |

So với kế hoạch đề cương, sản phẩm đã triển khai nhiều chức năng sớm hơn giai đoạn thiết kế giao diện dự kiến đầu tháng 09/2026. Tuy nhiên vẫn cần tiếp tục kiểm thử chức năng, responsive, PostgreSQL production, OAuth và lưu trữ file trước khi hoàn thiện đồ án.

## 11. Hạn chế và hướng phát triển

Các phần cần tiếp tục:

1. Kiểm thử đầy đủ trên desktop, tablet và mobile.
2. Kiểm tra hiển thị đủ bài trong từng chuyên mục và hoàn thiện phân trang.
3. Kiểm thử PostgreSQL/Render với dữ liệu thực tế.
4. Kiểm thử upload và lưu trữ file production.
5. Bổ sung kiểm thử tự động cho API quan trọng.
6. Tiếp tục tối ưu truy vấn và tải ảnh khi dữ liệu tăng.

## 12. Kết luận

Website Sổ tay sinh viên đã hình thành đầy đủ cấu trúc của một ứng dụng web bằng Go, có giao diện, tài khoản, bài viết, chuyên mục, tương tác cộng đồng, quản trị và CSDL. Tính năng Go Embed được áp dụng trực tiếp để đóng gói tài nguyên giao diện, phù hợp với nội dung nghiên cứu chính của đề tài.


## Cập nhật kiểm thử và sửa lỗi V1.24.21
Trong quá trình kiểm thử chức năng quản trị, nhóm phát hiện trường hợp thao tác lưu cấu hình chuyên mục trả về trạng thái thành công nhưng dữ liệu form chưa được cập nhật trong cơ sở dữ liệu. Nguyên nhân là dữ liệu phía trình duyệt được gửi bằng `multipart/form-data` trong khi handler sử dụng `ParseForm()`, khiến các trường động không xuất hiện trong `r.Form`.

Phiên bản V1.24.21 đã sửa bằng cách đồng bộ định dạng gửi/nhận dữ liệu, kiểm tra payload trước khi ghi, xác nhận số dòng bị tác động sau các lệnh cập nhật/xóa và đọc lại dữ liệu sau khi lưu. Đồng thời tên hiển thị của trường form do quản trị viên chỉnh sửa được áp dụng trực tiếp tại trang đăng/sửa bài viết. Các luồng cập nhật bài viết, quảng cáo, hồ sơ, tài khoản, xác thực và kiểm duyệt cũng được gia cố để không báo thành công giả khi dữ liệu không được thay đổi.

---

# TÀI LIỆU GỘP: `CHANGELOG_DACS_V1.1.md`

# DACS_V1.1

- Thiết kế lại thẻ thông báo giới hạn nội dung theo phong cách giao diện Sổ tay sinh viên: xanh thông Đà Lạt, điểm nhấn vàng, kích thước nhỏ gọn và responsive.
- Loại bỏ phần biểu tượng khóa bị phóng đại; đặt kích thước cứng 20×20 để không vỡ giao diện.
- Người chưa đăng nhập thấy tiêu đề “Đăng nhập để xem bài viết”; người đã đăng nhập nhưng thiếu điều kiện thấy “Cập nhật hồ sơ để xem bài viết”.
- Quản trị viên luôn bỏ qua giới hạn đối tượng của chuyên mục và bài viết.
- Quản trị viên có thể đăng vào chuyên mục `same_school` mà không cần hồ sơ sinh viên/trường học.
- Quản trị viên có thể sửa và xóa mọi bài viết qua giao diện bài viết, không chỉ bài do chính mình đăng.

## Footer chuẩn hóa
- Đưa footer trang chủ về cùng phong cách footer của các trang còn lại.
- Bỏ bo góc, shadow và khoảng trắng bao quanh kiểu card.
- Footer chạy toàn chiều ngang, chiều cao gọn hơn.
- Giữ logo + tên website bên trái và thông tin đồ án bên phải; mobile tự xếp dọc.

---

# TÀI LIỆU GỘP: `CHANGELOG_DACS_V1.2.md`

# DACS V1.2

## Thay đổi

- Hiển thị biểu tượng ổ khóa nhỏ cạnh tên chuyên mục khi tài khoản hiện tại chưa đủ điều kiện xem nội dung.
- Chuyên mục vẫn luôn hiển thị và vẫn có thể bấm mở; phần nội dung tiếp tục dùng card hướng dẫn đăng nhập/cập nhật hồ sơ.
- Tooltip của ổ khóa giải thích ngắn lý do quyền truy cập chưa đạt.
- Quản trị viên có toàn quyền nên không hiển thị ổ khóa ở bất kỳ chuyên mục nào.
- Giữ nguyên footer dạng chuẩn đã sửa ở bản trước.
- Không thay đổi dữ liệu chuyên mục, bài viết hoặc cấu hình form hiện có.

---

# TÀI LIỆU GỘP: `CHANGELOG_V1.24.31.md`

# DACS V1.24.31 — Auth Compact Clean

- Xóa mô tả: “Chỉ cần số điện thoại và mật khẩu. Hồ sơ có thể bổ sung sau.”
- Xóa ghi chú: “Thông tin cá nhân có thể bổ sung sau trong Hồ sơ.”
- Dọn CSS `.profile-note` không còn sử dụng.
- Giữ thanh chuyển Đăng nhập / Đăng ký ngay dưới tiêu đề.
- Không thay đổi endpoint, field form, Google Login hoặc logic xác thực.

---

# TÀI LIỆU GỘP: `CHANGELOG_V1.24.32.md`

# V1.24.32 — Auth Compact Flow

- Loại bỏ khoảng trắng lớn giữa nút đăng nhập/đăng ký và Google Sign-In.
- `auth-stage` không còn `flex: 1`; chiều cao vùng form tự khớp với nội dung đang hiển thị.
- Chuyển Login ↔ Register vẫn trượt/fade mượt, đồng thời chiều cao thẻ chuyển mượt theo nội dung.
- Khi trạng thái kiểm tra số điện thoại xuất hiện/ẩn, chiều cao vùng form tự cập nhật.
- Thẻ xác thực dùng chiều cao `auto` thay vì ép cao gần toàn màn hình.
- Giữ nguyên backend, action form, Google Login và logic kiểm tra tài khoản.

---

# TÀI LIỆU GỘP: `CHANGELOG_V1.24.33.md`

# DACS V1.24.33 — Adaptive Card Preview

- Bỏ chiều cao giả của tiêu đề bài viết: tiêu đề 1 dòng kéo nội dung sát lên ngay bên dưới.
- Thẻ không ảnh tận dụng phần không gian trống để hiển thị 3–4 dòng nội dung tùy chiều cao tiêu đề.
- Thẻ có ảnh hiển thị 2–3 dòng nội dung: tiêu đề càng ngắn thì phần mô tả càng được mở rộng.
- Khi nội dung dài hơn vùng hiển thị, tự cắt theo đúng kích thước thực tế và nối `… Xem thêm`.
- `Xem thêm` dẫn trực tiếp tới trang chi tiết bài viết.
- Giữ nguyên tỷ lệ thẻ 2:1, ảnh, vote, bình luận, lưu bài và bố cục 4 cột hiện tại.

---

# TÀI LIỆU GỘP: `CHANGELOG_V1.24.34.md`

# DACS V1.24.34 — Audience Content Gate

- Tất cả chuyên mục luôn xuất hiện trên thanh chuyên mục, không còn bị ẩn theo `audience_scope`.
- `Chuyên mục hiển thị cho ai?` được chuyển đúng nghĩa thành quyền xem **nội dung bên trong**, không phải quyền nhìn thấy tên chuyên mục.
- Khi mở chuyên mục bị giới hạn:
  - Chưa đăng nhập → chuyển tới Đăng nhập và giữ đường dẫn quay lại.
  - Đã đăng nhập nhưng hồ sơ chưa đạt điều kiện → chuyển tới Hồ sơ, hiển thị lý do cần cập nhật và giữ đường dẫn quay lại.
- Sau khi lưu hồ sơ từ luồng trên, hệ thống tự quay lại chuyên mục/bài viết người dùng đang muốn xem.
- Mở trực tiếp URL bài viết bị giới hạn cũng được kiểm tra quyền, không thể bỏ qua bằng liên kết trực tiếp.
- Quy tắc `Cùng trường` tiếp tục lọc nội dung theo trường của người xem và người đăng.
- Quyền đăng/sửa bài vẫn chỉ hiển thị các chuyên mục mà tài khoản có quyền thao tác, không nới lỏng luồng ghi dữ liệu.
- Cập nhật mô tả trong trang quản trị để làm rõ: chuyên mục luôn hiện, lựa chọn đối tượng chỉ giới hạn quyền xem nội dung.

---

# TÀI LIỆU GỘP: `CHANGELOG_V1.24.35.md`

# DACS V1.24.35 · Profile Gate + Custom Form Fields

## 1. Chặn nội dung chuyên mục ngay tại khu vực bài viết
- Tên/chuyên mục vẫn luôn xuất hiện trên thanh chuyên mục.
- Khi tài khoản chưa đáp ứng `Chuyên mục hiển thị cho ai?`, trang chuyên mục vẫn mở bình thường nhưng không truy vấn/hiển thị danh sách bài viết.
- Khu vực bài viết hiển thị thẻ thông báo `Cập nhật hồ sơ để xem bài viết` cùng lý do cụ thể.
- Chưa đăng nhập: nút hành động đưa tới Đăng nhập.
- Đã đăng nhập nhưng thiếu thông tin hồ sơ: nút hành động đưa tới Hồ sơ.
- Sau khi lưu hồ sơ thành công, tham số `next` đưa người dùng trở lại đúng chuyên mục đang xem.
- URL bài viết chi tiết vẫn được kiểm tra quyền ở phía server để không thể bỏ qua giới hạn bằng đường dẫn trực tiếp.

## 2. Cho phép tự tạo thêm trường trong Form Builder
- Áp dụng cho cả `Tạo chuyên mục` và `Cấu hình form` khi sửa chuyên mục.
- Thêm nút `+ Thêm trường` và nhóm `Trường tự tạo`.
- Các loại trường hỗ trợ: Văn bản ngắn, Nội dung dài, Số, Email, URL và Ngày.
- Trường tự tạo có thể đổi tên, bật/tắt, đặt bắt buộc, kéo sắp xếp, dùng làm bộ lọc/gợi ý và xóa.
- Cấu hình được lưu cùng `form_config` của chuyên mục và giữ nguyên khi mở lại để sửa.

## 3. Form đăng bài và dữ liệu bài viết
- Form đăng bài tự sinh các trường tùy chỉnh theo chuyên mục đã chọn.
- Giá trị được lưu trong `meta_json.custom_fields`, không yêu cầu thay đổi schema database.
- Hỗ trợ lưu nháp, sửa bài và hiển thị lại dữ liệu tùy chỉnh.
- Chi tiết bài viết hiển thị nhóm `Thông tin theo chuyên mục` nếu có dữ liệu tự tạo.
- Trường tự tạo có thể tham gia bộ lọc động khi được bật cấu hình lọc.

## 4. Kiểm tra
- `gofmt` đã chạy cho các file Go thay đổi.
- `node --check` đạt cho JavaScript admin, submit, home và post.
- Toàn bộ template HTML Go đã parse thành công bằng `html/template` với FuncMap tương ứng.
- `go test ./...` chưa thể hoàn tất trong môi trường đóng gói vì các module `github.com/lib/pq` và `modernc.org/sqlite` chưa có cache và mạng tới Go module proxy bị vô hiệu hóa.

## V1.4_DACS · Quảng cáo xen kẽ ngẫu nhiên

- Trang chủ: quảng cáo đang bật ở vị trí `feed-sponsored` được chọn và chèn ngẫu nhiên xen kẽ giữa các bài trong cột **Bài ghim** và **Xu hướng**.
- Trang chuyên mục: quảng cáo `feed-sponsored` được chèn ngẫu nhiên theo chiều dọc nhưng luôn khóa ở **cột đầu tiên** của lưới bài viết.
- Bỏ cách hiển thị quảng cáo `feed-sponsored` cố định ở đầu cột ghim.
- Tận dụng module Quảng cáo hiện có trong quản trị, không thay đổi dữ liệu quảng cáo cũ.
- Giới hạn mật độ để quảng cáo không lấn át nội dung: tối đa 2 quảng cáo/cột phụ ở trang chủ và tối đa 4 quảng cáo trong một chuyên mục.

## V1.8_DACS · Tối ưu độ thoáng thẻ Đăng nhập

- Tăng bề rộng thẻ đăng nhập từ 448px lên 520px để form không bị bó hẹp trên màn hình desktop/tablet.
- Tăng khoảng đệm và khoảng cách giữa logo, tiêu đề, tab, form, Google và nút quay lại theo từng nhóm nội dung.
- Tăng nhẹ chiều cao ô nhập, nút chính và vùng Google để thao tác dễ hơn nhưng vẫn giữ bố cục tập trung.
- Điều chỉnh chiều cao vùng chuyển Đăng nhập/Đăng ký để hiệu ứng chuyển tab không cắt nội dung.
- Giữ nguyên màu xanh thông - vàng ấm, toàn bộ handler đăng nhập/đăng ký, Google OAuth, session và database.
- Điều chỉnh breakpoint chiều cao để chế độ nén chỉ kích hoạt ở màn hình thực sự thấp.

## V1.9_DACS · Tối ưu toàn diện giao diện Mobile

- Chuẩn hóa vùng bấm của các nút quan trọng trên mobile lên khoảng 44px để thao tác bằng ngón tay dễ hơn.
- Bổ sung `viewport-fit=cover` và khoảng đệm safe-area cho thiết bị có tai thỏ / thanh Home.
- Tăng cỡ chữ ô nhập lên 16px ở mobile để hạn chế iPhone tự phóng to khi focus.
- Tối ưu Header mobile: nút Đăng bài/Tài khoản, ô tìm kiếm và thanh chuyên mục dưới màn hình có kích thước chạm lớn, khoảng cách đều và cuộn ngang mượt.
- Trang chủ/chuyên mục: chuyển feed về một cột rõ ràng, card thoáng hơn, ảnh theo tỷ lệ phù hợp mobile; cụm vote/bình luận/lưu vẫn nằm cùng hàng và dễ bấm.
- Trang chi tiết bài viết: tối ưu tiêu đề, nội dung, bình luận, vote, menu thao tác, modal và tài liệu cho màn hình nhỏ.
- Trang Hồ sơ: form một cột, input/nút dễ bấm, nhóm hành động tự xếp 2 cột hoặc 1 cột theo bề rộng màn hình; autocomplete trường học có vùng cuộn phù hợp.
- Trang Đăng bài: 3 nút đầu trang cùng hàng; editor cuộn ngang toolbar; nút Hủy/Lưu riêng/Đăng bài thành action bar gọn ở cuối form; modal chọn chuyên mục/xem trước tối ưu dạng bottom sheet.
- Trang Bài đã lưu: giữ ảnh thumbnail gọn bên trái, nút Bỏ lưu toàn chiều rộng; hai nút Chỉnh sửa/Xóa bản nháp được bố trí cùng hàng.
- Trang Đăng nhập/Đăng ký: input, checkbox, nút chính và Google có kích thước chạm hợp lý, thẻ tự co theo chiều rộng thiết bị.
- Trang Quản trị: menu module chuyển thành thanh cuộn ngang trên mobile, thống kê dạng 2 cột, bộ lọc và modal dễ thao tác, bảng vẫn cuộn ngang thay vì vỡ layout.
- Không thay đổi backend, database, route, logic đăng nhập, đăng bài, lưu bài, quảng cáo hoặc phân quyền.

## V1.10_DACS · Ghim thanh chuyên mục ở phía trên

- Thanh chuyên mục dùng chung được ghim ở **phía trên** khi cuộn trên desktop và mobile.
- Mobile bỏ cơ chế thanh chuyên mục cố định ở cạnh dưới; masthead có thể cuộn đi, thanh chuyên mục tiếp tục ở lại trên cùng.
- Trang **Đăng bài** và **Bài đã lưu** được bổ sung riêng thanh chuyên mục ghim, nhưng không đưa lại masthead lớn để giữ bố cục gọn.
- Trang đăng nhập/đăng ký giữ nguyên giao diện tập trung, không chèn thanh chuyên mục.
- Bỏ phần đệm đáy từng dùng để chừa chỗ cho thanh chuyên mục mobile ở cạnh dưới.
- Giữ nguyên logic bài viết, tài khoản, cơ sở dữ liệu, quảng cáo, chuẩn hóa trường/ngành và toàn bộ chức năng nghiệp vụ.



## V1.12_DACS · Ghim thanh chuyên mục ổn định trên mọi trang

- Sửa cơ chế ghim thanh chuyên mục theo một điểm neo chung cho Trang chủ, Chuyên mục, Chi tiết bài viết, Hồ sơ, Quản trị, Đăng bài và Bài viết đã lưu.
- Thanh chuyên mục cuộn theo tài liệu cho tới khi chạm mép trên, sau đó được cố định tại `top: 0`; kéo về đầu trang sẽ trở lại đúng vị trí ban đầu.
- Thay `position: sticky` phụ thuộc trình duyệt bằng cơ chế `fixed` + placeholder để tránh lỗi Safari/iOS/iPad khi trang có `overflow-x`.
- Xóa các rule mobile cũ còn ép thanh chuyên mục xuống đáy trong `home.css`, tránh xung đột giữa các phiên bản responsive.
- Giữ nguyên thanh cuộn ngang chuyên mục, trạng thái active, dark mode, vùng bấm mobile và các nút thao tác nhanh.
- Không thay đổi database, handler, bài viết, quảng cáo, tài khoản, phân quyền hoặc logic nghiệp vụ.

## V1.11_DACS · Sửa banner quảng cáo và luân phiên 5 giây

- Xóa vùng nền xám tách riêng phía trên banner quảng cáo lớn ở Trang chủ; ảnh quảng cáo nay phủ trọn khung, nhãn/nội dung chỉ nổi trực tiếp trên ảnh.
- Vị trí `home-right` nhận toàn bộ quảng cáo đang bật. Nếu có từ 2 quảng cáo trở lên, giao diện tự chuyển quảng cáo sau mỗi 5 giây với hiệu ứng mờ nhẹ, không tải lại trang.
- Giữ nguyên hệ thống quảng cáo xen kẽ trong feed.
- Làm rõ chế độ chỉnh sửa tại Quản trị → Quảng cáo: bấm `Sửa` sẽ nạp quảng cáo hiện tại vào form và đổi nút thành `Lưu thay đổi`; có thể sửa tiêu đề, nội dung, ảnh, liên kết, vị trí và trạng thái. Nếu không chọn ảnh mới, ảnh cũ được giữ nguyên.
- Không thay đổi cơ sở dữ liệu hoặc cấu trúc bảng quảng cáo.

## V1.13_DACS · Sửa, xóa quảng cáo và chỉnh ảnh trong xem trước

- Cột **Thao tác** trong danh sách quảng cáo được ghim ở cạnh phải để luôn thấy các nút **Xem trước / Sửa / Xóa**, kể cả khi bảng rộng hoặc phải cuộn ngang.
- Nút **Sửa** nạp lại toàn bộ dữ liệu quảng cáo vào form; nút **Xóa** xóa quảng cáo qua AJAX và cập nhật danh sách ngay, không cần tải lại trang.
- Khu vực bên phải đổi thành **Xem trước & chỉnh ảnh quảng cáo**.
- Ảnh trong bản xem trước có thể kéo trực tiếp để chọn vùng hiển thị; hỗ trợ thanh **Thu phóng**, **Căn giữa**, **Hoàn tác** và **Áp dụng ảnh đã chỉnh**.
- Khi áp dụng, trình duyệt tạo ảnh mới theo đúng tỷ lệ vị trí: banner Trang chủ `16:7`, feed `16:9`, quảng cáo cột bài viết `4:3`; ảnh mới được gửi cùng form khi lưu.
- Nếu đang chỉnh ảnh mà bấm Lưu luôn, hệ thống tự áp dụng phần căn/zoom hiện tại trước khi gửi dữ liệu.
- Khi thay ảnh của quảng cáo hoặc xóa quảng cáo, ảnh upload cũ được dọn khỏi Local Uploads/Supabase Storage khi có thể để tránh tệp mồ côi.
- Không thay đổi schema database, logic luân phiên banner 5 giây, quảng cáo xen kẽ feed, tài khoản hoặc bài viết.

## V2.0_DACS · Tối ưu tốc độ tải tối đa (Supabase / Render / Local)

### Mục tiêu
- Giảm thời gian chờ database từ nhiều request tuần tự sang batch + song song + cache ngắn hạn.
- Giữ nguyên giao diện, route, dữ liệu và chức năng hiện có của V1.13.
- Tối ưu cả lần tải đầu, lần tải lặp lại, ảnh/tài nguyên tĩnh và cold start PostgreSQL.

### Backend / database
- Xóa N+1 truy vấn tác giả ở Trang chủ, Chi tiết bài viết, Hồ sơ và Bài đã lưu; thêm `FindUsersByIDs` để lấy nhiều tài khoản trong 1 query.
- Trang chủ chạy song song các truy vấn độc lập: chuyên mục/tài khoản phiên, bài ghim, bài hôm nay, xu hướng, danh sách chặn và quảng cáo.
- Gộp thống kê bài của các cột thành 1 batch và trạng thái đã lưu thành 1 batch; hai batch này chạy song song.
- Trang chủ bình thường không còn tải 60 bài mới nhất chỉ để hiển thị 8 bài; chỉ truy vấn fallback nhỏ khi không có bài hôm nay.
- Chi tiết bài viết chạy song song comments, stats, vote, quảng cáo, gợi ý và trạng thái chặn; thống kê bài gợi ý lấy 1 batch thay vì từng bài.
- Hồ sơ bỏ việc tải toàn bộ danh sách yêu cầu xác thực chỉ để kiểm tra một số điện thoại; dùng query `HasPendingVerificationRequest` có index và chạy song song với bài hồ sơ/yêu cầu mới nhất.
- Trang Quản trị chạy song song các tập dữ liệu độc lập thay vì chờ tuần tự.
- Cache RAM có invalidation:
  - chuyên mục 60 giây;
  - quảng cáo đang bật 30 giây;
  - dữ liệu bộ lọc 45 giây;
  - bài ghim 10 giây;
  - bài hôm nay / xu hướng 6 giây;
  - danh sách tài khoản bị chặn 20 giây;
  - hồ sơ người dùng 20 giây.
- Cache được xóa ngay sau thao tác ghi liên quan trong cùng instance để tránh giữ dữ liệu cũ không cần thiết.
- PostgreSQL pool tăng lên 16 connection mở / 8 idle, có giới hạn lifetime/idle phù hợp đường đọc song song.
- `ListTodayPosts` PostgreSQL đổi từ `DATE(created_at...)` sang khoảng thời gian sargable để index `created_at` được sử dụng.
- Bổ sung index `posts(created_at DESC,id DESC)`, `saved_posts(user_id,post_id)` và `verification_requests(user_id,type,status,id DESC)`.
- Seed PostgreSQL khi khởi động được gộp từ khoảng 10 round-trip thành 1 round-trip.
- Production tắt access log chi tiết mặc định (`DACS_ACCESS_LOG=0`) để giảm I/O không cần thiết; vẫn có thể bật lại khi debug.

### Frontend / static assets
- Logo cũ `logo.svg` 133.233 byte (thực chất chứa ảnh nhúng) được thay bằng `logo.webp` 33.122 byte: giảm 100.111 byte (~97,8 KiB, 75,1%).
- Ảnh bài/feed/gợi ý/quảng cáo ngoài vùng nhìn dùng `loading="lazy"`, `decoding="async"`; ảnh chính được ưu tiên tải.
- Card dài ngoài viewport dùng `content-visibility:auto` khi trình duyệt hỗ trợ, giảm layout/paint ban đầu.
- CSS/JS/icon được version hóa `v=2.0.0` và cache 1 năm `immutable`; uploads dùng cache 1 năm vì URL ảnh upload thay đổi khi thay file.
- Gzip writer được tái sử dụng bằng `sync.Pool`; bỏ gzip cho WebP/JPEG/PNG/PDF/ZIP/WOFF2 vốn đã nén sẵn.

### Benchmark
Benchmark dưới đây đo **phần chờ database** theo chính cấu trúc request V1.13 và V2.0, với RTT Supabase giả lập 80 ms/query (median nhiều lần chạy trong môi trường đóng gói). Đây không phải số đo mạng production vì môi trường đóng gói không thể kết nối Supabase bên ngoài.

| Kịch bản Trang chủ đăng nhập | V1.13 | V2.0 | Giảm |
|---|---:|---:|---:|
| Cache lạnh | 2.885,3 ms | 322,8 ms | **2.562,5 ms (88,8%)** |
| Cache đã nóng | 2.885,3 ms | 80,7 ms | **2.804,6 ms (97,2%)** |

- Mô hình V1.13 điển hình: khoảng 36 round-trip DB tuần tự khi feed có khoảng 20 tác giả khác nhau.
- V2.0 cache lạnh còn khoảng 10 query thực nhưng được gom thành 4 pha I/O song song.
- V2.0 cache nóng: các dữ liệu feed/chuyên mục/quảng cáo/tác giả/chặn lấy từ RAM; phần DB còn lại chủ yếu là stats + trạng thái lưu và chạy song song.
- Thời gian tải browser thực tế còn phụ thuộc Render cold start, RTT thực tế tới Supabase, băng thông và kích thước ảnh bài/quảng cáo. Vì vậy không coi 80,7 ms là tổng thời gian tải trang production.

### Kiểm tra
- `gofmt` toàn bộ file Go thay đổi: đạt.
- Compile/test toàn project bằng stub driver offline cho `pq` và `modernc/sqlite`: `go test ./...` đạt tất cả package; test `internal/reference` đạt.
- 8 Go HTML template parse thành công.
- 8 file JavaScript qua `node --check`.
- 8 file CSS cân bằng cấu trúc.
- Chỉ còn 1 file Markdown: `README.md`.
- `go test ./...` với dependency thật không thể tải module trong môi trường đóng gói do DNS/mạng tới `proxy.golang.org` bị chặn; đây là giới hạn môi trường kiểm tra, không phải lỗi biên dịch của source V2.0.

## V2.1_DACS · Trang chủ hiển thị toàn bộ bài viết

- Trang chủ không còn giới hạn cột bài chính trong các bài đăng của hôm nay.
- Toàn bộ bài viết thường có quyền hiển thị được lấy theo thứ tự `created_at DESC, id DESC`, vì vậy bài vừa đăng tự nằm trên cùng và các bài cũ vẫn tiếp tục xuất hiện phía dưới.
- Danh sách toàn thời gian được cache RAM 6 giây và tự vô hiệu hóa khi thêm/sửa/xóa/ghim bài, giữ lại tối ưu tốc độ của V2.0.
- Giữ nguyên quy tắc trước đó: tài liệu dạng `document` không trộn vào feed Trang chủ.
- Xóa nhãn/ranking hiển thị `XU HƯỚNG` trên các thẻ ở cột nổi bật; cột vẫn sử dụng thuật toán tương tác hiện có nhưng giao diện không còn gắn nhãn xu hướng.
- Bài ghim được bổ sung cùng khối tác giả như bài thường: ảnh đại diện, tên, dấu xác minh/vai trò và ngày đăng.
- Không thay đổi database schema, quảng cáo, phân quyền, bộ lọc, đăng nhập hay các chức năng khác.

