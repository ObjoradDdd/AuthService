package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ObjoradDdd/AuthService/internal/db"
	"github.com/ObjoradDdd/AuthService/internal/handler"
	"github.com/ObjoradDdd/AuthService/internal/kafka"
	"github.com/ObjoradDdd/AuthService/internal/service"
	pb "github.com/ObjoradDdd/AuthService/proto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using system environment variables")
	}

	// Канал для получения ошибок от сервера
	serverErrors := make(chan error, 1)

	// Создание основного контекста, который будет отменён при получении сигнала завершения
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Инициализация базы данных
	db, err := initDatabase(ctx)
	if err != nil {
		slog.Error("Connection to database failed", "error", err)
		return
	}
	defer db.Close()

	// Инициализация Kafka продюсера
	kafkaProducer := kafka.NewProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	defer kafkaProducer.Close()

	// Загрузка RSA приватного ключа
	bytes, err := os.ReadFile(os.Getenv("JWT_PRIVATE_KEY_PATH"))
	if err != nil {
		slog.Error("Error reading RSA private key", "path", os.Getenv("JWT_PRIVATE_KEY_PATH"), "error", err)
		return
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		slog.Error("Error parsing RSA private key", "error", err)
		return
	}

	// Инициализация сервисов
	userService := service.NewUserService(db, kafkaProducer, privateKey)

	// Настройка TLS для gRPC сервера
	tlsConfig, err := loadTLSCredentials()
	if err != nil {
		slog.Error("failed to load TLS credentials", "error", err)
		return
	}

	// Создание группы ожидания
	var wg sync.WaitGroup

	// Запуск gRPC сервера в отдельной горутине
	grpcServer := handler.NewAuthGRPCServer(userService, privateKey)
	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterAuthServiceServer(s, grpcServer)

	lis, err := net.Listen("tcp", os.Getenv("GRPC_URL"))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("gRPC server is listening on", "url", os.Getenv("GRPC_URL"))
		if err := s.Serve(lis); err != nil {
			serverErrors <- fmt.Errorf("failed to serve: %w", err)
			return
		}
	}()

	// Ожидаем сигнала завершения или ошибки сервера
	select {
	case err := <-serverErrors:
		slog.Error("server failed to start", "error", err)
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	// Останавливаем gRPC сервер
	if s != nil {
		s.GracefulStop()
		slog.Info("gRPC server stopped")
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

func loadTLSCredentials() (*tls.Config, error) {
	// Загружаем сертификат сервера и ключ
	serverCert, err := tls.LoadX509KeyPair(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %w", err)
	}

	// Загружаем сертификат CA
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(os.Getenv("TLS_CA_PATH"))
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}
	certPool.AppendCertsFromPEM(ca)

	// Настраиваем конфиг TLS для сервера
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // Обязательная проверка
		ClientCAs:    certPool,
	}

	return tlsConfig, nil
}
