package controllers

import (
	"context"
	"math"
	"net/http"
	"time"

	"Scam/database"
	"Scam/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// Замена обыного рейтнига но Валв поинт систем
// Вычисляет ожидаемую вероятность победы на основе Valve Points
// Формула похожа на ело, но адаптирована под диапазон Valve (1500-2000)
func valveExpectedScore(valveA, valveB float64) float64 {
	// Valve использует диапазон ~1500-2000, поэтому делитель меньше чем в классической Elo (400)
	// Используем 250 для большей чувствительности к разнице в очках
	return 1.0 / (1.0 + math.Pow(10, (valveB-valveA)/250.0))
}

// Нормализует Valve Points для использования в расчётах
// Возвращает значение в диапазоне примерно 0.3-0.7 для команд в топ-30
func normalizeValvePoints(valvePoints float64) float64 {
	// Типичный диапазон Valve: 1500-2000
	// Нормализуем относительно среднего значения (1750)
	normalized := valvePoints / 3500.0 // делим на удвоенное среднее
	return clamp01(normalized)
}

type RecentMatch struct {
	Won              bool    // выиграли или проиграли
	OpponentValve    float64 // Valve Points противника
	RoundsDiff       int     // разница раундов (например, 13-10 = +3)
	ImportanceWeight float64 // вес матча: 1.0 обычный, 1.5 плейофф, 2.0 финал
}

// Рассчитывает скор формы с учётом качества противников
func calculateFormScore(recentMatches []RecentMatch, teamValve float64) float64 {
	if len(recentMatches) == 0 {
		return 0.5 // нейтральная форма
	}

	totalScore := 0.0
	totalWeight := 0.0

	for i, match := range recentMatches {
		// Более свежие матчи важнее (экспоненциальный вес)
		recencyWeight := math.Exp(-0.2 * float64(i)) // последний матч вес=1.0

		// Качество победы/поражения зависит от силы противника
		expectedWin := valveExpectedScore(teamValve, match.OpponentValve)

		var matchScore float64
		if match.Won {
			// Победа над сильным противником ценнее
			// Если ожидали проиграть (expectedWin < 0.5), но выиграли - это круто
			matchScore = 0.5 + 0.5*(1.0-expectedWin)
		} else {
			// Поражение от слабого противника хуже
			// Если ожидали выиграть (expectedWin > 0.5), но проиграли - это плохо
			matchScore = 0.5 * expectedWin
		}

		// Учитываем разницу раундов (доминирование)
		// tanh нормализует в диапазон -1 до 1, умножаем на 0.1 для небольшой корректировки
		roundBonus := math.Tanh(float64(match.RoundsDiff)/10.0) * 0.1
		matchScore += roundBonus
		matchScore = clamp01(matchScore)

		// Применяем веса
		weight := recencyWeight * match.ImportanceWeight
		totalScore += matchScore * weight
		totalWeight += weight
	}

	return totalScore / totalWeight
}

func calculateProbabilityValve(teamA, teamB models.Team) (float64, float64) {
	wValve := 0.60 // Valve Points (главный фактор, т.к. уже учитывает много)
	wForm := 0.25  // Форма команды
	wMap := 0.15   // Сила на картах

	valveA := float64(teamA.ValvePoints)
	valveB := float64(teamB.ValvePoints)

	// Если Valve Points = 0, используем fallback на HLTV rank
	if valveA == 0 {
		// Приблизительная конвертация: топ-1 ≈ 2000, топ-10 ≈ 1700, топ-50 ≈ 1500
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

	// компонент формы
	// Упрощённая версия (без детальных данных о матчах)
	formPowerA := float64(teamA.FormWins) / 5.0
	formPowerB := float64(teamB.FormWins) / 5.0

	// Корректировка: если команда слабее по Valve, но форма хорошая - это ценнее
	if valveA < valveB {
		formPowerA *= 1.15 // аутсайдер с хорошей формой получает бонус
	} else {
		formPowerB *= 1.15
	}

	if formPowerA+formPowerB == 0 {
		formPowerA = 0.5
		formPowerB = 0.5
	}
	scoreFormA := formPowerA / (formPowerA + formPowerB)

	// компоненты карты
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

	// финальный расчет
	pA := wValve*scoreValveA + wForm*scoreFormA + wMap*scoreMapA
	pA = clamp01(pA)
	pB := 1 - pA

	return pA, pB
}

func calculateProbabilityValveAdvanced(teamA, teamB models.Team,
	recentMatchesA, recentMatchesB []RecentMatch) (float64, float64) {

	wValve := 0.55
	wForm := 0.30
	wMap := 0.15

	// Valve Points
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

	// Форма с учётом качества противников
	formScoreA := calculateFormScore(recentMatchesA, valveA)
	formScoreB := calculateFormScore(recentMatchesB, valveB)
	scoreFormA := formScoreA / (formScoreA + formScoreB)

	// Карты (упрощённая версия)
	mapPowerA := (teamA.MapPool.Mirage + teamA.MapPool.Nuke + teamA.MapPool.Ancient) / 3.0
	mapPowerB := (teamB.MapPool.Mirage + teamB.MapPool.Nuke + teamB.MapPool.Ancient) / 3.0

	if valveA > valveB && mapPowerA > 0.6 {
		mapPowerA *= 1.1
	}
	if valveB > valveA && mapPowerB > 0.6 {
		mapPowerB *= 1.1
	}

	scoreMapA := mapPowerA / (mapPowerA + mapPowerB)

	// Финал
	pA := wValve*scoreValveA + wForm*scoreFormA + wMap*scoreMapA
	pA = clamp01(pA)
	pB := 1 - pA

	return pA, pB
}

func CalcOdds(c *gin.Context) {
	var req models.OddsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Margin <= 0 {
		req.Margin = 0.06 // стандартная маржа примерно 6%
	}

	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objA, err := primitive.ObjectIDFromHex(req.TeamAID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TeamA id"})
		return
	}
	objB, err := primitive.ObjectIDFromHex(req.TeamBID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TeamB id"})
		return
	}

	var teamA models.Team
	var teamB models.Team

	if err := collection.FindOne(ctx, bson.M{"_id": objA}).Decode(&teamA); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teamA not found"})
		return
	}
	if err := collection.FindOne(ctx, bson.M{"_id": objB}).Decode(&teamB); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teamB not found"})
		return
	}

	// Используем улучшенную формулу с Valve Points
	pA, pB := calculateProbabilityValve(teamA, teamB)

	// Добавляем маржу букмекера
	pA_m := pA * (1 + req.Margin)
	pB_m := pB * (1 + req.Margin)

	oddA := round2(1 / pA_m)
	oddB := round2(1 / pB_m)

	// Получаем Valve Points для отладки
	valveA := float64(teamA.ValvePoints)
	valveB := float64(teamB.ValvePoints)

	// Fallback если нет Valve Points
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

	resp := models.OddsResponse{
		TeamA: teamA.TeamName,
		TeamB: teamB.TeamName,
		ProbA: round2(pA),
		ProbB: round2(pB),
		OddA:  oddA,
		OddB:  oddB,
	}

	c.JSON(http.StatusOK, gin.H{
		"odds": resp,
		"debug": gin.H{
			"valvePointsA": round2(valveA),
			"valvePointsB": round2(valveB),
			"pointsDiff":   round2(valveA - valveB),
			"source": map[string]bool{
				"teamA_has_valve": teamA.ValvePoints > 0,
				"teamB_has_valve": teamB.ValvePoints > 0,
			},
		},
	})
}
