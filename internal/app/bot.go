package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// UpdateTimeoutSeconds defines timeout for getting updates from Telegram API
	UpdateTimeoutSeconds = 60
	// DefaultWorkerPoolSize defines the number of worker goroutines for message processing
	DefaultWorkerPoolSize = 10
	// WorkerChannelBufferSize defines the buffer size for the worker channel
	WorkerChannelBufferSize = 100
)

type CommandType string

const (
	CommandStart                    CommandType = "start"
	CommandHelp                     CommandType = "help"
	CommandTotal                    CommandType = "total"
	CommandTotalRu                  CommandType = "общий_баланс"
	CommandDeposits                 CommandType = "deposits"
	CommandDepositsRu1              CommandType = "депозиты"
	CommandDepositsRu2              CommandType = "вклады"
	CommandCreateDeposit            CommandType = "create_deposit"
	CommandCreateDepositRu          CommandType = "добавить_депозит"
	CommandBrokerageAccounts        CommandType = "brokerage_accounts"
	CommandBrokerageAccountsRu1     CommandType = "брокерские_счета"
	CommandBrokerageAccountsRu2     CommandType = "счета"
	CommandCreateBrokerageAccount   CommandType = "create_brokerage_account"
	CommandCreateBrokerageAccountRu CommandType = "создать_брокерский_счет"
	CommandSavingAccounts           CommandType = "saving_accounts"
	CommandSavingAccountsRu1        CommandType = "накопительные_счета"
	CommandSavingAccountsRu2        CommandType = "накопления"
	CommandCreateSavingAccount      CommandType = "create_saving_account"
	CommandCreateSavingAccountRu    CommandType = "создать_накопительный_счет"
	CommandCashHoldings             CommandType = "cash_holdings"
	CommandCashHoldingsRu1          CommandType = "наличные"
	CommandCashHoldingsRu2          CommandType = "наличные_счета"
	CommandCreateCashHolding        CommandType = "create_cash_holding"
	CommandCreateCashHoldingRu      CommandType = "создать_наличный_счет"
	CommandRates                    CommandType = "rates"
	CommandRatesRu1                 CommandType = "курсы"
	CommandRatesRu2                 CommandType = "валюты"
)

type BotAPI interface {
	GetLastEvents() <-chan tgBotAPI.Update
	SendMessage(chatID int64, text string) error
	Close()
}

type FinanceService interface {
	GetDepositsByUserID(ctx context.Context, userID domain.UserId) ([]domain.Deposit, error)
	CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*domain.Deposit, error)
	GetBrokerageAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.BrokerageAccount, error)
	CreateBrokerageAccount(ctx context.Context, req *models.CreateBrokerageAccountRequest) (*domain.BrokerageAccount, error)
	GetSavingAccountsByUserID(ctx context.Context, userID domain.UserId) ([]domain.SavingAccount, error)
	CreateSavingAccount(ctx context.Context, req *models.CreateSavingAccountRequest) (*domain.SavingAccount, error)
	GetCashHoldingsByUserID(ctx context.Context, userID domain.UserId) ([]domain.CashHolding, error)
	CreateCashHolding(ctx context.Context, req *models.CreateCashHoldingRequest) (*domain.CashHolding, error)
	EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error)
}

type CurrencyService interface {
	GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error)
}

type Bot struct {
	botAPI          BotAPI
	financeService  FinanceService
	currencyService CurrencyService
	workerPool      *WorkerPool
	sessionManager  *UserSessionManager
}

type WorkerPool struct {
	workers    int64
	jobChannel chan Job
	wg         sync.WaitGroup
}

type Message struct {
	ChatID   int64
	UserID   int64
	Username string
	Text     string
}

func (m *Message) Command() string {
	if len(m.Text) == 0 || m.Text[0] != '/' {
		return ""
	}

	parts := strings.Fields(m.Text)
	if len(parts) == 0 {
		return ""
	}

	command := parts[0][1:] // Remove the '/' prefix
	return command
}

type Job struct {
	ctx     context.Context
	message *Message
	bot     *Bot
}

func NewBot(botAPI BotAPI, financeService FinanceService, currencyService CurrencyService) *Bot {
	return &Bot{
		botAPI:          botAPI,
		financeService:  financeService,
		currencyService: currencyService,
		workerPool:      NewWorkerPool(DefaultWorkerPoolSize),
		sessionManager:  NewUserSessionManager(),
	}
}

func NewBotFromToken(token string, financeService FinanceService, currencyService CurrencyService) (*Bot, error) {
	botAPI, err := NewTelegramBotAPI(token)
	if err != nil {
		return nil, err
	}

	return NewBot(botAPI, financeService, currencyService), nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.workerPool.Start(ctx)
	defer b.workerPool.Stop()

	updates := b.botAPI.GetLastEvents()

	log.Printf("Bot started with %d workers", b.workerPool.workers)

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopping...")
			return ctx.Err()
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			msg := &Message{
				ChatID:   update.Message.Chat.ID,
				UserID:   update.Message.From.ID,
				Username: update.Message.From.UserName,
				Text:     update.Message.Text,
			}

			job := Job{
				ctx:     ctx,
				message: msg,
				bot:     b,
			}

			select {
			case b.workerPool.jobChannel <- job:
			case <-ctx.Done():
				return ctx.Err()
			default:
				log.Printf("Worker pool is full, dropping message from chat %d", update.Message.Chat.ID)
			}
		}
	}
}

