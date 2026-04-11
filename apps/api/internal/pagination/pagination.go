package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Params struct {
	Page   int
	Limit  int
	Offset int
	Search string
}

type Meta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

type Response struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

func ParseParams(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	if page < 1 {
		page = DefaultPage
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultPageSize)))
	if limit < 1 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	search := c.Query("search")

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
		Search: search,
	}
}

func NewMeta(total, page, limit int) Meta {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
