package handler

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

type errorResponse struct {
	Error string `json:"error"`
}

func generateToken(Id int, privateKey *rsa.PrivateKey) (string, error) {
	if privateKey == nil {
		return "", errors.New("private key is nil")
	}

	claims := jwt.MapClaims{
		"id":  Id,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func getUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "internal server error: failed to get user id from context")
		return 0, errors.New("user id not found in context")
	}
	return userID, nil
}

func decodeRequest(w http.ResponseWriter, r *http.Request, req any) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "bad request: invalid json")
		return err
	}
	return nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(buf.Bytes())
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func decodeAndValidateRequest(w http.ResponseWriter, r *http.Request, req any, validator *validator.Validate) error {
	if err := decodeRequest(w, r, req); err != nil {
		respondWithError(w, http.StatusBadRequest, "error decoding request body")
		return err
	}

	if err := validator.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation error")
		return err
	}

	return nil
}
