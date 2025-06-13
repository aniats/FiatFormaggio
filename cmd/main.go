package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
	"github.com/aniats/FiatFormaggio/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	repo, err := initRepository()
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	cbrService := initCBRService()
	financeService := service.NewFinanceService(repo, cbrService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := testServices(ctx, financeService, cbrService); err != nil {
		log.Printf("Service test error: %v", err)
	}

	// Initialize and start bot (when ready)
	// token := os.Getenv("TELEGRAM_BOT_TOKEN")
	// if err := startBot(ctx, token, financeService); err != nil {
	// 	log.Fatalf("Failed to start bot: %v", err)
	// }

	log.Println("Application started successfully")
}

// initRepository initializes the repository with proper error handling
func initRepository() (repository.Repository, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	repo, err := repository.NewPostgresRepository(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("repository health check failed: %w", err)
	}

	log.Println("Repository initialized successfully")
	return repo, nil
}

// initCBRService initializes the CBR service
func initCBRService() *service.CBRService {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	cbrService := service.NewCBRService(httpClient)
	log.Println("CBR service initialized successfully")
	return cbrService
}

// testServices performs basic tests on the services
func testServices(ctx context.Context, financeService *service.FinanceService, cbrService *service.CBRService) error {
	// Test finance service
	userID := domain.UserId(123456789)
	deposits, err := financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get deposits: %w", err)
	}
	log.Printf("Successfully retrieved %d deposits for user %d", len(deposits), userID)

	// Test CBR service
	rates, err := cbrService.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get currency rates: %w", err)
	}
	log.Printf("Successfully retrieved currency rates: %d currencies", len(rates))

	return nil
}

func startBot(ctx context.Context, token string, financeService *service.FinanceService) error {
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Uncomment when bot package is ready
	// bot, err := app.NewBot(token, financeService)
	// if err != nil {
	// 	return fmt.Errorf("failed to create bot: %w", err)
	// }

	// go func() {
	// 	if err := bot.Start(ctx); err != nil {
	// 		log.Printf("Bot error: %v", err)
	// 	}
	// }()

	log.Println("Bot started successfully")
	return nil
}
