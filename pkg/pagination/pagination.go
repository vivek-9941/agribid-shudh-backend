// Package pagination provides cursor/offset pagination helpers.
package pagination

import "math"

// Params holds pagination parameters.
type Params struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// DefaultParams returns default pagination parameters.
func DefaultParams() Params {
	return Params{Page: 1, PageSize: 20}
}

// Parse normalizes pagination parameters.
func Parse(page, pageSize int) Params {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return Params{Page: page, PageSize: pageSize}
}

// Offset calculates the SQL OFFSET from page and page_size.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// TotalPages calculates total pages from total count.
func (p Params) TotalPages(total int64) int {
	if total == 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(p.PageSize)))
}
