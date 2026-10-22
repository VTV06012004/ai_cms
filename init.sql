-- Tạo các danh mục mặc định
INSERT INTO categories (id, name, slug, description, created_at, updated_at) VALUES 
(1, 'Công nghệ', 'cong-nghe', 'Các bài viết về lập trình và công nghệ', NOW(), NOW()),
(2, 'Đời sống', 'doi-song', 'Chia sẻ góc nhìn cuộc sống', NOW(), NOW()),
(3, 'Review', 'review', 'Đánh giá sản phẩm, sách, phim', NOW(), NOW());

-- Bạn cũng có thể seed luôn tài khoản admin ở đây nếu muốn (đảm bảo mật khẩu đã được mã hóa bcrypt nhé, hoặc tạo qua code Go seeder)