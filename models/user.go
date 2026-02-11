package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Роли пользователей
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username"`
	Email        string             `json:"email" bson:"email"`
	PasswordHash string             `json:"-" bson:"passwordHash"`
	Balance      float64            `json:"balance" bson:"balance"`

	// НОВОЕ ПОЛЕ - РОЛЬ
	Role string `json:"role" bson:"role"` // "user" или "admin"

	// Статистика
	TotalBets     int     `json:"totalBets" bson:"totalBets"`
	WonBets       int     `json:"wonBets" bson:"wonBets"`
	LostBets      int     `json:"lostBets" bson:"lostBets"`
	TotalWagered  float64 `json:"totalWagered" bson:"totalWagered"`
	TotalWinnings float64 `json:"totalWinnings" bson:"totalWinnings"`
	ProfitLoss    float64 `json:"profitLoss" bson:"profitLoss"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type BetStatus string

const (
	BetStatusPending   BetStatus = "pending"
	BetStatusWon       BetStatus = "won"
	BetStatusLost      BetStatus = "lost"
	BetStatusCancelled BetStatus = "cancelled"
)

type Bet struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"userId" bson:"userId"`
	MatchID        primitive.ObjectID `json:"matchId" bson:"matchId"`
	Team           string             `json:"team" bson:"team"`
	Amount         float64            `json:"amount" bson:"amount"`
	Odds           float64            `json:"odds" bson:"odds"`
	PotentialWin   float64            `json:"potentialWin" bson:"potentialWin"`
	Status         BetStatus          `json:"status" bson:"status"`
	ActualWinnings float64            `json:"actualWinnings,omitempty" bson:"actualWinnings,omitempty"`

	// Дополнительная информация для отображения
	MatchInfo string     `json:"matchInfo,omitempty" bson:"matchInfo,omitempty"`
	CreatedAt time.Time  `json:"createdAt" bson:"createdAt"`
	SettledAt *time.Time `json:"settledAt,omitempty" bson:"settledAt,omitempty"`
}

type PlaceBetRequest struct {
	MatchID string  `json:"match_id"`
	Team    string  `json:"team"` // "team_a" или "team_b"
	Amount  float64 `json:"amount"`
}
