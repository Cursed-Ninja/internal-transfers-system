package storage

// Represents in-mem model for database object
type Account struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}
