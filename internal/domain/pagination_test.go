package domain

import "testing"

func TestNewPageRequestNormalizesInput(t *testing.T) {
	cases := []struct {
		page, perPage      int
		wantPage, wantSize int
	}{
		{0, 0, 1, DefaultPerPage},
		{-3, -1, 1, DefaultPerPage},
		{2, 50, 2, 50},
		{1, 100000, 1, MaxPerPage},
	}
	for _, tc := range cases {
		got := NewPageRequest(tc.page, tc.perPage)
		if got.Page != tc.wantPage || got.PerPage != tc.wantSize {
			t.Errorf("NewPageRequest(%d, %d) = %+v", tc.page, tc.perPage, got)
		}
	}
}

func TestPageRequestOffsetAndTotalPages(t *testing.T) {
	page := NewPageRequest(3, 20)
	if page.Offset() != 40 {
		t.Fatalf("offset = %d, want 40", page.Offset())
	}
	for total, want := range map[int64]int{0: 0, 1: 1, 20: 1, 21: 2, 100: 5} {
		if got := page.Info(total).TotalPages(); got != want {
			t.Errorf("total %d: pages = %d, want %d", total, got, want)
		}
	}
}
