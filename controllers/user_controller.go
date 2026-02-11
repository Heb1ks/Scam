package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"Scam/config"
	"Scam/models"
	"Scam/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// Регистрация пользователя
func (uc *UserController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.RespondError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	ctx := context.Background()
	usersColl := config.GetCollection("users")

	// Проверяем, существует ли пользователь
	var existing models.User
	err := usersColl.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": req.Email},
			{"username": req.Username},
		},
	}).Decode(&existing)

	if err == nil {
		utils.RespondError(w, http.StatusConflict, "User already exists")
		return
	}

	// Создаём пользователя
	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: utils.HashPassword(req.Password),
		Balance:      100.0, // Начальный бонус
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = usersColl.InsertOne(ctx, user)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"balance":  user.Balance,
		},
	})
}

// Вход пользователя
func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	usersColl := config.GetCollection("users")

	var user models.User
	err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"balance":  user.Balance,
		},
	})
}

// Получение профиля пользователя
func (uc *UserController) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.RespondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	ctx := context.Background()
	usersColl := config.GetCollection("users")

	var user models.User
	err = usersColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Рассчитываем win rate
	winRate := 0.0
	if user.TotalBets > 0 {
		winRate = float64(user.WonBets) / float64(user.TotalBets) * 100
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"id":            user.ID.Hex(),
		"username":      user.Username,
		"email":         user.Email,
		"balance":       user.Balance,
		"totalBets":     user.TotalBets,
		"wonBets":       user.WonBets,
		"lostBets":      user.LostBets,
		"winRate":       round2(winRate),
		"totalWagered":  user.TotalWagered,
		"totalWinnings": user.TotalWinnings,
		"profitLoss":    user.ProfitLoss,
	})
}

// Получение баланса
func (uc *UserController) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.RespondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	ctx := context.Background()
	usersColl := config.GetCollection("users")

	var user models.User
	err = usersColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"balance": user.Balance,
	})
}

// Добавление баланса (для тестирования)
func (uc *UserController) AddBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Amount <= 0 {
		utils.RespondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	objID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	ctx := context.Background()
	usersColl := config.GetCollection("users")

	result, err := usersColl.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$inc": bson.M{"balance": req.Amount},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)

	if err != nil || result.MatchedCount == 0 {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Получаем обновлённый баланс
	var user models.User
	usersColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Balance added successfully",
		"balance": user.Balance,
	})
}

func round2(x float64) float64 {
	return float64(int(x*100+0.5)) / 100
}
