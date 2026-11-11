-- 0002: bỏ trạng thái 'approved' khỏi enum posts.status (domain đã bỏ trạng thái này).
-- Bài (và bản revision) đang ở 'approved' quay về 'review' để editor duyệt lại, không tự đăng.
-- Phải đổi dữ liệu TRƯỚC khi đổi enum, nếu không ALTER sẽ lỗi hoặc làm hỏng giá trị.

UPDATE posts SET status = 'review' WHERE status = 'approved';
UPDATE revisions SET status = 'review' WHERE status = 'approved';

ALTER TABLE posts
    MODIFY COLUMN status ENUM('draft','review','rejected','published','scheduled') NOT NULL DEFAULT 'draft';