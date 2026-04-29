package request

type PaginationRequest struct {
	AfterCursor  string `query:"after" validate:"omitempty"`
	BeforeCursor string `query:"before" validate:"omitempty"`
	Limit        int    `query:"limit" validate:"omitempty,number" default:"50"`
	Sort         string `query:"sort" validate:"omitempty,oneof=asc desc" default:"asc"`
	SortBy       string `query:"sort_by" validate:"omitempty,alphanum" default:"created_at"`
	Search       string `query:"search" validate:"omitempty"`
}
