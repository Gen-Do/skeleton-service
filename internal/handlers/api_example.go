package handlers

// Пример структуры для API обработчиков
// Этот файл показывает, как организовать обработчики для сгенерированного API

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	// "github.com/gendo/service-skeleton/internal/generated/api"
)

// APIHandlers содержит все зависимости для API обработчиков
type APIHandlers struct {
	Logger *logrus.Logger
	// DB     database.Interface
	// Services *services.Services
}

// Пример реализации обработчиков, соответствующих сгенерированному API:

// GetUsers обрабатывает GET /users
// func (h *APIHandlers) GetUsers(w http.ResponseWriter, r *http.Request) {
//     // Парсинг query параметров
//     params := api.GetUsersParams{}
//     if err := runtime.BindQueryParams(r.URL.Query(), &params); err != nil {
//         h.writeErrorResponse(w, "Invalid query parameters", http.StatusBadRequest)
//         return
//     }
//
//     // Логирование запроса
//     h.Logger.WithFields(logrus.Fields{
//         "page":  params.Page,
//         "limit": params.Limit,
//     }).Info("Getting users list")
//
//     // Бизнес-логика
//     users, pagination, err := h.Services.UserService.GetUsers(r.Context(), params.Page, params.Limit)
//     if err != nil {
//         h.Logger.WithError(err).Error("Failed to get users")
//         h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
//         return
//     }
//
//     // Формирование ответа
//     response := api.UsersResponse{
//         Users:      users,
//         Pagination: pagination,
//     }
//
//     h.writeJSONResponse(w, response, http.StatusOK)
// }

// CreateUser обрабатывает POST /users
// func (h *APIHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
//     var req api.CreateUserRequest
//     if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//         h.writeErrorResponse(w, "Invalid JSON body", http.StatusBadRequest)
//         return
//     }
//
//     h.Logger.WithFields(logrus.Fields{
//         "email": req.Email,
//         "name":  req.Name,
//     }).Info("Creating new user")
//
//     user, err := h.Services.UserService.CreateUser(r.Context(), req)
//     if err != nil {
//         h.Logger.WithError(err).Error("Failed to create user")
//         if errors.Is(err, services.ErrUserAlreadyExists) {
//             h.writeErrorResponse(w, "User already exists", http.StatusConflict)
//             return
//         }
//         h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
//         return
//     }
//
//     h.writeJSONResponse(w, user, http.StatusCreated)
// }

// GetUserById обрабатывает GET /users/{userId}
// func (h *APIHandlers) GetUserById(w http.ResponseWriter, r *http.Request) {
//     userID := chi.URLParam(r, "userId")
//     if userID == "" {
//         h.writeErrorResponse(w, "User ID is required", http.StatusBadRequest)
//         return
//     }
//
//     h.Logger.WithField("user_id", userID).Info("Getting user by ID")
//
//     user, err := h.Services.UserService.GetUserByID(r.Context(), userID)
//     if err != nil {
//         h.Logger.WithError(err).Error("Failed to get user")
//         if errors.Is(err, services.ErrUserNotFound) {
//             h.writeErrorResponse(w, "User not found", http.StatusNotFound)
//             return
//         }
//         h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
//         return
//     }
//
//     h.writeJSONResponse(w, user, http.StatusOK)
// }

// UpdateUser обрабатывает PUT /users/{userId}
// func (h *APIHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
//     userID := chi.URLParam(r, "userId")
//     if userID == "" {
//         h.writeErrorResponse(w, "User ID is required", http.StatusBadRequest)
//         return
//     }
//
//     var req api.UpdateUserRequest
//     if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//         h.writeErrorResponse(w, "Invalid JSON body", http.StatusBadRequest)
//         return
//     }
//
//     h.Logger.WithFields(logrus.Fields{
//         "user_id": userID,
//         "name":    req.Name,
//     }).Info("Updating user")
//
//     user, err := h.Services.UserService.UpdateUser(r.Context(), userID, req)
//     if err != nil {
//         h.Logger.WithError(err).Error("Failed to update user")
//         if errors.Is(err, services.ErrUserNotFound) {
//             h.writeErrorResponse(w, "User not found", http.StatusNotFound)
//             return
//         }
//         h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
//         return
//     }
//
//     h.writeJSONResponse(w, user, http.StatusOK)
// }

// DeleteUser обрабатывает DELETE /users/{userId}
// func (h *APIHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
//     userID := chi.URLParam(r, "userId")
//     if userID == "" {
//         h.writeErrorResponse(w, "User ID is required", http.StatusBadRequest)
//         return
//     }
//
//     h.Logger.WithField("user_id", userID).Info("Deleting user")
//
//     err := h.Services.UserService.DeleteUser(r.Context(), userID)
//     if err != nil {
//         h.Logger.WithError(err).Error("Failed to delete user")
//         if errors.Is(err, services.ErrUserNotFound) {
//             h.writeErrorResponse(w, "User not found", http.StatusNotFound)
//             return
//         }
//         h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
//         return
//     }
//
//     w.WriteHeader(http.StatusNoContent)
// }

// Вспомогательные методы для работы с HTTP ответами

func (h *APIHandlers) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.Logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *APIHandlers) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.Logger.WithError(err).Error("Failed to encode error response")
	}
}

// Пример middleware для дополнительной валидации
func (h *APIHandlers) ValidateUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")

		// Валидация UUID формата
		// if !isValidUUID(userID) {
		//     h.writeErrorResponse(w, "Invalid user ID format", http.StatusBadRequest)
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}
