package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MatchStatus string

const (
	MatchStatusUpcoming  MatchStatus = "upcoming"
	MatchStatusLive      MatchStatus = "live"
	MatchStatusFinished  MatchStatus = "finished"
	MatchStatusCancelled MatchStatus = "cancelled"
)

type Team struct {
	Name      string  `bson:"name" json:"name"`
	Logo      string  `bson:"logo" json:"logo"`
	Odds      float64 `bson:"odds" json:"odds"`
	TotalBets float64 `bson:"total_bets" json:"total_bets"`
}

type Match struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamA      Team               `bson:"team_a" json:"team_a"`
	TeamB      Team               `bson:"team_b" json:"team_b"`
	Tournament string             `bson:"tournament" json:"tournament"`
	Format     string             `bson:"format" json:"format"`
	Status     MatchStatus        `bson:"status" json:"status"`
	StartTime  time.Time          `bson:"start_time" json:"start_time"`
	Winner     string             `bson:"winner,omitempty" json:"winner,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateMatchRequest struct {
	TeamAName  string    `json:"team_a_name"`
	TeamALogo  string    `json:"team_a_logo"`
	TeamBName  string    `json:"team_b_name"`
	TeamBLogo  string    `json:"team_b_logo"`
	Tournament string    `json:"tournament"`
	Format     string    `json:"format"`
	StartTime  time.Time `json:"start_time"`
}
