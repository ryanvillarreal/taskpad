package model

// ListParams holds pagination and sorting parameters for list queries.
type ListParams struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
	Order  string `json:"order"` // "asc" or "desc"
}

// ListResult wraps a paginated list response.
type ListResult[T any] struct {
	Data   []T `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// TodoFilters holds optional filters for listing todos.
type TodoFilters struct {
	Completed *bool
	Priority  *Priority
}

// NoteFilters holds optional filters for listing notes.
type NoteFilters struct {
	Tag    *string
	Search *string
}
