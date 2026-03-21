package handler

import (
	"context"
	"crypto/rsa"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

func getUserIDFromMetadata(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata not found in context")
	}

	values := md.Get("x-user-id")
	if len(values) == 0 || values[0] == "" {
		return 0, status.Error(codes.Unauthenticated, "user-id not found")
	}

	id, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "invalid user-id format")
	}

	return id, nil
}
