package service

import (
	"fmt"

	"github.com/ObjoradDdd/AuthService/internal/db"
	"github.com/ObjoradDdd/AuthService/internal/model"
)

type UserService struct {
	storage *db.Storage
}

func NewUserService(storage *db.Storage) *UserService {
	return &UserService{storage: storage}
}

func (s *UserService) RegisterUser(user *model.User, password string) (*model.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return &model.User{}, err
	}

	return s.storage.CreateUser(user, hash)
}

func (s *UserService) UpdateUserHash(user *model.User, newPassword string, currentPassword string) error {
	currentHash, err := s.storage.GetHashById(user.Id)
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

	return s.storage.UpdateUserHash(user, hash)
}

func (s *UserService) Login(u *model.User, password string) (*model.User, error) {
	user, userHash, err := s.storage.GetUserAndHashByLogin(u.Login)
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(password, userHash) {
		return nil, fmt.Errorf("invalid password for user: %s", user.Login)
	}

	return user, nil

}

func (s *UserService) DeleteUserByID(id int) error {
	return s.storage.DeleteUserById(id)
}

func (s *UserService) UpdateUserLogin(user *model.User) error {
	return s.storage.UpdateUserLogin(user)
}
