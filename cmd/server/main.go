package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/wozhdeleniye/avito-tech-internship/internal/app/router"
	"github.com/wozhdeleniye/avito-tech-internship/internal/config"
	"github.com/wozhdeleniye/avito-tech-internship/internal/pkg/db/database"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/migrations"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
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

	fmt.Println(cfg.Database.Password)

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

	userRepo := postgresrepository.NewUserRepository(db)
	teamRepo := postgresrepository.NewTeamRepository(db)
	prRepo := postgresrepository.NewPReqRepository(db)

	teamService := services.NewTeamService(prRepo, teamRepo, userRepo)
	prService := services.NewPReqService(prRepo, teamRepo, userRepo)

	r := router.NewApp(prService, teamService)

	addr := cfg.Server.Port
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
