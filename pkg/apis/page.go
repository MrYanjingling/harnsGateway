package apis

const (
	// DefaultPage is the default page number when none is provided.
	DefaultPage = 1
	// DefaultPageSize is the default page size when none is provided.
	DefaultPageSize = 10
	// MaxPageSize caps the page size to protect the backend.
	MaxPageSize = 100
)

// PageQuery is the shared pagination request, embeddable into any query DTO.
type PageQuery struct {
	Page     int `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int `json:"pageSize" form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// Normalize fills in defaults and clamps the page size into range.
func (q *PageQuery) Normalize() {
	if q.Page < 1 {
		q.Page = DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
}

// Offset returns the SQL OFFSET for the current page.
func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// Limit returns the SQL LIMIT for the current page.
func (q PageQuery) Limit() int {
	return q.PageSize
}

// PageResp is the shared pagination response, carried in response.Result.Data.
type PageResp[T any] struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	List     []T   `json:"list"`
}

// NewPageResp builds a PageResp from a page of items and the total count.
func NewPageResp[T any](list []T, total int64, q PageQuery) *PageResp[T] {
	return &PageResp[T]{
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
		List:     list,
	}
}
