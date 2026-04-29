package common

import "testing"

func TestPage2LimitOffset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		size       int
		wantLimit  int
		wantOffset int
	}{
		{"normal input", 1, 10, 10, 0},
		{"page is 0", 0, 10, 10, 0},
		{"page is negative", -5, 10, 10, 0},
		{"size is 0", 1, 0, 10, 0},
		{"size is negative", 1, -3, 10, 0},
		{"size exceeds max", 1, 200, 100, 0},
		{"page 3 size 20", 3, 20, 20, 40},
		{"page 2 size 50", 2, 50, 50, 50},
		{"page 5 size 100", 5, 100, 100, 400},
		{"page 5 size 150", 5, 150, 100, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset := Page2LimitOffset(tt.page, tt.size)
			if limit != tt.wantLimit {
				t.Errorf("Page2LimitOffset() limit = %v, want %v", limit, tt.wantLimit)
			}
			if offset != tt.wantOffset {
				t.Errorf("Page2LimitOffset() offset = %v, want %v", offset, tt.wantOffset)
			}
		})
	}
}
