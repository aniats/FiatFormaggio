package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aniats/FiatFormaggio/internal/app"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/metrics"
	"github.com/aniats/FiatFormaggio/internal/repository"
	"github.com/aniats/FiatFormaggio/internal/repository/postgres"
	"github.com/aniats/FiatFormaggio/internal/service/cbr"
	"github.com/aniats/FiatFormaggio/internal/service/currency"
	"github.com/aniats/FiatFormaggio/internal/service/finance"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	_ "net/http/pprof"
)

type Application struct {
	repository     repository.Repository
	tracingCleanup func()

	cbrService      *cbr.CBRService
	currencyService *currency.CachedCurrencyService
	financeService  *finance.FinanceService

	bot *app.Bot

	config *Config
}

type Config struct {
	DatabaseURL      string
	TelegramBotToken string
	MetricsPort      string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config := loadConfig()

	application, err := NewApplication(config)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	defer application.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return application.WaitForShutdown()
}

func NewApplication(config *Config) (*Application, error) {
	app := &Application{
		config: config,
	}

	if err := app.initTracing(); err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	if err := app.initRepository(); err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	if err := app.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := app.initBot(); err != nil {
		return nil, fmt.Errorf("failed to initialize bot: %w", err)
	}

	log.Println("✅ Application initialized successfully")
	return app, nil
}

func (a *Application) Start(ctx context.Context) error {
	if err := a.currencyService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start currency service: %w", err)
	}

	if err := a.testServices(ctx); err != nil {
		log.Printf("Service test error: %v", err)
	}

	metrics.Init()
	go a.startMetricsServer()

	go func() {
		if err := a.bot.Start(ctx); err != nil {
			log.Printf("Bot error: %v", err)
		}
	}()

	log.Println("✅ Application started successfully")
	return nil
}

func (a *Application) WaitForShutdown() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🔄 Shutting down...")
	return nil
}

func (a *Application) Cleanup() {
	if a.bot != nil {
		a.bot.Stop()
	}

	if a.currencyService != nil {
		a.currencyService.Stop()
	}

	if a.repository != nil {
		if err := a.repository.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}

	if a.tracingCleanup != nil {
		a.tracingCleanup()
	}

	log.Println("✅ Application cleanup completed")
}

// initTracing initializes distributed tracing
func (a *Application) initTracing() error {
	cleanup, err := app.InitTracing()
	if err != nil {
		return err
	}
	a.tracingCleanup = cleanup
	return nil
}

func (a *Application) initRepository() error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(context.Background(), "Application.initRepository")
	defer span.End()

	if a.config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	repo, err := postgres.New(a.config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := repo.HealthCheck(healthCtx); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}

	a.repository = repo
	log.Println("✅ Repository initialized successfully")
	return nil
}

func (a *Application) initServices() error {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	a.cbrService = cbr.NewCBRService(httpClient)
	log.Println("✅ CBR service initialized successfully")

	a.currencyService = currency.NewCachedCurrencyService(a.repository, a.cbrService)
	log.Println("✅ Currency caching service initialized successfully")

	a.financeService = finance.NewFinanceService(a.repository, a.currencyService)
	log.Println("✅ Finance service initialized successfully")

	return nil
}

func (a *Application) initBot() error {
	if a.config.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	bot, err := app.NewBotFromToken(a.config.TelegramBotToken, a.financeService, a.currencyService)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	a.bot = bot
	log.Println("✅ Bot initialized successfully")
	return nil
}

func (a *Application) testServices(ctx context.Context) error {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Application.testServices")
	defer span.End()

	userID := domain.UserId(123456789)
	deposits, err := a.financeService.GetDepositsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get deposits: %w", err)
	}
	log.Printf("✅ Successfully retrieved %s deposits for user %s",
		app.FormatInteger(int64(len(deposits))),
		app.FormatInteger(int64(userID)))

	rates, err := a.currencyService.GetCurrencyRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get currency rates: %w", err)
	}
	log.Printf("✅ Successfully retrieved cached currency rates: %s currencies",
		app.FormatInteger(int64(len(rates))))

	return nil
}

func (a *Application) startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("📊 Metrics server starting on %s", a.config.MetricsPort)
	if err := http.ListenAndServe(a.config.MetricsPort, nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}

func loadConfig() *Config {
	return &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		MetricsPort:      getEnvWithDefault("METRICS_PORT", ":8080"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
