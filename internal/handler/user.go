package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ObjoradDdd/AuthService/internal/model"
	"github.com/ObjoradDdd/AuthService/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Login(&model.User{Login: req.Login}, req.Password)
	if err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.Id)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	resp := loginResponse{Token: token}
	json.NewEncoder(w).Encode(resp)

}

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type registerResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println(req)

	user, err := h.userService.RegisterUser(&model.User{Login: req.Login}, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := generateToken(user.Id)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	resp := registerResponse{Token: token}
	json.NewEncoder(w).Encode(resp)
}

type deleteResponse struct {
	Id int `json:"id"`
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.Context().Value("userId").(int)

	err := h.userService.DeleteUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := deleteResponse{Id: id}
	json.NewEncoder(w).Encode(resp)
}

type updateHashRequest struct {
	NewPassword     string `json:"password"`
	CurrentPassword string `json:"currentPassword"`
}

type updateHashResponse struct {
	Id int `json:"id"`
}

func (h *UserHandler) UpdateHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req updateHashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id := r.Context().Value("userId").(int)

	err = h.userService.UpdateUserHash(&model.User{Id: id}, req.NewPassword, req.CurrentPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updateHashResponse{Id: id})
}
