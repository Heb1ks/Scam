package services

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"cs2-betting-platform/config"
	"cs2-betting-platform/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BettingService struct {
	oddsService *OddsService
	mu          sync.Mutex
}

func NewBettingService(oddsService *OddsService) *BettingService {
	return &BettingService{
		oddsService: oddsService,
	}
}

func (s *BettingService) PlaceBet(ctx context.Context, userID, matchID primitive.ObjectID, selectedTeam string, amount float64) (*models.Bet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userColl := config.GetCollection("users")
	var user models.User
	err := userColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	matchColl := config.GetCollection("matches")
	var match models.Match
	err = matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		return nil, errors.New("match not found")
	}

	if match.Status != models.MatchStatusUpcoming {
		return nil, errors.New("betting is closed for this match")
	}

	if selectedTeam != "team_a" && selectedTeam != "team_b" {
		return nil, errors.New("invalid team selection")
	}

	var odds float64
	if selectedTeam == "team_a" {
		odds = match.TeamA.Odds
	} else {
		odds = match.TeamB.Odds
	}

	bet := &models.Bet{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		MatchID:      matchID,
		SelectedTeam: selectedTeam,
		Amount:       amount,
		Odds:         odds,
		PotentialWin: amount * odds,
		Status:       models.BetStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	betsColl := config.GetCollection("bets")
	_, err = betsColl.InsertOne(ctx, bet)
	if err != nil {
		return nil, err
	}

	user.Balance -= amount
	user.UpdatedAt = time.Now()
	_, err = userColl.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"balance": user.Balance, "updated_at": user.UpdatedAt}},
	)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Bet placed: User %s, Match %s, Team %s, Amount %.2f, Odds %.2f",
		userID.Hex(), matchID.Hex(), selectedTeam, amount, odds)

	go func() {
		if err := s.oddsService.RecalculateOdds(matchID); err != nil {
			log.Printf("Error recalculating odds: %v", err)
		}
	}()

	return bet, nil
}

func (s *BettingService) SettleMatch(ctx context.Context, matchID primitive.ObjectID, winner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	matchColl := config.GetCollection("matches")
	var match models.Match
	err := matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		return errors.New("match not found")
	}

	if winner != "team_a" && winner != "team_b" && winner != "cancelled" {
		return errors.New("invalid winner")
	}

	match.Status = models.MatchStatusFinished
	match.Winner = winner
	match.UpdatedAt = time.Now()

	_, err = matchColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{"$set": match},
	)
	if err != nil {
		return err
	}

	betsColl := config.GetCollection("bets")
	cursor, err := betsColl.Find(ctx, bson.M{
		"match_id": matchID,
		"status":   models.BetStatusPending,
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	userColl := config.GetCollection("users")

	for cursor.Next(ctx) {
		var bet models.Bet
		if err := cursor.Decode(&bet); err != nil {
			continue
		}

		var newStatus models.BetStatus
		var winAmount float64

		if winner == "cancelled" {

			newStatus = models.BetStatusRefund
			winAmount = bet.Amount
		} else if bet.SelectedTeam == winner {

			newStatus = models.BetStatusWon
			winAmount = bet.PotentialWin
		} else {

			newStatus = models.BetStatusLost
			winAmount = 0
		}
		bet.Status = newStatus
		bet.WinAmount = winAmount
		bet.UpdatedAt = time.Now()

		_, err = betsColl.UpdateOne(
			ctx,
			bson.M{"_id": bet.ID},
			bson.M{"$set": bet},
		)
		if err != nil {
			log.Printf("Error updating bet %s: %v", bet.ID.Hex(), err)
			continue
		}

		if winAmount > 0 {
			_, err = userColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.UserID},
				bson.M{
					"$inc": bson.M{"balance": winAmount},
					"$set": bson.M{"updated_at": time.Now()},
				},
			)
			if err != nil {
				log.Printf("Error updating user balance %s: %v", bet.UserID.Hex(), err)
			}
		}

		log.Printf("🎲 Bet settled: ID %s, Status %s, Win %.2f",
			bet.ID.Hex(), newStatus, winAmount)
	}

	log.Printf("🏁 Match %s settled. Winner: %s", matchID.Hex(), winner)

	return nil
}

func (s *BettingService) GetUserBets(ctx context.Context, userID primitive.ObjectID) ([]models.BetWithDetails, error) {
	betsColl := config.GetCollection("bets")
	matchColl := config.GetCollection("matches")

	cursor, err := betsColl.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.BetWithDetails

	for cursor.Next(ctx) {
		var bet models.Bet
		if err := cursor.Decode(&bet); err != nil {
			continue
		}

		var match models.Match
		err := matchColl.FindOne(ctx, bson.M{"_id": bet.MatchID}).Decode(&match)
		if err != nil {
			continue
		}

		results = append(results, models.BetWithDetails{
			Bet:       bet,
			MatchInfo: match,
		})
	}

	return results, nil
}
