package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cs2-betting-platform/config"
	"cs2-betting-platform/models"
	"cs2-betting-platform/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.RespondError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	userColl := config.GetCollection("users")
	ctx := context.Background()

	var existingUser models.User
	err := userColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		utils.RespondError(w, http.StatusConflict, "User with this email already exists")
		return
	}

	hashedPassword := utils.HashPassword(req.Password)

	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Balance:   1000.0, // стартовый баланс
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Balance > 0 {
		user.Balance = req.Balance
	}

	_, err = userColl.InsertOne(ctx, user)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	log.Printf("User registered: %s (%s)", user.Username, user.Email)

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userColl := config.GetCollection("users")
	ctx := context.Background()

	var user models.User
	err := userColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	log.Printf("User logged in: %s", user.Email)

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})
}

func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userColl := config.GetCollection("users")
	ctx := context.Background()

	var user models.User
	err = userColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (uc *UserController) GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userColl := config.GetCollection("users")
	ctx := context.Background()

	var user models.User
	err = userColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"balance": user.Balance,
	})
}

func (uc *UserController) AddBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userColl := config.GetCollection("users")
	ctx := context.Background()

	result, err := userColl.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$inc": bson.M{"balance": req.Amount},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil || result.MatchedCount == 0 {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Balance updated successfully",
	})
}
