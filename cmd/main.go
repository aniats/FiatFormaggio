package main

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/service/cbr"
	"github.com/aniats/FiatFormaggio/internal/service/currency"
	"github.com/aniats/FiatFormaggio/internal/service/finance"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	_ "net/http/pprof"

	"github.com/aniats/FiatFormaggio/internal/app"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/metrics"
	"github.com/aniats/FiatFormaggio/internal/repository"
	"github.com/aniats/FiatFormaggio/internal/repository/postgres"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	tracingCleanup, err := app.InitTracing()
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer tracingCleanup()

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
	currencyService := initCurrencyService(repo, cbrService)
	financeService := finance.NewFinanceService(repo, currencyService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := currencyService.Start(ctx); err != nil {
		log.Fatalf("Failed to start currency service: %v", err)
	}
	defer currencyService.Stop()

	if err := testServices(ctx, financeService, currencyService); err != nil {
		log.Printf("Service test error: %v", err)
	}

	metrics.Init()
	go startMetricsServer()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if err := startBot(ctx, token, financeService, currencyService); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	log.Println("Application started successfully")
}

func initRepository() (repository.Repository, error) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(context.Background(), "initRepository")
	defer span.End()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	repo, err := postgres.New(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := repo.HealthCheck(healthCtx); err != nil {
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

func initCurrencyService(repo repository.Repository, cbrService *cbr.CBRService) *currency.CachedCurrencyService {
	currencyService := currency.NewCachedCurrencyService(repo, cbrService)
	log.Println("Currency caching service initialized successfully")
	return currencyService
}

func testServices(ctx context.Context, financeService *finance.FinanceService, currencyService *currency.CachedCurrencyService) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "testServices")
	defer span.End()

	userID := domain.UserId(123456789)
	deposits, err := financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get deposits: %w", err)
	}
	log.Printf("Successfully retrieved %d deposits for user %d", len(deposits), userID)

	rates, err := currencyService.GetCurrencyRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get currency rates: %w", err)
	}
	log.Printf("Successfully retrieved cached currency rates: %d currencies", len(rates))

	return nil
}

func startBot(ctx context.Context, token string, financeService *finance.FinanceService, currencyService *currency.CachedCurrencyService) error {
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := app.NewBotFromToken(token, financeService, currencyService)
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

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Metrics server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}
