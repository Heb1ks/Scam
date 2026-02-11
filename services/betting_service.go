package services

import (
	"context"
	"errors"
	"time"

	"Scam/config"
	"Scam/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BettingService struct {
	oddsService *OddsService
}

func NewBettingService(oddsService *OddsService) *BettingService {
	return &BettingService{
		oddsService: oddsService,
	}
}

// Размещение ставки
func (s *BettingService) PlaceBet(userID, matchID primitive.ObjectID, team string, amount float64) (*models.Bet, error) {
	ctx := context.Background()
	usersColl := config.GetCollection("users")
	matchesColl := config.GetCollection("matches")
	betsColl := config.GetCollection("bets")

	// Проверяем пользователя
	var user models.User
	err := usersColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Проверяем баланс
	if user.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	// Проверяем матч
	var match models.Match
	err = matchesColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		return nil, errors.New("match not found")
	}

	// Проверяем статус матча
	if match.Status != models.MatchStatusUpcoming {
		return nil, errors.New("betting is only allowed for upcoming matches")
	}

	// Определяем коэффициент
	var odds float64
	if team == "team_a" {
		odds = match.TeamA.Odds
	} else if team == "team_b" {
		odds = match.TeamB.Odds
	} else {
		return nil, errors.New("invalid team")
	}

	potentialWin := round2(amount * odds)

	// Создаём ставку
	bet := models.Bet{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		MatchID:      matchID,
		Team:         team,
		Amount:       amount,
		Odds:         odds,
		PotentialWin: potentialWin,
		Status:       models.BetStatusPending,
		MatchInfo:    match.TeamA.Name + " vs " + match.TeamB.Name,
		CreatedAt:    time.Now(),
	}

	// Начинаем транзакцию (атомарная операция)
	//  Вычитаем деньги у пользователя
	_, err = usersColl.UpdateOne(
		ctx,
		bson.M{"_id": userID, "balance": bson.M{"$gte": amount}},
		bson.M{
			"$inc": bson.M{
				"balance":      -amount,
				"totalBets":    1,
				"totalWagered": amount,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		return nil, errors.New("failed to update user balance")
	}

	//  Добавляем ставку в базу
	_, err = betsColl.InsertOne(ctx, bet)
	if err != nil {
		// Откатываем баланс пользователя
		usersColl.UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{
				"$inc": bson.M{
					"balance":      amount,
					"totalBets":    -1,
					"totalWagered": -amount,
				},
			},
		)
		return nil, errors.New("failed to place bet")
	}

	//  Обновляем сумму ставок на команду в матче
	updateField := "teamA.totalBets"
	if team == "team_b" {
		updateField = "teamB.totalBets"
	}

	matchesColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{
			"$inc": bson.M{
				updateField:       amount,
				"totalBetsAmount": amount,
				"betsCount":       1,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)

	//  Пересчитываем коэффициенты (они меняются динамически!)
	s.oddsService.UpdateMatchOdds(matchID)

	return &bet, nil
}

// Расчёт ставок по завершению матча
func (s *BettingService) SettleMatch(matchID primitive.ObjectID, winner string) error {
	ctx := context.Background()
	matchesColl := config.GetCollection("matches")
	betsColl := config.GetCollection("bets")
	usersColl := config.GetCollection("users")

	// Обновляем статус матча
	_, err := matchesColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{
			"$set": bson.M{
				"status":    models.MatchStatusFinished,
				"winner":    winner,
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}

	// Если матч отменён - возвращаем деньги
	if winner == "cancelled" {
		cursor, _ := betsColl.Find(ctx, bson.M{
			"matchId": matchID,
			"status":  models.BetStatusPending,
		})

		var bets []models.Bet
		cursor.All(ctx, &bets)
		cursor.Close(ctx)

		settleTime := time.Now()

		for _, bet := range bets {
			// Возвращаем деньги
			usersColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.UserID},
				bson.M{
					"$inc": bson.M{
						"balance":      bet.Amount,
						"totalWagered": -bet.Amount,
						"totalBets":    -1,
					},
				},
			)

			// Обновляем статус ставки
			betsColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.ID},
				bson.M{
					"$set": bson.M{
						"status":    models.BetStatusCancelled,
						"settledAt": &settleTime,
					},
				},
			)
		}

		return nil
	}

	// Обрабатываем выигрышные и проигрышные ставки
	cursor, _ := betsColl.Find(ctx, bson.M{
		"matchId": matchID,
		"status":  models.BetStatusPending,
	})

	var bets []models.Bet
	cursor.All(ctx, &bets)
	cursor.Close(ctx)

	settleTime := time.Now()

	for _, bet := range bets {
		if bet.Team == winner {
			// ВЫИГРЫШ
			winnings := bet.PotentialWin

			usersColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.UserID},
				bson.M{
					"$inc": bson.M{
						"balance":       winnings,
						"wonBets":       1,
						"totalWinnings": winnings,
						"profitLoss":    winnings - bet.Amount,
					},
				},
			)

			betsColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.ID},
				bson.M{
					"$set": bson.M{
						"status":         models.BetStatusWon,
						"actualWinnings": winnings,
						"settledAt":      &settleTime,
					},
				},
			)
		} else {
			// ПРОИГРЫШ
			usersColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.UserID},
				bson.M{
					"$inc": bson.M{
						"lostBets":   1,
						"profitLoss": -bet.Amount,
					},
				},
			)

			betsColl.UpdateOne(
				ctx,
				bson.M{"_id": bet.ID},
				bson.M{
					"$set": bson.M{
						"status":    models.BetStatusLost,
						"settledAt": &settleTime,
					},
				},
			)
		}
	}

	return nil
}

// Получение ставок пользователя
func (s *BettingService) GetUserBets(userID primitive.ObjectID) ([]models.Bet, error) {
	ctx := context.Background()
	betsColl := config.GetCollection("bets")

	cursor, err := betsColl.Find(ctx, bson.M{
		"userId": userID,
	}, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bets []models.Bet
	err = cursor.All(ctx, &bets)
	return bets, err
}
