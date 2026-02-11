package services

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"Scam/config"
	"Scam/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OddsService struct {
	mu sync.RWMutex
}

func NewOddsService() *OddsService {
	return &OddsService{}
}

// ============================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// ============================================
// АЛГОРИТМ ПОДСЧЁТА КОЭФФИЦИЕНТОВ (VALVE POINTS)
// ============================================

// Вычисляет ожидаемую вероятность победы на основе Valve Points
func valveExpectedScore(valveA, valveB float64) float64 {
	// Valve использует диапазон ~1500-2000
	// Используем 250 для большей чувствительности к разнице в очках
	return 1.0 / (1.0 + math.Pow(10, (valveB-valveA)/250.0))
}

// Рассчитывает вероятности победы команд с учётом всех факторов
func (s *OddsService) CalculateProbabilities(teamA, teamB models.Team) (float64, float64) {
	// Веса факторов
	wValve := 0.60 // Valve Points (главный фактор)
	wForm := 0.25  // Форма команды
	wMap := 0.15   // Сила на картах

	valveA := float64(teamA.ValvePoints)
	valveB := float64(teamB.ValvePoints)

	// Fallback на HLTV если нет Valve Points
	if valveA == 0 {
		valveA = 2000 - math.Log(float64(teamA.HLTVRank))*100
		if valveA < 1500 {
			valveA = 1500
		}
	}
	if valveB == 0 {
		valveB = 2000 - math.Log(float64(teamB.HLTVRank))*100
		if valveB < 1500 {
			valveB = 1500
		}
	}

	scoreValveA := valveExpectedScore(valveA, valveB)

	// Компонент формы
	formPowerA := float64(teamA.FormWins) / 5.0
	formPowerB := float64(teamB.FormWins) / 5.0

	// Корректировка: если команда слабее по Valve, но форма хорошая - это ценнее
	if valveA < valveB {
		formPowerA *= 1.15
	} else {
		formPowerB *= 1.15
	}

	if formPowerA+formPowerB == 0 {
		formPowerA = 0.5
		formPowerB = 0.5
	}
	scoreFormA := formPowerA / (formPowerA + formPowerB)

	// Компонент карт
	mapPowerA := (teamA.MapPool.Mirage + teamA.MapPool.Nuke + teamA.MapPool.Ancient) / 3.0
	mapPowerB := (teamB.MapPool.Mirage + teamB.MapPool.Nuke + teamB.MapPool.Ancient) / 3.0

	// Корректировка: сильная команда с хорошими картами получает бонус
	if valveA > valveB && mapPowerA > 0.6 {
		mapPowerA *= 1.1
	}
	if valveB > valveA && mapPowerB > 0.6 {
		mapPowerB *= 1.1
	}

	if mapPowerA+mapPowerB == 0 {
		mapPowerA = 0.5
		mapPowerB = 0.5
	}
	scoreMapA := mapPowerA / (mapPowerA + mapPowerB)

	// Финальный расчёт
	pA := wValve*scoreValveA + wForm*scoreFormA + wMap*scoreMapA
	pA = clamp01(pA)
	pB := 1 - pA

	return pA, pB
}

// ============================================
// ДИНАМИЧЕСКИЕ КОЭФФИЦИЕНТЫ (НОВОЕ!)
// ============================================

// Рассчитывает коэффициенты с учётом ставок (динамика)
func (s *OddsService) CalculateDynamicOdds(pA, pB, totalBetsA, totalBetsB, margin float64) (float64, float64) {
	totalBets := totalBetsA + totalBetsB

	if totalBets == 0 {
		// Нет ставок - используем базовые коэффициенты
		pA_m := pA * (1 + margin)
		pB_m := pB * (1 + margin)
		return round2(1 / pA_m), round2(1 / pB_m)
	}

	// Процент ставок на каждую команду
	percentA := totalBetsA / totalBets
	percentB := totalBetsB / totalBets

	// Корректируем вероятности на основе распределения ставок
	// Если на команду A поставили много - её коэффициент падает
	// Используем логарифмическую шкалу для плавности
	betInfluence := 0.25 // сила влияния ставок (25%)

	adjustedPA := pA + (percentB-percentA)*betInfluence
	adjustedPB := pB + (percentA-percentB)*betInfluence

	// Нормализуем обратно к 1.0
	sum := adjustedPA + adjustedPB
	adjustedPA /= sum
	adjustedPB /= sum

	// Применяем маржу
	pA_m := adjustedPA * (1 + margin)
	pB_m := adjustedPB * (1 + margin)

	oddA := round2(1 / pA_m)
	oddB := round2(1 / pB_m)

	// Минимальный коэффициент 1.01
	if oddA < 1.01 {
		oddA = 1.01
	}
	if oddB < 1.01 {
		oddB = 1.01
	}

	return oddA, oddB
}

// Рассчитывает начальные коэффициенты для нового матча
func (s *OddsService) CalculateInitialOdds() (float64, float64) {
	// Базовые коэффициенты 50/50 с 6% маржой
	margin := 0.06
	p := 0.5
	pWithMargin := p * (1 + margin)
	odds := round2(1 / pWithMargin)
	return odds, odds
}

// ============================================
// ОБНОВЛЕНИЕ КОЭФФИЦИЕНТОВ В РЕАЛЬНОМ ВРЕМЕНИ
// ============================================

// Обновляет коэффициенты для всех активных матчей
func (s *OddsService) UpdateMatchOdds(matchID primitive.ObjectID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	matchColl := config.GetCollection("matches")
	teamsColl := config.GetCollection("teams")

	// Получаем матч
	var match models.Match
	err := matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		return err
	}

	// Получаем полные данные команд
	var teamA, teamB models.Team
	teamsColl.FindOne(ctx, bson.M{"_id": match.TeamA.ID}).Decode(&teamA)
	teamsColl.FindOne(ctx, bson.M{"_id": match.TeamB.ID}).Decode(&teamB)

	// Рассчитываем базовые вероятности
	pA, pB := s.CalculateProbabilities(teamA, teamB)

	// Применяем динамику на основе ставок
	margin := 0.06
	oddA, oddB := s.CalculateDynamicOdds(pA, pB, match.TeamA.TotalBets, match.TeamB.TotalBets, margin)

	// Обновляем коэффициенты в базе
	_, err = matchColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{
			"$set": bson.M{
				"teamA.odds": oddA,
				"teamB.odds": oddB,
				"updatedAt":  time.Now(),
			},
		},
	)

	return err
}

// Воркер для автоматического обновления коэффициентов
func (s *OddsService) StartOddsUpdateWorker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("🔄 Odds update worker started (interval: %v)", interval)

		for range ticker.C {
			ctx := context.Background()
			matchColl := config.GetCollection("matches")

			// Обновляем коэффициенты только для upcoming и live матчей
			cursor, err := matchColl.Find(ctx, bson.M{
				"status": bson.M{
					"$in": []string{"upcoming", "live"},
				},
			})
			if err != nil {
				continue
			}

			var matches []models.Match
			cursor.All(ctx, &matches)
			cursor.Close(ctx)

			for _, match := range matches {
				s.UpdateMatchOdds(match.ID)
			}
		}
	}()
}
