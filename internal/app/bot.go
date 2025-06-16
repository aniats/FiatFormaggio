package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"

	"log"
	"strings"
	"sync"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	// UpdateTimeoutSeconds defines timeout for getting updates from Telegram API
	UpdateTimeoutSeconds = 60
	// DefaultWorkerPoolSize defines the number of worker goroutines for message processing
	DefaultWorkerPoolSize = 10
	// WorkerChannelBufferSize defines the buffer size for the worker channel
	WorkerChannelBufferSize = 100
)

type BotAPI interface {
	GetLastEvents() <-chan tgbotapi.Update
	SendMessage(chatID int64, text string) error
	Close()
}

type FinanceService interface {
	GetDepositsByUserID(ctx context.Context, userID domain.UserId) ([]domain.Deposit, error)
	CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*domain.Deposit, error)
	EnsureUserExists(ctx context.Context, userID domain.UserId, username string) (bool, error)
}

type CBRService interface {
	GetCurrencyRates(ctx context.Context, date time.Time) ([]*domain.CurrencyRateCBR, error)
}

type Bot struct {
	botAPI         BotAPI
	financeService FinanceService
	cbrService     CBRService
	workerPool     *WorkerPool
	sessionManager *UserSessionManager
}

type WorkerPool struct {
	workers    int
	jobChannel chan Job
	wg         sync.WaitGroup
}

type Job struct {
	ctx     context.Context
	message *tgbotapi.Message
	bot     *Bot
}

func NewBot(botAPI BotAPI, financeService FinanceService, cbrService CBRService) *Bot {
	return &Bot{
		botAPI:         botAPI,
		financeService: financeService,
		cbrService:     cbrService,
		workerPool:     NewWorkerPool(DefaultWorkerPoolSize),
		sessionManager: NewUserSessionManager(),
	}
}

func NewBotFromToken(token string, financeService FinanceService, cbrService CBRService) (*Bot, error) {
	botAPI, err := NewTelegramBotAPI(token)
	if err != nil {
		return nil, err
	}

	return NewBot(botAPI, financeService, cbrService), nil
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

			job := Job{
				ctx:     ctx,
				message: update.Message,
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

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	command := strings.ToLower(message.Command())
	chatID := message.Chat.ID
	userID := domain.UserId(message.From.ID)

	telegramUsername := message.From.UserName

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
	case "start", "help":
		b.handleHelp(chatID)
	case "deposits", "депозиты", "вклады":
		b.handleDepositsCommand(ctx, chatID, domain.UserId(message.From.ID))
	case "create_deposit", "добавить_депозит":
		b.startSession(ctx, message, SessionCreateDeposit)
	case "rates", "курсы", "валюты":
		b.handleCurrencyRatesCommand(ctx, chatID)
	default:
		b.sendMessage(chatID, "Неизвестная команда. Введите /help для списка команд.")
	}
}

func (b *Bot) handleHelp(chatID int64) {
	helpText := `Доступные команды:
		/total - Общий баланс
		/deposits - Показать депозиты
		/create_deposit - Создать депозит
		/rates - Курсы валют ЦБ РФ
	`

	b.sendMessage(chatID, helpText)
}

func (b *Bot) startSession(ctx context.Context, msg *tgbotapi.Message, sessionType SessionType) {
	userID := domain.UserId(msg.From.ID)
	chatID := msg.Chat.ID

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

func (b *Bot) handleSessionMessage(ctx context.Context, msg *tgbotapi.Message, session *UserSession) {
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

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers:    workers,
		jobChannel: make(chan Job, WorkerChannelBufferSize),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
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

func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
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

			// Process the job
			job.bot.handleMessage(job.ctx, job.message)
		}
	}
}
