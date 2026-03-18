package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ObjoradDdd/AuthService/internal/db"
	"github.com/ObjoradDdd/AuthService/internal/handler"
	"github.com/ObjoradDdd/AuthService/internal/kafka"
	"github.com/ObjoradDdd/AuthService/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

func main() {

	// Создание основного контекста, который будет отменён при получении сигнала завершения
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Используем WaitGroup для ожидания завершения всех горутин при shutdown
	var wg sync.WaitGroup

	// Создание канала для ошибок сервера, чтобы можно было корректно обработать их при завершении
	serverErrors := make(chan error, 1)

	// Инициализация базы данных
	db, err := initDatabase(ctx)
	if err != nil {
		slog.Error("Connection to database failed", "error", err)
		serverErrors <- err
	}
	defer db.Close()

	// Инициализация Kafka продюсера
	kafkaProducer := kafka.NewProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	defer kafkaProducer.Close()

	// Инициализация сервисов
	userService := service.NewUserService(db, kafkaProducer)

	// Загрузка RSA приватного ключа
	bytes, err := os.ReadFile(os.Getenv("RSA_PRIVATE_KEY_PATH"))
	if err != nil {
		slog.Error("Error reading RSA private key", "path", os.Getenv("RSA_PRIVATE_KEY_PATH"), "error", err)
		serverErrors <- err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		slog.Error("Error parsing RSA private key", "error", err)
		serverErrors <- err
	}

	// Настройка HTTP сервера
	router := setupRouter(userService, privateKey)

	var srv *http.Server
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		slog.Warn("invalid SERVER_PORT", "error", err)
		serverErrors <- err
	} else {
		srv = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("server is starting", "port", port, "url", fmt.Sprintf("http://localhost:%d", port))
			serverErrors <- srv.ListenAndServe()
		}()

	}

	// Ожидание сигнала завершения или ошибки сервера
	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			slog.Error("server failed prematurely", "error", err)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	// Ждем завершения всех горутин
	wg.Wait()
	slog.Info("shutting down server")
}

func initDatabase(ctx context.Context) (*db.Storage, error) {
	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	return db.New(ctx, db.Config{
		ConnString:      conn,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
}

func setupRouter(userService *service.UserService, privateKey *rsa.PrivateKey) *http.ServeMux {
	mux := http.NewServeMux()

	userHandler := handler.NewUserHandler(userService, privateKey)

	mux.HandleFunc("POST /user/login", userHandler.Login)
	mux.HandleFunc("POST /user/register", userHandler.Register)
	mux.HandleFunc("DELETE /user/delete", handler.Middleware(userHandler.Delete))
	mux.HandleFunc("PUT /user/reset-password", handler.Middleware(userHandler.UpdateHash))

	return mux
}
