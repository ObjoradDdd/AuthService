package handler

import (
	"crypto/rsa"
	"net/http"

	"github.com/ObjoradDdd/AuthService/internal/model"
	"github.com/ObjoradDdd/AuthService/internal/service"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService *service.UserService
	privateKey  *rsa.PrivateKey
	validator   *validator.Validate
}

func NewUserHandler(userService *service.UserService, privateKey *rsa.PrivateKey) *UserHandler {
	return &UserHandler{
		userService: userService,
		privateKey:  privateKey,
		validator:   validator.New(),
	}
}

type loginRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeAndValidateRequest(w, r, &req, h.validator); err != nil {
		return
	}

	user, err := h.userService.Login(r.Context(), &model.User{Login: req.Login}, req.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid login or password")
		return
	}

	token, err := generateToken(user.Id, h.privateKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error generating token")
		return
	}

	resp := loginResponse{Token: token}
	respondWithJSON(w, http.StatusOK, resp)
}

type registerRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6"`
}

type registerResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeAndValidateRequest(w, r, &req, h.validator); err != nil {
		return
	}

	user, err := h.userService.RegisterUser(r.Context(), &model.User{Login: req.Login}, req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := generateToken(user.Id, h.privateKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error generating token")
		return
	}

	resp := registerResponse{Token: token}
	respondWithJSON(w, http.StatusOK, resp)
}

type deleteResponse struct {
	Id int `json:"id"`
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getUserID(w, r)
	if err != nil {
		return
	}

	err = h.userService.DeleteUserByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := deleteResponse{Id: id}
	respondWithJSON(w, http.StatusOK, resp)
}

type updateHashRequest struct {
	NewPassword     string `json:"password" validate:"required,min=6"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
}

type updateHashResponse struct {
	Id int `json:"id"`
}

func (h *UserHandler) UpdateHash(w http.ResponseWriter, r *http.Request) {
	var req updateHashRequest
	if err := decodeAndValidateRequest(w, r, &req, h.validator); err != nil {
		return
	}

	id, err := getUserID(w, r)
	if err != nil {
		return
	}

	err = h.userService.UpdateUserHash(r.Context(), &model.User{Id: id}, req.NewPassword, req.CurrentPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updateHashResponse{Id: id})
}
