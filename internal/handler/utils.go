package handler

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func init() {
	godotenv.Load()

	privPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	pubPath := os.Getenv("JWT_PUBLIC_KEY_PATH")

	if privPath != "" {
		bytes, err := os.ReadFile(privPath)
		if err != nil {
			panic("ошибка чтения приватного ключа: " + err.Error())
		}
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(bytes)
		if err != nil {
			panic("ошибка парсинга приватного ключа: " + err.Error())
		}
	}

	if pubPath != "" {
		bytes, err := os.ReadFile(pubPath)
		if err != nil {
			panic("ошибка чтения публичного ключа: " + err.Error())
		}
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM(bytes)
		if err != nil {
			panic("ошибка парсинга публичного ключа: " + err.Error())
		}
	}
}

func generateToken(Id int) (string, error) {
	if privateKey == nil {
		return "", errors.New("приватный ключ не загружен")
	}

	claims := jwt.MapClaims{
		"id":  Id,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}
