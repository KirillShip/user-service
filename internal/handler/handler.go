package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/KirillShip/user-service/internal/service"
)

type contextKey string

const authUserKey contextKey = "authUser"

type AuthUser struct {
	ID   int64
	Role string
}

type Handler struct {
	userService *service.UserService
	JWTManager  *service.JWTManager
}

func NewHandler(userService *service.UserService, JWTManager *service.JWTManager) *Handler {
	return &Handler{
		userService: userService,
		JWTManager:  JWTManager,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	h.initPublicRoutes(mux)
	h.initProtectedRoutes(mux)

	return mux
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getBearerToken(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		claims, err := h.JWTManager.Verify(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		ctx := context.WithValue(
			r.Context(),
			authUserKey,
			AuthUser{
				ID:   claims.UserID,
				Role: claims.UserRole,
			},
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getBearerToken(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")

	if authorization == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "

	if !strings.HasPrefix(authorization, prefix) {
		return "", errors.New("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))

	if token == "" {
		return "", errors.New("missing bearer token")
	}

	return token, nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
