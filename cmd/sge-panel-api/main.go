package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"sakin-go/cmd/sge-panel-api/config"
	"sakin-go/cmd/sge-panel-api/handlers"
	"sakin-go/cmd/sge-panel-api/services"
	"sakin-go/pkg/database"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("[Warning] No .env file found")
	}

	cfg := config.LoadConfig()
	log.Println("[Panel API] Starting Backend...")
	log.Printf("[Debug] ClickHouse Config - Host: %s, Port: %d, User: %s, DB: %s, PassLen: %d",
		cfg.ClickHouseHost, cfg.ClickHousePort, cfg.ClickHouseUser, cfg.ClickHouseDB, len(cfg.ClickHousePass))

	// 1. DB Clients
	// ClickHouse
	chCfg := &database.ClickHouseConfig{
		Host:     cfg.ClickHouseHost,
		Port:     cfg.ClickHousePort,
		Database: cfg.ClickHouseDB,
		Username: cfg.ClickHouseUser,
		Password: cfg.ClickHousePass,
		Debug:    true,
	}
	ch, err := database.NewClickHouseClient(chCfg)
	if err != nil {
		log.Fatalf("[Panel API] ClickHouse Init Failed: %v", err)
	}

	// Init Schema
	if err := ch.InitializeSchema(context.Background()); err != nil {
		log.Printf("[Warning] ClickHouse Schema Init Failed: %v", err)
	}

	// Postgres
	pgCfg := &database.PostgresConfig{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		Username: cfg.PostgresUser,
		Password: cfg.PostgresPass,
		Database: cfg.PostgresDB,
		SSLMode:  "disable",
	}
	pg, err := database.NewPostgresClient(pgCfg)
	if err != nil {
		log.Fatalf("[Panel API] Postgres Init Failed: %v", err)
	}

	// 2. Services & Handlers
	dashboardSvc := services.NewDashboardService(ch, pg)
	dashboardHandler := handlers.NewDashboardHandler(dashboardSvc)

	// 3. App
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000", // Allow Next.js frontend
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	api := app.Group("/api/v1")

	api.Get("/dashboard/stats", dashboardHandler.GetStats)

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// 4. Start
	log.Printf("[Panel API] Listening on %s", cfg.Port)
	log.Fatal(app.Listen(cfg.Port))
}
