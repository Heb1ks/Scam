package models

type User struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}

type Bet struct {
	ID     int     `json:"id"`
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"` // "pending", "won", "lost"
}
