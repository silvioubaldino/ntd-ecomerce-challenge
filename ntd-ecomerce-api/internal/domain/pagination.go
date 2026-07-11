package domain

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PageRequest struct {
	Limit  int
	Cursor *Cursor
}

func DefaultPageRequest() PageRequest {
	return PageRequest{Limit: DefaultLimit}
}

func ValidateLimit(n int) (code string, ok bool) {
	if n < 1 || n > MaxLimit {
		return "must_be_between_1_and_100", false
	}
	return "", true
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
}
