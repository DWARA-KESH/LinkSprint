package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"

	"github.com/DWARA-KESH/LinkSprint/internal/cache"
	"github.com/DWARA-KESH/LinkSprint/internal/handler"
	"github.com/DWARA-KESH/LinkSprint/internal/repository"
)

func main() {
	app := fiber.New()

	// CORS middleware (essential for frontend)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // For local testing. Change to your Render frontend URL in production.
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// 1. Initialize CockroachDB
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgresql://root@localhost:26257/linksprint?sslmode=disable" // Fallback for local Docker Compose
		log.Println("DATABASE_URL environment variable not set, falling back to local CockroachDB.")
	}
	db, err := repository.InitDB(dbConnStr)
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
		redisAddr = "localhost:6379" // Fallback for local Docker Compose
		log.Println("REDIS_ADDR environment variable not set, falling back to local Redis.")
	}

	// --- NEW CODE FOR PARSING REDIS URL ---
	parsedRedisURL, err := url.Parse(redisAddr)
	if err != nil {
		log.Fatalf("failed to parse Redis URL '%s': %v", redisAddr, err)
	}

	// Use Host (which is hostname:port) from the parsed URL
	redisHostPort := parsedRedisURL.Host
	if redisHostPort == "" {
		// Fallback if parsing didn't yield a host:port, might happen for "localhost:6379" directly
		redisHostPort = redisAddr
	}
	// --- END NEW CODE ---

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHostPort, // <--- Use the parsed host:port here
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
		serviceBaseURL = "http://localhost:3000" // Default for local
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
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
