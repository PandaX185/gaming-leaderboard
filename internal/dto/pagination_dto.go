package dto

type PaginationParams struct {
	Page     int `form:"page" validate:"gte=1" example:"1"`
	PageSize int `form:"page_size" validate:"gte=1,lte=1000" example:"10"`
}

type PaginatedResponse struct {
	TotalItems int  `json:"total_items" example:"100"`
	TotalPages int  `json:"total_pages" example:"10"`
	Page       int  `json:"page" example:"1"`
	PageSize   int  `json:"page_size" example:"10"`
	HasNext    bool `json:"has_next" example:"true"`
	HasPrev    bool `json:"has_prev" example:"false"`
	Items      any  `json:"items"`
}
