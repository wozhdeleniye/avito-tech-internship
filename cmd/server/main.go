package main

import (
	"log"
	"net/http"

	"github.com/wozhdeleniye/avito-tech-internship/internal/app/router"
	"github.com/wozhdeleniye/avito-tech-internship/internal/config"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/migrations"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
	redisrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/redis_repository"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
	"github.com/wozhdeleniye/avito-tech-internship/pkg/database"
	"github.com/wozhdeleniye/avito-tech-internship/pkg/redis"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgresConnection(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Println("Ошибка получения sql.DB:", err)
			return
		}
		sqlDB.Close()
	}()

	migrator := migrations.NewGormMigrator(db)
	migrator.Migrate()

	redisClient, err := redis.NewRedisConnection(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	defer redisClient.Close()

	userRepo := postgresrepository.NewUserRepository(db)
	tokenRepo := redisrepository.NewTokenRepository(redisClient)

	authService := services.NewAuthService(userRepo, tokenRepo, cfg.JWT)

	r := router.NewApp(authService)

	http.ListenAndServe(cfg.Server.Port, r)
}
