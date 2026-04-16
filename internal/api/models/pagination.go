package apiModels

import (
	"math"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

type PaginationParams struct {
	Page  int `form:"page" example:"1"`
	Limit int `form:"limit" example:"10"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = defaultPage
	}
	if p.Limit < 1 || p.Limit > maxPageSize {
		p.Limit = defaultPageSize
	}
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginatedResponse[T any] struct {
	Total      int64 `json:"total_objects" example:"20"`
	Limit      int   `json:"limit_per_page" example:"10"`
	Page       int   `json:"current_page" example:"1"`
	TotalPages int   `json:"total_pages" example:"2"`
	Data       []T   `json:"data"`
}

func NewPaginatedResponse[T any](data []T, total int64, p PaginationParams) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Total:      total,
		Limit:      p.Limit,
		Page:       p.Page,
		TotalPages: int(math.Ceil(float64(total) / float64(p.Limit))),
		Data:       data,
	}
}
