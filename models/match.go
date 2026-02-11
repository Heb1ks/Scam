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

type Match struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TeamA      TeamInMatch        `json:"teamA" bson:"teamA"`
	TeamB      TeamInMatch        `json:"teamB" bson:"teamB"`
	Tournament string             `json:"tournament" bson:"tournament"`
	Format     string             `json:"format" bson:"format"`
	Status     MatchStatus        `json:"status" bson:"status"`
	StartTime  time.Time          `json:"startTime" bson:"startTime"`

	Winner     string `json:"winner,omitempty" bson:"winner,omitempty"`
	FinalScore string `json:"finalScore,omitempty" bson:"finalScore,omitempty"`

	TotalBetsAmount float64 `json:"totalBetsAmount" bson:"totalBetsAmount"`
	BetsCount       int     `json:"betsCount" bson:"betsCount"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type CreateMatchRequest struct {
	TeamAID    string    `json:"teamAId"`
	TeamBID    string    `json:"teamBId"`
	TeamAName  string    `json:"teamAName"`
	TeamALogo  string    `json:"teamALogo"`
	TeamBName  string    `json:"teamBName"`
	TeamBLogo  string    `json:"teamBLogo"`
	Tournament string    `json:"tournament"`
	Format     string    `json:"format"`
	StartTime  time.Time `json:"startTime"`
}