func (b *Bot) Stop() {
	b.botAPI.Close()
	b.workerPool.Stop()
}

func (b *Bot) sendMessage(chatID int64, text string) {
	if err := b.botAPI.SendMessage(chatID, text); err != nil {
		log.Printf("Error sending message to chat %d: %v", chatID, err)
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *Message) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.handleMessage")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", message.UserID),
		attribute.Int64("chat.id", message.ChatID),
		attribute.String("message.command", message.Command()),
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	command := CommandType(strings.ToLower(message.Command()))
	chatID := message.ChatID
	userID := domain.UserId(message.UserID)

	telegramUsername := message.Username

	isNewUser, err := b.financeService.EnsureUserExists(ctx, userID, telegramUsername)
	if err != nil {
		log.Printf("Failed to ensure user exists for userID %d: %v", userID, err)
		b.sendMessage(chatID, "❌ Ошибка инициализации пользователя. Попробуйте позже.")
		return
	}

	if isNewUser {
		b.sendMessage(chatID, "🎉 Добро пожаловать! Вы зарегистрированы в системе. Введите /help для начала работы.")
	}

	if session := b.sessionManager.GetSession(userID); session != nil {
		b.handleSessionMessage(ctx, message, session)
		return
	}

	switch command {
	case CommandStart, CommandHelp:
		b.sendHelp(chatID)
	case CommandTotal, CommandTotalRu:
		b.sendTotalBalanceCommand(ctx, chatID, domain.UserId(message.UserID))
	case CommandDeposits, CommandDepositsRu1, CommandDepositsRu2:
		b.sendDepositsCommand(ctx, chatID, domain.UserId(message.UserID))
	case CommandCreateDeposit, CommandCreateDepositRu:
		b.startSession(ctx, message, SessionCreateDeposit)
	case CommandBrokerageAccounts, CommandBrokerageAccountsRu1, CommandBrokerageAccountsRu2:
		b.sendBrokerageAccountsCommand(ctx, chatID, domain.UserId(message.UserID))
	case CommandCreateBrokerageAccount, CommandCreateBrokerageAccountRu:
		b.startSession(ctx, message, SessionCreateBrokerageAccountSession)
	case CommandSavingAccounts, CommandSavingAccountsRu1, CommandSavingAccountsRu2:
		b.sendSavingAccountsCommand(ctx, chatID, domain.UserId(message.UserID))
	case CommandCreateSavingAccount, CommandCreateSavingAccountRu:
		b.startSession(ctx, message, SessionCreateSavingAccount)
	case CommandCashHoldings, CommandCashHoldingsRu1, CommandCashHoldingsRu2:
		b.sendCashHoldingsCommand(ctx, chatID, domain.UserId(message.UserID))
	case CommandCreateCashHolding, CommandCreateCashHoldingRu:
		b.startSession(ctx, message, SessionCreateCashHolding)
	case CommandRates, CommandRatesRu1, CommandRatesRu2:
		b.sendCurrencyRatesCommand(ctx, chatID)
	default:
		b.sendMessage(chatID, "Неизвестная команда. Введите /help для списка команд.")
	}
}

func (b *Bot) startSession(ctx context.Context, msg *Message, sessionType SessionType) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.startSession")
	defer span.End()

	userID := domain.UserId(msg.UserID)
	span.SetAttributes(
		attribute.Int64("user.id", msg.UserID),
		attribute.String("session.type", string(sessionType)),
	)
	chatID := msg.ChatID

	session, err := b.sessionManager.StartSession(userID, chatID, sessionType)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Не удалось начать сессию: %s", err.Error()))
		return
	}

	if err := session.Handler.HandleStep(ctx, b, session, msg); err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		b.sessionManager.ClearSession(userID)
	}
}

func (b *Bot) handleSessionMessage(ctx context.Context, msg *Message, session *UserSession) {
	tracer := otel.Tracer("fiat-formaggio")
	ctx, span := tracer.Start(ctx, "Bot.handleSessionMessage")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", int64(session.UserID)),
		attribute.String("session.step", string(session.CurrentStep)),
		attribute.String("session.type", string(session.Type)),
	)

	b.sessionManager.UpdateLastActivity(session.UserID)

	if strings.ToLower(msg.Text) == "/cancel" || strings.ToLower(msg.Text) == "отмена" {
		b.cancelSession(session.UserID, session.ChatID)
		return
	}

	if err := session.Handler.HandleStep(ctx, b, session, msg); err != nil {
		b.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
	}
}

func (b *Bot) cancelSession(userID domain.UserId, chatID int64) {
	b.sessionManager.ClearSession(userID)
	b.sendMessage(chatID, "❌ Операция отменена.")
}

func NewWorkerPool(workers int64) *WorkerPool {
	return &WorkerPool{
		workers:    workers,
		jobChannel: make(chan Job, WorkerChannelBufferSize),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := int64(0); i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i+1)
	}
	log.Printf("Started %d workers", wp.workers)
}

func (wp *WorkerPool) Stop() {
	close(wp.jobChannel)
	wp.wg.Wait()
	log.Println("All workers stopped")
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int64) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping due to context cancellation", workerID)
			return
		case job, ok := <-wp.jobChannel:
			if !ok {
				log.Printf("Worker %d stopping due to closed channel", workerID)
				return
			}

			job.bot.handleMessage(job.ctx, job.message)
		}
	}
}
