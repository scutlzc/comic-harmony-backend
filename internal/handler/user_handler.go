package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/response"
	"github.com/muyue/comic-harmony-backend/internal/service"
)

type UserHandler struct {
	svc       service.UserService
	jwtSecret string
}

func NewUserHandler(svc service.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{svc: svc, jwtSecret: jwtSecret}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, resp)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, resp)
}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(int64)

	user, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}

	response.Success(w, user)
}

// Middleware

type contextKey string

const UserIDKey contextKey = "user_id"

func (h *UserHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			response.JSON(w, http.StatusUnauthorized, response.APIResponse{
				Code: 401, Message: "missing authorization header",
			})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if tokenStr == auth {
			response.JSON(w, http.StatusUnauthorized, response.APIResponse{
				Code: 401, Message: "invalid authorization format",
			})
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.JSON(w, http.StatusUnauthorized, response.APIResponse{
				Code: 401, Message: "invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.JSON(w, http.StatusUnauthorized, response.APIResponse{
				Code: 401, Message: "invalid token claims",
			})
			return
		}

		userID := int64(claims["user_id"].(float64))
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
