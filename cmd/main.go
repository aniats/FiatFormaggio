package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aniats/FiatFormaggio/internal/service"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "fiatuser"
	dbPassword = "fiatpassword"
	dbName     = "fiatdb"
)

func main() {
	// token := os.Getenv("TELEGRAM_BOT_TOKEN")
	// if token == "" {
	// 	log.Fatal("TELEGRAM_BOT_TOKEN is not present")
	// }

	// connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	// 	dbHost, dbPort, dbUser, dbPassword, dbName)

	// _, err := repository.NewPostgresRepository(connStr)
	// if err != nil {
	// 	log.Fatalf("Couldn't connect to database: %v", err)
	// }
	cbrService := service.NewCBRService(&http.Client{Timeout: 10 * time.Second})
	// financeService := service.NewFinanceService(repo, cbrService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := cbrService.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		fmt.Println(result)
	}

	// bot, err := app.NewBot(token, financeService)
	// if err != nil {
	// 	log.Fatalf("Error on creating bot: %v", err)
	// }

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// go bot.Start(ctx)
}
