package services

import (
	"context"
	"log"
	"math"
	"time"

	"cs2-betting-platform/config"
	"cs2-betting-platform/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OddsService struct{}

func NewOddsService() *OddsService {
	return &OddsService{}
}

func (s *OddsService) CalculateInitialOdds() (float64, float64) {
	return 2.0, 2.0
}

func (s *OddsService) RecalculateOdds(matchID primitive.ObjectID) error {
	ctx := context.Background()

	var match models.Match
	matchColl := config.GetCollection("matches")
	err := matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		return err
	}

	betsColl := config.GetCollection("bets")
	cursor, err := betsColl.Find(ctx, bson.M{"match_id": matchID, "status": models.BetStatusPending})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var totalTeamA, totalTeamB float64

	for cursor.Next(ctx) {
		var bet models.Bet
		if err := cursor.Decode(&bet); err != nil {
			continue
		}

		if bet.SelectedTeam == "team_a" {
			totalTeamA += bet.Amount
		} else {
			totalTeamB += bet.Amount
		}
	}

	// Обновляем total_bets в матче
	match.TeamA.TotalBets = totalTeamA
	match.TeamB.TotalBets = totalTeamB

	// Расчет коэффициентов по формуле:
	// Коэффициент = (Total Pool / Team Bets) * margin
	// Margin = 0.95 (5% комиссия букмекера)

	totalPool := totalTeamA + totalTeamB
	margin := 0.95

	var newOddsA, newOddsB float64

	if totalPool > 0 {
		if totalTeamA > 0 {
			newOddsA = (totalPool / totalTeamA) * margin
		} else {
			newOddsA = 2.0 // дефолтный коэффициент
		}

		if totalTeamB > 0 {
			newOddsB = (totalPool / totalTeamB) * margin
		} else {
			newOddsB = 2.0
		}

		// Ограничиваем коэффициенты разумными пределами
		newOddsA = math.Max(1.01, math.Min(newOddsA, 50.0))
		newOddsB = math.Max(1.01, math.Min(newOddsB, 50.0))
	} else {
		// Если ставок нет, используем начальные коэффициенты
		newOddsA, newOddsB = s.CalculateInitialOdds()
	}

	// Округляем до 2 знаков
	match.TeamA.Odds = math.Round(newOddsA*100) / 100
	match.TeamB.Odds = math.Round(newOddsB*100) / 100

	// Обновляем матч в БД
	match.UpdatedAt = time.Now()
	_, err = matchColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{"$set": match},
	)

	if err != nil {
		return err
	}

	log.Printf("📊 Odds recalculated for match %s: TeamA %.2f, TeamB %.2f",
		matchID.Hex(), match.TeamA.Odds, match.TeamB.Odds)

	return nil
}

// StartOddsUpdateWorker запускает фоновый воркер для обновления коэффициентов
// Это goroutine, который периодически пересчитывает коэффициенты для активных матчей
func (s *OddsService) StartOddsUpdateWorker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Println("🔄 Odds update worker started")

		for range ticker.C {
			ctx := context.Background()
			matchColl := config.GetCollection("matches")

			// Находим все активные матчи (upcoming и live)
			cursor, err := matchColl.Find(ctx, bson.M{
				"status": bson.M{"$in": []models.MatchStatus{
					models.MatchStatusUpcoming,
					models.MatchStatusLive,
				}},
			})

			if err != nil {
				log.Printf("Error finding active matches: %v", err)
				continue
			}

			for cursor.Next(ctx) {
				var match models.Match
				if err := cursor.Decode(&match); err != nil {
					continue
				}

				// Пересчитываем коэффициенты для каждого матча
				if err := s.RecalculateOdds(match.ID); err != nil {
					log.Printf("Error recalculating odds for match %s: %v", match.ID.Hex(), err)
				}
			}
			cursor.Close(ctx)
		}
	}()
}
