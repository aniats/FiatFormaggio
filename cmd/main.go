package main

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/service/cbr"
	"github.com/aniats/FiatFormaggio/internal/service/finance"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/aniats/FiatFormaggio/internal/app"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
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
	financeService := finance.NewFinanceService(repo, cbrService) // Now uses repository.Repository interface

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := testServices(ctx, financeService, cbrService); err != nil {
		log.Printf("Service test error: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if err := startBot(ctx, token, financeService); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	log.Println("Application started successfully")
}

func initRepository() (repository.Repository, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	repo, err := repository.NewPostgresRepository(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("repository health check failed: %w", err)
	}

	log.Println("Repository initialized successfully")
	return repo, nil
}

func initCBRService() *cbr.CBRService {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	cbrService := cbr.NewCBRService(httpClient)
	log.Println("CBR service initialized successfully")
	return cbrService
}

func testServices(ctx context.Context, financeService *finance.FinanceService, cbrService *cbr.CBRService) error {
	userID := domain.UserId(123456789)
	deposits, err := financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get deposits: %w", err)
	}
	log.Printf("Successfully retrieved %d deposits for user %d", len(deposits), userID)

	rates, err := cbrService.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get currency rates: %w", err)
	}
	log.Printf("Successfully retrieved currency rates: %d currencies", len(rates))

	return nil
}

func startBot(ctx context.Context, token string, financeService *finance.FinanceService) error {
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := app.NewBotFromToken(token, financeService)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	defer bot.Stop()

	go func() {
		if err := bot.Start(ctx); err != nil {
			log.Printf("Bot error: %v", err)
		}
	}()

	log.Println("Bot started successfully")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")

	return nil
}
