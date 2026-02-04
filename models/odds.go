package models

type OddsRequest struct {
	TeamAID string  `json:"teamAId"`
	TeamBID string  `json:"teamBId"`
	Margin  float64 `json:"margin"` // например 0.67
}

type OddsResponse struct {
	TeamA string `json:"teamA"`
	TeamB string `json:"teamB"`

	ProbA float64 `json:"probA"`
	ProbB float64 `json:"probB"`

	OddA float64 `json:"oddA"`
	OddB float64 `json:"oddB"`
}
