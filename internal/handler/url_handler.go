package handler

import (
	"context"
	"fmt"
	"log"
	"net/url" // For more robust URL validation
	"time"

	"github.com/DWARA-KESH/LinkSprint/internal/cache"      // Corrected import path
	"github.com/DWARA-KESH/LinkSprint/internal/model"      // Corrected import path
	"github.com/DWARA-KESH/LinkSprint/internal/repository" // Corrected import path
	"github.com/DWARA-KESH/LinkSprint/pkg/utils"           // Corrected import path
	"github.com/gofiber/fiber/v2"
)

var serviceBaseURL string // Variable to store the base URL for generating short links

var (
	repo     *repository.URLRepository
	urlCache *cache.URLCache
)

// InitHandler now accepts the base URL
func InitHandler(r *repository.URLRepository, c *cache.URLCache, baseURL string) {
	repo = r
	urlCache = c
	serviceBaseURL = baseURL // Store the base URL
	log.Println("Handlers initialized.")
}

func ShortenURL(c *fiber.Ctx) error {
	var req model.URLRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body in ShortenURL: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.URL == "" {
		log.Println("ShortenURL request missing URL")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required",
		})
	}

	// Basic URL validation
	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || !parsedURL.IsAbs() || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		log.Printf("Invalid original URL provided: %s", req.URL)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid original URL. Must be a valid absolute HTTP/HTTPS URL.",
		})
	}

	if req.CustomSlug != nil && *req.CustomSlug == "" {
		log.Println("Custom slug provided but is empty")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Custom slug cannot be empty if provided",
		})
	}
	// Simple validation for custom slug length
	if req.CustomSlug != nil {
		if len(*req.CustomSlug) < 3 || len(*req.CustomSlug) > 20 {
			log.Printf("Invalid custom slug length: %s", *req.CustomSlug)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Custom slug must be between 3 and 20 characters.",
			})
		}
	}

	var shortCode string
	if req.CustomSlug != nil {
		shortCode = *req.CustomSlug
	} else {
		shortCode = utils.GenerateShortCode(6)
	}

	urlToSave := model.URL{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		ClickCount:  0,
		CustomSlug:  req.CustomSlug,
	}

	err = repo.SaveURL(context.Background(), urlToSave)
	if err != nil {
		if err.Error() == fmt.Sprintf("custom slug '%s' already exists", shortCode) {
			log.Printf("Custom slug conflict: %s", shortCode)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		log.Printf("Failed to save URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to shorten URL due to a server error.",
		})
	}

	// Cache the mapping with a default fixed expiration
	cacheExpiration := 24 * time.Hour
	err = urlCache.Set(context.Background(), shortCode, req.URL, cacheExpiration)
	if err != nil {
		log.Printf("Failed to cache URL after saving to DB: %v", err)
	}

	// Use the serviceBaseURL for the response
	shortURL := fmt.Sprintf("%s/%s", serviceBaseURL, shortCode)
	log.Printf("URL shortened successfully: %s -> %s", shortCode, req.URL)
	return c.JSON(model.URLResponse{
		ShortURL: shortURL,
	})
}

func Redirect(c *fiber.Ctx) error {
	code := c.Params("code")
	log.Printf("Attempting to redirect URL: %s", code)

	originalURL, err := urlCache.Get(context.Background(), code)
	if err == nil {
		log.Printf("Redirecting from cache for code: %s", code)
		go func() {
			incErr := repo.IncrementClickCount(context.Background(), code)
			if incErr != nil {
				log.Printf("Failed to increment click count from cache hit for %s: %v", code, incErr)
			}
		}()
		return c.Redirect(originalURL, fiber.StatusFound)
	}

	url, err := repo.GetURL(context.Background(), code)
	if err != nil {
		log.Printf("Short URL not found in DB or cache: %s, Error: %v", code, err)
		return c.Status(fiber.StatusNotFound).SendString("Short URL not found.")
	}

	// Found in DB, cache it for future requests
	cacheExpiration := 24 * time.Hour
	err = urlCache.Set(context.Background(), url.ShortCode, url.OriginalURL, cacheExpiration)
	if err != nil {
		log.Printf("Failed to cache URL after DB lookup for %s: %v", url.ShortCode, err)
	}

	// Increment click count asynchronously
	go func() {
		incErr := repo.IncrementClickCount(context.Background(), url.ShortCode)
		if incErr != nil {
			log.Printf("Failed to increment click count from DB hit for %s: %v", url.ShortCode, incErr)
		}
	}()

	log.Printf("Redirecting from DB for code: %s -> %s", url.ShortCode, url.OriginalURL)
	return c.Redirect(url.OriginalURL, fiber.StatusFound)
}
