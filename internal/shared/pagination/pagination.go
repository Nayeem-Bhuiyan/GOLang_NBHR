package pagination

import (
	"math"
	"strconv"

	"nbhr/internal/constants"

	"github.com/gin-gonic/gin"
)

// Params holds incoming pagination parameters.
type Params struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

// Meta holds computed pagination metadata for responses.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// FromContext extracts and validates pagination params from gin context.
func FromContext(c *gin.Context) *Params {
	page := parseIntParam(c.Query("page"), constants.DefaultPage)
	pageSize := parseIntParam(c.Query("page_size"), constants.DefaultPageSize)

	if page < 1 {
		page = constants.DefaultPage
	}
	if pageSize < 1 || pageSize > constants.MaxPageSize {
		pageSize = constants.DefaultPageSize
	}

	order := c.Query("order")
	if order != constants.SortDesc {
		order = constants.SortAsc
	}

	sort := c.Query("sort")
	if sort == "" {
		sort = "created_at"
	}

	return &Params{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
}

// Offset calculates the DB offset.
func (p *Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// OrderClause returns the ORDER BY clause string.
func (p *Params) OrderClause() string {
	return p.Sort + " " + p.Order
}

// NewMeta builds a Meta from total count and params.
func NewMeta(total int64, p *Params) *Meta {
	totalPages := int(math.Ceil(float64(total) / float64(p.PageSize)))
	return &Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}

func parseIntParam(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return v
}