package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type MapPool struct {
	Mirage   float64 `json:"mirage" bson:"mirage"`
	Nuke     float64 `json:"nuke" bson:"nuke"`
	Ancient  float64 `json:"ancient" bson:"ancient"`
	Overpass float64 `json:"overpass" bson:"overpass"`
	Inferno  float64 `json:"inferno" bson:"inferno"`
	Dust2    float64 `json:"dust2" bson:"dust2"`
	Anubis   float64 `json:"anubis" bson:"anubis"`
}

type Team struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TeamName string             `json:"teamName" bson:"teamName"`

	// ОСНОВНЫЕ РЕЙТИНГИ
	ValvePoints int `json:"valvePoints" bson:"valvePoints"` // 🆕 Официальный Valve рейтинг (1500-2000)
	HLTVRank    int `json:"hltvRank" bson:"hltvRank"`       // Fallback если нет Valve Points

	// ФОРМА
	FormWins int `json:"formWins" bson:"formWins"` // последние 5 побед
	FormLoss int `json:"formLoss" bson:"formLoss"` // последние 5 поражений

	// КАРТЫ
	MapPool MapPool `json:"mapPool" bson:"mapPool"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt"`
}
