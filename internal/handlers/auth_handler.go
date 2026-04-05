package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/elllban/test-backend-elllban/internal/domain"
	"github.com/elllban/test-backend-elllban/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	jwtSecret []byte
	userRepo  repository.UserRepository
}

func NewAuthHandler(jwtSecret []byte, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		jwtSecret: jwtSecret,
		userRepo:  userRepo,
	}
}

// DummyLogin @Summary Получить тестовый JWT
// @Description Возвращает JWT токен для указанной роли (admin/user)
// @Tags Auth
// @Param request body domain.DummyLoginRequest true "Роль пользователя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /dummyLogin [post]
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req domain.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		sendError(w, "INVALID_REQUEST", "role must be admin or user", http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	if req.Role == "admin" {
		userID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	} else {
		userID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	}

	_, err := h.userRepo.CreateOrGetByID(userID, req.Role)
	if err != nil {
		sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    req.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		sendError(w, "INTERNAL_ERROR", "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// InfoHandler @HealthCheck Информация о сервисе
// @Description Всегда возвращает 200 OK
// @Tags Health
// @Success 200 {object} map[string]string
// @Router /_info [get]
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func sendError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errResp := domain.ErrorResponse{}
	errResp.Error.Code = code
	errResp.Error.Message = message
	json.NewEncoder(w).Encode(errResp)
}
