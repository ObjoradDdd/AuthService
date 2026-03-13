package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ObjoradDdd/AuthService/internal/db"
	"github.com/ObjoradDdd/AuthService/internal/handler"
	"github.com/ObjoradDdd/AuthService/internal/kafka"
	"github.com/ObjoradDdd/AuthService/internal/service"
)

func main() {

	mux := http.NewServeMux()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	connectionString := db.GetConnectionString(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_TYPE"),
	)

	db, err := db.New(ctx, db.Config{
		ConnString:      connectionString,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	kafkaProducer := kafka.NewProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))

	defer kafkaProducer.Close()

	userServise := service.NewUserService(db, kafkaProducer)

	userHandler := handler.NewUserHandler(userServise)

	mux.HandleFunc("POST /user/login", userHandler.Login)
	mux.HandleFunc("POST /user/register", userHandler.Register)
	mux.HandleFunc("DELETE /user/delete", handler.Middleware(userHandler.Delete))
	mux.HandleFunc("PUT /user/reset-password", handler.Middleware(userHandler.UpdateHash))

	err = http.ListenAndServe(":5134", mux)
	if err != nil {
		log.Fatal(err)
	}
}
