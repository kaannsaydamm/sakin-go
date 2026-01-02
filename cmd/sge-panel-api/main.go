package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"sakin-go/cmd/sge-panel-api/config"
	"sakin-go/cmd/sge-panel-api/handlers"
	"sakin-go/cmd/sge-panel-api/services"
	"sakin-go/pkg/database"
)

func main() {
	cfg := config.LoadConfig()
	log.Println("[Panel API] Starting Backend...")

	// 1. DB Clients
	// ClickHouse
	chCfg := &database.ClickHouseConfig{
		Host: "localhost", Port: 9000, Database: "sge_logs", Username: "default",
	}
	ch, err := database.NewClickHouseClient(chCfg)
	if err != nil {
		log.Fatalf("[Panel API] ClickHouse Init Failed: %v", err)
	}

	// Postgres
	pgCfg := &database.PostgresConfig{
		Host: cfg.PostgresAddr, Port: 5432, Username: "postgres", Database: "sge_db", SSLMode: "disable",
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
