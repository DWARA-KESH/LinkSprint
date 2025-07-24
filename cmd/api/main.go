package main

import (
	"context"
	"fmt"
	"log"
	"os" // Import os for Getenv

	// For cache expiration
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"

	"github.com/DWARA-KESH/LinkSprint/internal/cache"      // Corrected import path
	"github.com/DWARA-KESH/LinkSprint/internal/handler"    // Corrected import path
	"github.com/DWARA-KESH/LinkSprint/internal/repository" // Corrected import path
)

func main() {
	app := fiber.New()

	// CORS middleware (essential for frontend)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// 1. Initialize CockroachDB
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgresql://root@localhost:26257/LinkSprint?sslmode=disable"
		log.Println("DATABASE_URL environment variable not set, falling back to local CockroachDB.")
	}
	db, err := repository.InitDB(dbConnStr) // Corrected: Expects 2 return values, passes argument
	if err != nil {
		log.Fatalf("failed to connect to CockroachDB: %v", err)
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Error closing DB connection: %v", err)
		} else {
			log.Println("CockroachDB connection closed.")
		}
	}()
	urlRepo := repository.NewURLRepository(db)
	log.Println("CockroachDB connected successfully.")

	// 2. Initialize Redis Client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		log.Println("REDIS_ADDR environment variable not set, falling back to local Redis.")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Use address from env var or fallback
		DB:   0,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Redis connection closed.")
		}
	}()
	urlCache := cache.NewURLCache(redisClient)
	log.Println("Redis connected successfully.")

	// Determine the base URL for shortened links (use Render's URL in production)
	serviceBaseURL := os.Getenv("SERVICE_BASE_URL")
	if serviceBaseURL == "" {
		serviceBaseURL = "http://localhost:3000"
	}

	// 3. Initialize handler with repository, cache, and the base URL
	handler.InitHandler(urlRepo, urlCache, serviceBaseURL)

	// 4. Define Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("LinkSprint API is up!")
	})

	app.Post("/shorten", handler.ShortenURL)
	app.Get("/:code", handler.Redirect)

	// 5. Start Fiber App
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port if not set by environment
	}
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port))) // Listen on port from env or fallback
}
