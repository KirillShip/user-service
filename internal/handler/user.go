package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/KirillShip/user-service/internal/model"
	"github.com/KirillShip/user-service/internal/service"
)

func (h *Handler) initPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/register", h.registerUser)
	mux.HandleFunc("POST /api/v1/login", h.loginUser)
}

func (h *Handler) initProtectedRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/user/{id}", h.authMiddleware(http.HandlerFunc(h.getUserByID)))
	mux.Handle("PUT /api/v1/user/{id}", h.authMiddleware(http.HandlerFunc(h.updateUser)))
	mux.Handle("DELETE /api/v1/user/{id}", h.authMiddleware(http.HandlerFunc(h.deleteUserByID)))
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Name, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrNameInvalid) ||
			errors.Is(err, service.ErrLoginInvalid) ||
			errors.Is(err, service.ErrPasswordInvalid) ||
			errors.Is(err, service.ErrLoginAlreadyUsed) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		} else {
			log.Printf("CreateUser error: %v \n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
	}
	JWTToken, err := h.JWTManager.Create(user.ID, user.Role)
	if err != nil {
		log.Printf("CreateJWTToken error: %v \n", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}
	writeJSON(w, http.StatusCreated, JWTToken)
}

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}
	user, err := h.userService.CheckUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid login or password",
			})
			return
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
	}
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "incorrect login or password",
		})
		return
	} else {
		JWTToken, err := h.JWTManager.Create(user.ID, user.Role)
		if err != nil {
			log.Printf("CreateJWTToken error: %v \n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
		writeJSON(w, http.StatusAccepted, JWTToken)
	}
}

func (h *Handler) getUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid user id",
		})
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrIDInvalid) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		} else if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
			return
		} else {
			log.Printf("GetUser error: %v \n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid user id",
		})
		return
	}

	authUser, ok := r.Context().Value(authUserKey).(AuthUser)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	if authUser.Role != model.RoleAdmin && authUser.ID != id {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "forbidden",
		})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}
	user, err := h.userService.UpdateUser(r.Context(), id, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrNameInvalid) ||
			errors.Is(err, service.ErrIDInvalid) ||
			errors.Is(err, service.ErrPasswordInvalid) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		} else {
			log.Printf("UpdateUser error: %v \n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) deleteUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid user id",
		})
		return
	}

	authUser, ok := r.Context().Value(authUserKey).(AuthUser)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	if authUser.Role != model.RoleAdmin && authUser.ID != id {
		writeJSON(w, http.StatusForbidden, map[string]string{
			"error": "forbidden",
		})
		return
	}

	err = h.userService.DeleteUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrIDInvalid) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		} else if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
			return
		} else {
			log.Printf("DeleteUser error: %v \n", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal server error",
			})
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
