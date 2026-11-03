package domain

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type PageRequest struct {
	Page    int
	PerPage int
}

// NewPageRequest chuẩn hóa giá trị: page >= 1, 1 <= perPage <= MaxPerPage.
func NewPageRequest(page, perPage int) PageRequest {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return PageRequest{Page: page, PerPage: perPage}
}

func (p PageRequest) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p PageRequest) Info(total int64) PageInfo {
	return PageInfo{Page: p.Page, PerPage: p.PerPage, Total: total}
}

type PageInfo struct {
	Page    int
	PerPage int
	Total   int64
}

func (i PageInfo) TotalPages() int {
	if i.Total == 0 || i.PerPage < 1 {
		return 0
	}
	return int((i.Total + int64(i.PerPage) - 1) / int64(i.PerPage))
}
