package domain

import "errors"

var (
	ErrPostNotInReview         = errors.New("chỉ có thể kiểm duyệt bài viết đang chờ duyệt")
	ErrRejectionReasonRequired = errors.New("phải nhập lý do từ chối bài viết")
	ErrInvalidPostStatus       = errors.New("trạng thái không hợp lệ")
	ErrInvalidRole             = errors.New("vai trò không hợp lệ")
	ErrCannotChangeOwnRole     = errors.New("bạn không thể tự thay đổi quyền của chính tài khoản mình")
	ErrCannotChangeRootAdmin   = errors.New("không được phép thay đổi quyền của quản trị viên gốc")
	ErrUserNameRequired        = errors.New("tên người dùng là bắt buộc")
	ErrInvalidEmail            = errors.New("email không hợp lệ")
	ErrPasswordHashRequired    = errors.New("mật khẩu đã mã hóa là bắt buộc")
	ErrPostTitleRequired       = errors.New("tiêu đề bài viết là bắt buộc")
	ErrPostSlugRequired        = errors.New("slug bài viết là bắt buộc")
	ErrPostCategoryRequired    = errors.New("danh mục bài viết là bắt buộc")
	ErrPostNotFound            = errors.New("không tìm thấy bài viết")
	ErrPostAccessDenied        = errors.New("bạn không có quyền thao tác bài viết này")
	ErrPublishNotAllowed       = errors.New("vai trò hiện tại không có quyền xuất bản bài viết")
	ErrPostStatusNotAllowed    = errors.New("vai trò hiện tại không được đặt trạng thái bài viết này")
	ErrCategoryNameRequired    = errors.New("tên danh mục là bắt buộc")
	ErrCategorySlugRequired    = errors.New("slug danh mục là bắt buộc")
	ErrMenuNameRequired        = errors.New("tên menu là bắt buộc")
	ErrMenuItemTitleRequired   = errors.New("tiêu đề mục menu là bắt buộc")
)
