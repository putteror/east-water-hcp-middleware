package schema

// PaginationQuery maps query parameters for list endpoints
type PaginationQuery struct {
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Sort  string `form:"sort" binding:"omitempty"`
	Order string `form:"order" binding:"omitempty,oneof=asc desc ASC DESC"`
}

// SetDefaults applies standard pagination defaults if omitted
func (p *PaginationQuery) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
}

// PaginationData represents the page metadata in list responses
type PaginationData struct {
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
	TotalPages  int `json:"total_pages"`
	TotalRows   int `json:"total_rows"`
}

// IDsRequest maps a list of IDs for batch operations
type IDsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}
