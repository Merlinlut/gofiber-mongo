package model

// MetaInfo -> informasi pagination & filter
type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}
type UserResponse struct {
	User []User `json:"user"`
	Meta MetaInfo `json:"meta"` 
}

// APIResponse digunakan untuk format respons umum
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// Generic paged response for a slice of any data (in Go generics not needed here,
// we'll compose concrete responses in services)
