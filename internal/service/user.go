package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ObjoradDdd/AuthService/internal/db"
	"github.com/ObjoradDdd/AuthService/internal/kafka"
	"github.com/ObjoradDdd/AuthService/internal/model"
)

type UserService struct {
	storage *db.Storage
	kafka   *kafka.Producer
}

func NewUserService(storage *db.Storage, kafka *kafka.Producer) *UserService {
	return &UserService{storage: storage, kafka: kafka}
}

func (s *UserService) RegisterUser(ctx context.Context, user *model.User, password string) (*model.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return &model.User{}, err
	}

	return s.storage.CreateUser(ctx, user, hash)
}

func (s *UserService) UpdateUserHash(ctx context.Context, user *model.User, newPassword string, currentPassword string) error {
	currentHash, err := s.storage.GetHashById(ctx, user.Id)
	if err != nil {
		return fmt.Errorf("Error fetching current hash for user ID %d: %w", user.Id, err)
	}

	if !CheckPasswordHash(currentPassword, currentHash) {
		return fmt.Errorf("Current password is incorrect for user ID %d", user.Id)
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.storage.UpdateUserHash(ctx, user, hash)
}

func (s *UserService) Login(ctx context.Context, u *model.User, password string) (*model.User, error) {
	user, userHash, err := s.storage.GetUserAndHashByLogin(ctx, u.Login)
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, userHash) {
		return nil, fmt.Errorf("invalid password for user: %s", user.Login)
	}

	return user, nil

}

func (s *UserService) DeleteUserByID(ctx context.Context, id int) error {
	err := s.storage.DeleteUserById(ctx, id)
	if err != nil {
		return err
	}

	go func(id int) {
		for i := 0; i < 5; i++ {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := s.kafka.SendUserDeleted(bgCtx, id)
			if err == nil {
				slog.Info("User deleted event sent to Kafka", "userID", id)
				return
			}
			time.Sleep(2 * time.Second)
		}
		slog.Error("Failed to send user deleted event", "userID", id)
	}(id)

	return nil
}

func (s *UserService) UpdateUserLogin(ctx context.Context, user *model.User) error {
	return s.storage.UpdateUserLogin(ctx, user)
}
