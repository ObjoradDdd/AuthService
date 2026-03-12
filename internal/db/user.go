package db

import (
	"fmt"

	"log/slog"

	"github.com/ObjoradDdd/AuthService/internal/model"
)

func (s *Storage) GetUserAndHashByLogin(login string) (*model.User, string, error) {
	var hash string
	var user model.User

	err := s.db.QueryRow("SELECT * FROM users WHERE login = $1", login).Scan(&user.Id, &user.Login, &hash)
	if err != nil {
		slog.Error("Error fetching user hash by login", "login", login, "error", err)
		return nil, "", fmt.Errorf("Error fetching user hash by login: %s", login)
	}

	return &user, hash, nil
}

func (s *Storage) CreateUser(user *model.User, hash string) (*model.User, error) {
	err := s.db.QueryRow("INSERT INTO users (login, hash) VALUES ($1, $2) RETURNING Id", user.Login, hash).Scan(&user.Id)
	if err != nil {
		slog.Error("Error creating user", "user", user, "error", err)
		return nil, fmt.Errorf("Error creating user: %w", err)
	}

	return user, nil
}

func (s *Storage) DeleteUserById(Id int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE Id = $1", Id)
	if err != nil {
		slog.Error("Error deleting user by Id", "Id", Id, "error", err)
		return fmt.Errorf("Error deleting user by Id: %d", Id)
	}

	return nil
}

func (s *Storage) UpdateUserLogin(user *model.User) error {
	_, err := s.db.Exec("UPDATE users SET login = $1 WHERE Id = $2", user.Login, user.Id)

	if err != nil {
		slog.Error("Error updating user login", "Id", user.Id, "newLogin", user.Login, "error", err)
		return fmt.Errorf("Error updating user login: %d", user.Id)
	}

	return nil
}

func (s *Storage) UpdateUserHash(user *model.User, hash string) error {
	_, err := s.db.Exec("UPDATE users SET hash = $1 WHERE Id = $2", hash, user.Id)

	if err != nil {
		slog.Error("Error updating user hash", "Id", user.Id, "error", err)
		return fmt.Errorf("Error updating user hash: %d", user.Id)
	}

	return nil
}

func (s *Storage) GetHashById(id int) (string, error) {
	var hash string

	err := s.db.QueryRow("SELECT hash FROM users WHERE Id = $1", id).Scan(&hash)
	if err != nil {
		slog.Error("Error fetching user hash by Id", "Id", id, "error", err)
		return "", fmt.Errorf("Error fetching user hash by Id: %d", id)
	}

	return hash, nil
}
