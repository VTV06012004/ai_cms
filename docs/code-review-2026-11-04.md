# Review ai_cms — 04/11/2026

**Người review:** senior engineer BKTech (mentor) · **Gửi:** intern phụ trách ai_cms · **PM:** nguyen
**Phạm vi:** `origin/main` @ `08e9464` (Merge PR #15, 04/11/2026). Đây là code mới nhất. `origin/master` gần như đứng yên từ 26/10, xem mục 3.6.
**Đối chiếu với:** review cũ `docs/code-review-2026-10-26.html` (commit `08cd604`) và chuẩn BKTech (api.bkgiaoduc, api.register-account, web.thoikhoabieu, `claude-workflow.md`).

---

## 1. Tóm tắt

Em tiến bộ rõ trong 9 ngày. **8/8 lỗi nghiêm trọng của review cũ đều đã có bản sửa và có test**, phần lớn sửa đúng gốc chứ không vá bề mặt. Vài chỗ làm tốt:

- Generated column `active_slug` cho unique slug, kèm integration test chạy trên MySQL thật trong CI.
- Upload dùng `MaxBytesReader` + `image.DecodeConfig` + `nosniff` + CSP sandbox.
- Bảng `errorRules` map lỗi sang HTTP status.
- Package `config` có test.
- Graceful shutdown.

`gofmt`, `go build`, `go vet` và `go test -race` đều sạch.

Nhưng **chưa sẵn sàng production**, vì 3 nhóm vấn đề:

1. **Lỗi nghiệp vụ mới** sinh ra từ tính năng hẹn giờ và phân quyền. Mình đã chạy test xác nhận cả ba:
   - Author không sửa được bài đã đăng của chính mình.
   - Sửa bài hẹn giờ thì bị lỗi.
   - Bài đăng trực tiếp không có `published_at`.
2. **Vận hành DB:** schema do AutoMigrate cộng `ALTER TABLE` tự chạy mỗi lần khởi động, không có migration theo version.
3. **Quy trình git:** `master` bị force-push làm mất merge của PR #2/#3, 8 PR mở bị bỏ quên, PR #13/#15 tự merge không ai review, và commit tên "update".

### Mức sẵn sàng

| Hạng mục | Đánh giá |
|---|---|
| Đọc hiểu và tổ chức code | 🟢 Tốt |
| Đúng nghiệp vụ | 🟡 Còn 3 lỗi mức Cao (mục 3.1) |
| Bảo mật | 🟡 Đã vá phần lớn; còn rotate secret cũ, rate limit, Host header |
| DB / migration / transaction | 🔴 Chưa có migration versioned, chưa dùng `context` |
| Test | 🟡 Usecase 57%, domain 89%, nhưng persistence/handler còn mỏng và FE chưa có test |
| Git / quy trình BKTech | 🔴 Cần học lại ngay |
| **Kết luận** | **Có thể join dự án chính thức sau lộ trình 2 tuần (mục 4)**, bắt đầu bằng ticket nhỏ, có người review từng PR |

---

## 2. Đối chiếu review cũ (26/10)

| # | Lỗi cũ | Trạng thái | Ghi chú |
|---|---|---|---|
| 1 | Người lạ xem được bài nháp | ✅ Đã sửa | `post_usecase.go:30-48`. Khách chỉ thấy bài published, author thấy bài của mình cộng bài published. |
| 2 | Không ai lên admin được | ✅ Đã sửa | `user_usecase.go:73-119` `EnsureInitialAdmin` lấy từ env, có check lại trong transaction. Đã bỏ logic "id = 1". |
| 3 | Trùng slug | ✅ Đã sửa tốt | `post_migrations.go:9-38` generated column + unique index. `post_repo_mysql_test.go` chạy trên MySQL thật trong CI. |
| 4 | Xem trộm revision | ✅ Đã sửa | `revision_usecase.go:19-28`. Compare/Restore cũng check quyền. |
| 5 | Upload không an toàn | ✅ Đã sửa tốt | `media_handler.go:24-87`, `media_usecase.go:41-59`. Còn 1 vấn đề mới (mục 3.2-M3). |
| 6 | JWT secret mặc định / 24h | ✅ Đã sửa | Bắt buộc `JWT_SECRET`, token 15 phút. Phát sinh vấn đề UX (3.2-M6). |
| 7 | `.env` trong git | ⚠️ Sửa một nửa | Đã gỡ `.env` khỏi git, có `.gitignore` và `.env.example`. **Lịch sử git vẫn còn secret cũ.** Em liên hệ mentor để biết chi tiết và xác nhận đã đổi mật khẩu DB. |
| 8 | Sửa bài thì bị gỡ | ✅ Đã sửa | `post.go:177-179`. Nhưng **cùng lớp lỗi đã quay lại** với `scheduled_at` (3.1-H2). |

**Mục "Cần cải thiện thêm":**

| Mục | Trạng thái |
|---|---|
| Thông báo lỗi / status code | 🟡 Có `writeError`, nhưng còn 19 chỗ tự viết `c.JSON(...)` lỗi. Nhất quán chưa đủ (3.2-M5). |
| Đừng giấu lỗi | ❌ `post_handler.go:48-51` vẫn biến **mọi** lỗi (kể cả DB sập) thành 404. |
| Phân trang | ✅ posts/users/audit/media. Categories/tags/menus chưa có, chấp nhận được vì dữ liệu nhỏ. |
| Test usecase + phân quyền | ✅ Có, usecase 57.1%. |
| Một quy tắc một chỗ | ✅ `EnsureCanSetPostStatus`, đã bỏ `approved`/`CanApprove` ở domain. Enum DB vẫn còn `approved` (3.1-H4). |
| Gom config | ✅ `internal/config`, test 94%. |
| Tắt app an toàn | ✅ `main.go:114-147`. |
| FE dùng API client chung | 🟡 Có interceptor ở `lib/api.ts`, nhưng 16 chỗ vẫn tự gắn header token (3.2-M7). |

**Tính năng mới** (danh mục/tag/menu, hẹn giờ, thư viện ảnh, so sánh phiên bản, tài liệu API) đều đã có. Riêng hẹn giờ có lỗi, xem dưới.

---

## 3. Phát hiện theo mức độ

> Đường dẫn file tính từ gốc repo, trên `origin/main@08e9464`. Các lỗi đánh dấu **[đã xác nhận]** là mình đã chạy test tạm để tái hiện (test đó không commit vào repo).

### 3.1 🔴 Cao: sửa trước khi làm gì khác

**H1. Author không sửa được bài đã đăng của chính mình [đã xác nhận]**

- **Ở đâu:**
  - `web/app/posts/edit/[id]/page.tsx:115` nạp `status = post.status` (ví dụ `published`).
  - Dòng `196` luôn gửi lại `status`.
  - Dòng `423-434`: dropdown của author không có option `published`, nhưng state vẫn giữ giá trị đó.
  - Backend `internal/usecase/post_usecase.go:114-119` thấy status khác rỗng nên gọi `EnsureCanSetPostStatus(published)`, trả **403** "vai trò hiện tại không có quyền xuất bản bài viết".
- **Hậu quả:** Author mở bài published của mình, sửa một chữ, bấm Lưu thì lỗi. Bài scheduled cũng vậy.
- **Cách sửa:**
  1. Quyết định nghiệp vụ trước, hỏi PM. Author sửa bài đã đăng thì (a) giữ nguyên published, hay (b) bài phải quay về `review`?
  2. Backend chỉ kiểm quyền khi status **thay đổi** (`requestedStatus != post.Status`). Đây là quy tắc domain, nên đặt trong `Post.Update` hoặc `ChangeStatus`, không đặt ở usecase.
  3. FE chỉ gửi `status` khi người dùng thực sự đổi.
  4. Viết test usecase: "author sửa bài published của mình, không đổi status → 200".

**H2. Sửa bài hẹn giờ mà không gửi `scheduled_at` sẽ bị 400 [đã xác nhận]**

- **Ở đâu:**
  - `internal/domain/post.go:224-231`: `ApplySchedule(nil)` trên bài `scheduled` trả `ErrInvalidScheduleTime`.
  - `post_usecase.go:125` luôn gọi hàm này với `input.ScheduledAt`.
- **Hậu quả:** Client nào (app, script, Postman) PUT bài hẹn giờ mà không gửi lại giờ thì bị từ chối. Ngoài ra, nếu giờ hẹn cũ đã qua nhưng scheduler chưa kịp chạy, FE gửi lại giờ cũ cũng bị 400.
- **Đây đúng là bài học "không gửi = không đổi" của lỗi #8.** Em đã sửa cho `status` nhưng chưa áp cho trường mới.
- **Cách sửa:** `ApplySchedule` nhận `nil` thì giữ `p.scheduledAt` hiện có, chỉ validate khi có giá trị mới. Thêm test cho cả hai nhánh.

**H3. Bài đăng trực tiếp không có `published_at` [đã xác nhận]**

- **Ở đâu:** `published_at` chỉ được set trong `Post.Moderate` (`post.go:200`) và `PublishScheduled` (`post.go:244`). Khi admin/editor tạo hoặc sửa bài với `status=published` (`NewPost` `post.go:64-86`, `Update` `post.go:165-189`) thì `published_at = NULL`.
- **Hậu quả:** Sắp xếp theo ngày đăng, hiển thị "đăng lúc", RSS, sitemap về sau đều sai.
- **Cách sửa:** Đưa việc chuyển trạng thái vào **một** method domain, ví dụ `transitionTo(status, now)`: lần đầu chuyển sang `published` thì set `publishedAt`. Thêm test.

**H4. Migration không có version, `ALTER TABLE` chạy lại mỗi lần khởi động**

- **Ở đâu:**
  - `internal/infrastructure/persistence/post_migrations.go:40-46`: `ALTER TABLE posts MODIFY COLUMN status ENUM(...)` chạy **vô điều kiện** ở mọi lần start (`cmd/server/main.go:196-213`). Bảng lớn thì mỗi lần deploy là một lần lock hoặc rebuild bảng.
  - Schema là AutoMigrate trộn raw SQL.
  - Enum vẫn có `'approved'` (`post_model.go:17`, `post_migrations.go:42`) dù domain đã bỏ.
- **Chuẩn BKTech:** api.bkgiaoduc **không dùng AutoMigrate**. Migration là file `.sql` đặt theo ticket trong `migrations/` và **không sửa file đã chạy**. Khi lên prod, PR phải ghi rõ "migration TRƯỚC binary".
- **Cách sửa:**
  1. Tạo `migrations/0001_init.sql` từ schema hiện tại, các thay đổi sau thêm file mới.
  2. Chạy migration bằng lệnh riêng (`make migrate` hoặc tool, xem 3.2-M1), không chạy trong `main`.
  3. Bỏ `AutoMigrate` khỏi đường khởi động.
  4. Dọn `'approved'` bằng một migration riêng (kiểm tra dữ liệu trước).

**H5. Git hygiene và secret trong repo public**

- **Secret:** Lịch sử git vẫn còn secret cũ, liên hệ mentor để biết chi tiết. Phải coi secret đó là đã lộ và đổi mật khẩu tương ứng. Viết lại lịch sử chỉ là phụ; đổi mật khẩu mới là bắt buộc.
- **Nhánh lệch nhau:** Default branch là `master`, nhưng code thật nằm ở `main`.
  - PR #2 và #3 đã "MERGED" vào `master`, nhưng merge commit (`9e41c98…`) **không còn trong `master` lẫn `main`**, tức `master` đã bị force-push.
  - 8 PR #4–#11 vẫn OPEN trỏ vào `master`, trong khi code tương ứng đã vào `main` qua nhánh `improvements` bằng merge local ("Merge issue_1"…).
  - PR #13 và #15 tự merge, **0 review**.
- **Cách sửa:**
  - Chọn **một** nhánh chính và đặt làm default.
  - Đóng các PR đã lỗi thời, ghi chú "đã vào main qua #13".
  - Bật branch protection: không force-push, bắt buộc CI xanh và ít nhất 1 review.
  - Từ nay mỗi ticket một nhánh, một PR, không merge local rồi push.

### 3.2 🟡 Trung bình: làm trong lộ trình 2 tuần

**M1. Không có `context.Context` ở bất kỳ tầng nào.** `grep "context\."` trong `internal/` không trả kết quả.

- Request bị huỷ, timeout hay shutdown đều không dừng được query.
- Scheduler (`main.go:151-173`) có ctx nhưng không truyền xuống `PublishDuePosts` / DB.
- Chuẩn BKTech: repo dùng `dbWithContext(ctx)`, và "goroutine phải có lifecycle do context quản lý".
- **Sửa:** thêm `ctx context.Context` làm tham số đầu cho usecase và repo, rồi `db.WithContext(ctx)`. Handler truyền `c.Request.Context()`.

**M2. Scheduler có thể đăng trùng.** Comment ở `post_usecase.go:217` ghi "đọc lại trong transaction để tránh đăng trùng", nhưng `GetByID` không khoá dòng (không `FOR UPDATE`). Chạy 2 instance (hoặc rolling deploy) thì cả hai cùng thấy `scheduled`, cùng publish, cùng ghi audit.

- **Sửa:** dùng update có điều kiện `UPDATE posts SET status='published', published_at=scheduled_at, scheduled_at=NULL WHERE id=? AND status='scheduled' AND scheduled_at<=?`, rồi kiểm `RowsAffected == 1` mới ghi audit. Hoặc `SELECT … FOR UPDATE SKIP LOCKED`.
- Đồng thời sửa comment cho đúng sự thật.

**M3. URL ảnh lấy từ header `Host` và luôn là `http`.** `media_handler.go:68-75` và `96-102`.

- Đứng sau reverse proxy HTTPS thì URL thành `http://…`, bị chặn mixed-content.
- Client gửi `Host` giả thì URL giả bị lưu vĩnh viễn vào nội dung bài.
- **Sửa:** thêm `PUBLIC_BASE_URL` vào `config`, hoặc chỉ lưu đường dẫn tương đối `/uploads/x.png` (FE đã có `mediaUrl()` để ghép).

**M4. Tự viết PBKDF2 và không có rate limit.**

- `internal/infrastructure/security/credentials.go:14-62` tự cài PBKDF2. Cài đúng, nhưng quy tắc là **không tự viết crypto**. `golang.org/x/crypto` đã có trong `go.sum`: dùng `bcrypt` hoặc `argon2id`, hoặc ít nhất `x/crypto/pbkdf2`.
- 600.000 vòng mỗi lần `/login` và `/register` mà không có rate limit, nên vài chục request song song là đủ ăn hết CPU.
- **Sửa:** thêm rate limit theo IP và email cho `/api/auth/*`. Trong `Register`, validate tên và email **trước** khi hash (`user_usecase.go:51-58` đang hash trước).

**M5. Xử lý lỗi chưa nhất quán.**

- `post_handler.go:48-51`: mọi lỗi thành 404. Mục "đừng giấu lỗi" của review cũ vẫn còn ở đây.
- `revision_handler.go:49-56` tự `switch`, `media_handler.go:57-66` tự map. Tổng cộng 19 chỗ `c.JSON(http.Status…)` lỗi viết tay, so với 27 chỗ dùng `writeError`.
- Đã có `parseIDParam` và `currentUserID` (`params.go`) nhưng `post_handler.go` và `revision_handler.go` không dùng, nên chấp nhận cả `id=0`.
- **Sửa:** mọi lỗi đi qua `writeError`. Đưa `ErrInvalidImage` và `ErrImageTooLarge` vào `errorRules`. Mọi ID dùng `parseIDParam`.
- **Envelope chuẩn BKTech** (api.bkgiaoduc): thành công `{"data": …}`, lỗi `{"errors":[{"code","message","detail"}]}`. Chưa cần đổi ngay, nhưng em phải biết khi vào dự án.

**M6. JWT 15 phút không có refresh token.**

- Token hết hạn thì FE tự đăng xuất (`web/lib/api.ts:107-112`).
- Trang sửa bài dài 700 dòng không lưu nháp, nên người viết lâu hơn 15 phút **mất bài** khi bấm Lưu.
- **Sửa** (chọn một):
  - Refresh token (cookie httpOnly) + access token ngắn.
  - Hoặc access token dài hơn nhưng mỗi request kiểm role từ DB, kèm autosave nháp ở FE.
- Nên bàn với PM trước khi làm.

**M7. Frontend.**

- 16 chỗ vẫn tự gắn `headers: { Authorization: … }` dù interceptor đã làm. Ví dụ `posts/edit/[id]/page.tsx` có 7 chỗ, `admin/users/page.tsx` 3, `admin/logs/page.tsx` 2.
- `localStorage` đọc rải rác ở từng trang.
- Interceptor gắn vào **axios global** (`lib/api.ts:116-125`) thay vì `axios.create()`.
- Dùng `alert()` và `err: any`.
- **Không có test và không có test runner.** Comment ở `lib/api.ts:65` nói "exported để test" nhưng chưa có test nào.
- `package.json` không có script `typecheck` hay `test`.
- Điểm tốt: render nội dung bài qua JSX, không dùng `dangerouslySetInnerHTML`, nên không có XSS từ nội dung.
- **Sửa:**
  - Một `apiClient = axios.create()` + một hook `useSession`.
  - Bỏ mọi header thủ công.
  - Thêm Vitest + Testing Library: test `installSessionHandling` và form sửa bài (case H1).
- Tham khảo `web.thoikhoabieu`: `src/services/Api.service.js` + `utils/axiosClient.js`, TanStack Query cho server state, Vitest. Lưu ý repo đó dùng **react-router-dom 7**, không phải TanStack Router.

**M8. Repository ghi đè toàn bộ cột.**

- `post_repo.go:86` `tx.Save(&model)` ghi **mọi** cột từ domain. `Excerpt`, `SEOTitle`, `SEODescription` không có đường nào để set (domain không có setter), nên luôn bị ghi thành rỗng. Đó là field chết: hoặc làm cho xong, hoặc bỏ.
- `replacePostTags` (`post_repo.go:115-123`) **lặng lẽ bỏ qua** `tag_id` không tồn tại. Nên trả `ErrTagNotFound`.

**M9. Log và vận hành.**

- Dùng `log.Printf` của stdlib, không có request-id, không có field có cấu trúc.
- `gin.Default()` (`main.go:92`) chạy debug mode khi chưa set `GIN_MODE=release`.
- `http.Server` chỉ có `ReadHeaderTimeout` (`main.go:108-112`), thiếu `ReadTimeout`, `WriteTimeout`, `IdleTimeout`.
- Chuẩn BKTech: logrus `WithFields` / `WithError`, **không log password/token/OTP**.

**M10. Test và CI.**

- Coverage: persistence 1.8% (integration test bị skip khi không có `MYSQL_TEST_DSN`; trên CI thì có chạy), storage 0%, handler 24.8%.
- CI và Dockerfile dùng **Go 1.21**, đã hết hỗ trợ.
- CI chưa chạy `gofmt` và lint (Dockerfile có gofmt, CI thì không).
- **Sửa:** test handler bằng `httptest` cho từng route, nhất là các mã 401/403/404/409. Thêm `golangci-lint` hoặc tối thiểu `go vet` + `gofmt -l` vào CI. Nâng Go lên 1.23+.

### 3.3 🟢 Thấp: dọn khi tiện

- `routes.go:27`: health trả "MySQL 9.7" trong khi compose dùng `mysql:8.4`.
- `docker-compose.yml:22`: `DB_DSN` fallback `cms_password`, trong khi service `db` (dòng 7) bắt buộc phải set `MYSQL_PASSWORD`. Bỏ fallback để hai bên khớp nhau.
- `media_handler.go:129`: `var _ = domain.ErrMediaNotFound` là import thừa.
- `main.go:258`: `Access-Control-Allow-Credentials: true` không cần, vì đang dùng Bearer header chứ không dùng cookie.
- Media không lưu DB, nên không biết ai upload, upload/xoá không có audit, và editor xoá ảnh đang được bài dùng là bài vỡ ảnh. `ListImages` đọc cả thư mục mỗi request (`media_usecase.go:61-88`).
- `go.mod`: module tên `api`, nên đặt theo đường dẫn repo.
- `Login` không chuẩn hoá email (trim/lowercase) như lúc `Register`.

### 3.4 Về kiến trúc DDD: thừa hay thiếu?

**Thừa so với quy mô.** Khoảng 8k dòng Go cho một CMS CRUD, nhưng mỗi entity có tới 4 bản sao: `Post` (field private) → `PostData` → `PostModel` → `postResponse`, cộng `ReconstitutePost`/`Data()` copy từng con trỏ. Thêm một field là phải sửa ở 6–8 chỗ. `UnitOfWork` gom tất cả repo vào mọi transaction.

**Thiếu ở chỗ quan trọng:**

- Không có `context`.
- Quy tắc chuyển trạng thái bài viết rải ở usecase thay vì nằm gọn trong aggregate. H1, H2, H3 đều là hệ quả của việc này.
- Không có migration versioned.

**Chuẩn BKTech không phải DDD.** api.bkgiaoduc dùng `handlers → services → repos`: repo gồm 3 file `<x>.go` (cache wrapper), `<x>_sql.go`, `<x>_interface.go`, service theo mẫu `I<X>Service`/`New<X>Service`. api.register-account chia package phẳng theo chức năng.

**Khuyến nghị:** **không viết lại**. Giữ cấu trúc hiện tại cho ai_cms, nhưng khi vào dự án thì theo đúng pattern của repo đó. Quy tắc BKTech: "tìm pattern có sẵn trước khi tạo cái mới; nhất quán hơn mới lạ".

### 3.5 Điểm làm tốt (giữ nguyên)

- Mỗi lỗi có test chứng minh. Integration test MySQL chạy trong CI bằng service container.
- `errorRules` không lộ lỗi DB ra client (`errors.go:24-83`).
- Upload ảnh phòng thủ nhiều lớp.
- `config.load(getenv)` có thể test được, fail nhanh khi thiếu secret.
- Graceful shutdown chờ cả scheduler.
- Có OpenAPI (`/docs`).
- README có Quick start, Troubleshooting, Tests.
- `.env.example` đủ biến, có chú thích.

### 3.6 Về quy trình làm việc

Commit trên `main` có những tên như `update` (2 lần), `cai thien 3`, `fea 1 menu tag cate`, `necessary changes for merge first 2 issues`. Chuẩn BKTech:

- Commit: `type(TICKET): mô tả tiếng Việt`, ví dụ `fix(TTT-1395): …`. Mỗi commit một ý.
- Branch: gắn ticket Jira, ví dụ `fix/ttt-1395-<slug>`.
- Mọi thứ vào nhánh chính **qua PR có review**.
- Repo prod chỉ mở PR, người có quyền merge và deploy. Prod và test là hai nhánh khác nhau, sửa song song hai nhánh `fix/x` và `fix/x-main`, "soi từng file, không cherry-pick mù".
- Trước khi báo xong: **chạy binary thật** (route gin trùng thì panic lúc start, `go build` không bắt được), chạy lint và test cả package.

---

## 4. Lộ trình 2 tuần

### Tuần 1: Đúng và an toàn

- [ ] **Ngày 1 – Dọn git (H5).**
  - [ ] Xác nhận đã đổi mật khẩu DB của `.env` cũ.
  - [ ] Chọn `main` hoặc `master` làm nhánh duy nhất, set default.
  - [ ] Đóng PR #4–#11 kèm ghi chú.
  - [ ] Bật branch protection (CI xanh + 1 review, cấm force-push).
- [ ] **Ngày 1 – Hỏi PM:** author sửa bài đã đăng thì giữ published hay quay về review? Ghi câu trả lời vào README phần nghiệp vụ.
- [ ] **Ngày 2 – H1 + H2 + H3.** Gom chuyển trạng thái bài viết vào một method domain. Mỗi lỗi một test đỏ trước, xanh sau. PR mô tả 3 ý: lỗi gì, sửa thế nào, thử ra sao.
- [ ] **Ngày 3–4 – H4 migration.**
  - [ ] `migrations/0001_init.sql` khớp schema hiện tại.
  - [ ] `make migrate` chạy riêng, bỏ AutoMigrate và `ALTER` lúc start.
  - [ ] Thử trên DB trống **và** trên DB có dữ liệu cũ.
- [ ] **Ngày 5 – M2 scheduler** (update có điều kiện + `RowsAffected`) **+ M3** `PUBLIC_BASE_URL`. Viết test chạy 2 lần `PublishDuePosts` song song.

### Tuần 2: Chuẩn production và chuẩn BKTech

- [ ] **Ngày 6–7 – M1 `context.Context`** xuyên handler → usecase → repo (`db.WithContext`). Mỗi tầng một commit nhỏ.
- [ ] **Ngày 8 – M4 + M5:**
  - [ ] Chuyển sang `bcrypt`/`argon2id`, có tương thích hash cũ bằng cách re-hash khi login.
  - [ ] Rate limit `/api/auth/*`.
  - [ ] Mọi lỗi qua `writeError`, mọi ID qua `parseIDParam`.
- [ ] **Ngày 9 – M9 + M10:**
  - [ ] Logger có cấu trúc + request-id, `GIN_MODE=release`, thêm timeouts cho server.
  - [ ] CI thêm `gofmt -l`, `go vet`, `golangci-lint`. Nâng Go ≥ 1.23.
  - [ ] Test handler `httptest` cho các route quyền.
- [ ] **Ngày 10 – M7 FE:**
  - [ ] `axios.create()` + bỏ 16 header thủ công.
  - [ ] Thêm Vitest + test `installSessionHandling` + test form sửa bài.
  - [ ] Thêm script `typecheck` và `test`.
- [ ] **Cuối tuần 2 – Bài tập "giả lập BKTech":**
  - Lấy 1 lỗi trong danh sách trên, đặt mã giả `CMS-001`.
  - Branch `fix/cms-001-<slug>`, commit `fix(CMS-001): …`, PR có mục "Migration (nếu có) + thứ tự chạy", "Cách thử".
  - Chạy binary thật, chụp kết quả.
  - Mentor review như PR thật.
- [ ] (Nếu PM chốt) **M6**: refresh token hoặc autosave nháp.

**Tiêu chí "join dự án":** Hết H1–H5. CI có lint + test. Ít nhất 3 PR liên tiếp đúng format, không phải sửa quy trình.

---

## 5. Chuẩn BKTech cần học trước khi join

1. **Jira và git flow.**
   - Mọi nhánh gắn ticket (`LMS-xxx`, `TTT-xxx`, FE có `QTTH-xxx`). Mỗi ticket một worktree.
   - Biết nhánh test và nhánh prod của từng repo: hệ Thi `thi_dev → thi_prod`, LMS `elearning → master`. Hai nhánh này **code khác nhau**.
   - Ticket xong thì chuyển trạng thái, dọn worktree và nhánh.
2. **Commit và PR.** `type(TICKET): mô tả tiếng Việt`, định danh code giữ tiếng Anh. PR prod ghi rõ migration và thứ tự chạy (migration trước binary).
3. **Kiến trúc Go:**
   - handler → service → repo.
   - Không gọi repo thẳng từ handler khi đã có service.
   - Ghi qua cache wrapper.
   - Không `panic` trong request.
   - Wrap lỗi bằng `%w`.
4. **API contract:**
   - Envelope `{"data":…}` / `{"errors":[{code,message,detail}]}`.
   - Phân trang `page_id` + `per_page`.
   - Header `Authorization` chứa token thô, **không có `Bearer`** (ở api.bkgiaoduc).
5. **DB.** Migration là file `.sql` đặt theo ticket, không sửa file đã chạy, không AutoMigrate. Transaction viết tay trong `*_sql.go`. Dữ liệu cá nhân (email/phone/tên) được mã hoá khi lưu.
6. **Config, secret, log.** `.env.example` có, `.env` không bao giờ commit. logrus `WithFields`, không log password/token/OTP.
7. **Test.**
   - Test chạy offline được, theo kiểu api.register-account: `httptest` + `miniredis`, `go test ./... -race -count=1`.
   - Một hàm test mỗi tình huống, tên nói rõ kết quả, ví dụ `TestRegisterSendOTP_Partial_409NeedsSync`.
8. **Frontend (web.thoikhoabieu).**
   - Một `axiosClient` chung, TanStack Query cho server state.
   - Đọc spec module trước khi sửa `src/`.
   - Lint + Prettier + Husky, **cấm `--no-verify`**.
   - Vitest + Playwright, kiểm UI ở desktop và 390px.
   - Text UI qua i18n.
9. **Tinh thần chung:** đọc trước khi sửa, diff tối thiểu, không refactor ké, dùng dependency có sẵn, luồng mới không được ảnh hưởng luồng cũ.

---

## 6. Kết quả lệnh đã chạy

Môi trường: macOS, `go1.23.0` (goenv). Mình review trên một worktree tạm của `origin/main@08e9464`, không sửa repo của em.

> Ghi chú: máy review có `go` Homebrew 1.27 lệch `GOROOT` của goenv 1.23, nên lần chạy đầu báo `compile: version "go1.23.0" does not match go tool version "go1.27.0"`. Đây là lỗi môi trường của máy review, **không phải lỗi repo**. Chạy lại bằng go 1.23.0 thì sạch.

```text
$ gofmt -l .                       → (rỗng)  ✅
$ go build ./...                   → exit 0  ✅
$ go vet ./...                     → exit 0  ✅
$ go test -race -cover -count=1 ./...
  api/cmd/server                          coverage: 0.0%
  ok  api/internal/config                 coverage: 94.4%
  api/internal/infrastructure/storage     coverage: 0.0%
  ok  api/internal/delivery/http          coverage: 24.8%
  ok  api/internal/delivery/http/middleware coverage: 73.8%
  ok  api/internal/domain                 coverage: 89.1%
  ok  api/internal/infrastructure/database coverage: 28.6%
  ok  api/internal/infrastructure/persistence coverage: 1.8%  (MySQL test skip vì không có MYSQL_TEST_DSN)
  ok  api/internal/infrastructure/security coverage: 23.4%
  ok  api/internal/usecase                coverage: 57.1%
  → tất cả PASS, không có data race ✅
```

Test tạm để xác nhận H1–H3 (đã xoá sau khi chạy):

```text
edit scheduled w/o scheduled_at -> thời gian hẹn đăng phải ở tương lai          (H2)
author resend published         -> vai trò hiện tại không có quyền xuất bản bài viết (H1)
editor create published         -> PublishedAt=<nil>                               (H3)
```

Kiểm tra khác:

- `git ls-files | grep env` → chỉ còn `.env.example` ✅. lịch sử git còn secret cũ, liên hệ mentor ⚠️.
- GitHub Actions (`gh run list`): các run gần nhất trên `main`/`features` đều **success** ✅.
- `gh pr list`: PR #13 và #15 merge với **0 review**. PR #4–#11 còn OPEN. Merge commit của PR #2 (`9e41c98`) không còn trong `master` lẫn `main` ⚠️.
- **Chưa chạy** lint/typecheck/build frontend: máy review không có `web/node_modules` và không được cài package. Em tự chạy `npm ci && npm run lint && npx tsc --noEmit && npm run build` rồi dán kết quả vào PR tiếp theo.
- **Chưa chạy** binary thật hay docker compose (cần MySQL). Phần "chạy thật" để em làm trong lộ trình.

---

Em đã làm được nhiều trong thời gian ngắn, và cách em viết test chứng minh cho từng lỗi là thói quen rất đáng giữ. Hai tuần tới tập trung vào **đúng nghiệp vụ, migration, context và quy trình git**. Mấy thứ này không hào nhoáng, nhưng là thứ quyết định một người có được giao code chạy production hay không. Có gì chưa rõ thì hỏi trước khi làm.
