package response

type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GenericSuccessResponse[T any] struct {
	GenericResponse
	Data T `json:"data"`
}

type PaginationMeta struct {
	Before  *string `json:"before"`
	After   *string `json:"after"`
	HasNext bool    `json:"has_next"`
	HasPrev bool    `json:"has_prev"`
}

type PagedDataResponse[T any] struct {
	GenericResponse
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code  int         `json:"code"`
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}
