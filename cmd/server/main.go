package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/KirillShip/user-service/internal/config"
	"github.com/KirillShip/user-service/internal/database"
	"github.com/KirillShip/user-service/internal/handler"
	"github.com/KirillShip/user-service/internal/repository"
	"github.com/KirillShip/user-service/internal/server"
	"github.com/KirillShip/user-service/internal/service"
)

func main() {
	config := config.LoadConfig()
	if config == nil {
		log.Fatalf("failed to load config")
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pgpool, err := database.NewPostgresPool(dbCtx, config.DatabaseDSN)
	if err != nil {
		log.Fatalf("error : %s\n", err)
	}
	defer pgpool.Close()
	userRepo := repository.NewUserRepository(pgpool)
	userService := service.NewUserService(userRepo)
	authService := service.NewJWTManager(config.JWTSecretkey, 15*time.Minute)
	h := handler.NewHandler(userService, authService)
	router := h.InitRoutes()
	srv := server.NewServer(config, router)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error : %s\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("server stopped.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %s", err)
	}

	log.Println("server exited")
}
