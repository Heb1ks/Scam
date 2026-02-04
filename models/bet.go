package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BetStatus string

const (
	BetStatusPending BetStatus = "pending"
	BetStatusWon     BetStatus = "won"
	BetStatusLost    BetStatus = "lost"
	BetStatusRefund  BetStatus = "refund"
)

type Bet struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	MatchID      primitive.ObjectID `bson:"match_id" json:"match_id"`
	SelectedTeam string             `bson:"selected_team" json:"selected_team"`
	Amount       float64            `bson:"amount" json:"amount"`
	Odds         float64            `bson:"odds" json:"odds"`
	PotentialWin float64            `bson:"potential_win" json:"potential_win"`
	Status       BetStatus          `bson:"status" json:"status"`
	WinAmount    float64            `bson:"win_amount,omitempty" json:"win_amount,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type PlaceBetRequest struct {
	MatchID      string  `json:"match_id"`
	SelectedTeam string  `json:"selected_team"`
	Amount       float64 `json:"amount"`
}

type BetWithDetails struct {
	Bet
	MatchInfo Match  `json:"match_info"`
	Username  string `json:"username"`
}
