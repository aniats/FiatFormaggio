package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/metrics"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"github.com/aniats/FiatFormaggio/internal/service/chatgpt"
	"github.com/aniats/FiatFormaggio/internal/service/finance/models"
	"github.com/aniats/FiatFormaggio/internal/tracing"
	"github.com/aniats/FiatFormaggio/internal/utils"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	UpdateTimeoutSeconds    = 60
	DefaultWorkerPoolSize   = 10
	WorkerChannelBufferSize = 100
)

type BotAPI interface {
	GetLastEvents() <-chan tgBotAPI.Update
	SendMessage(chatID int64, text string) error
	SendMessageWithKeyboard(chatID int64, text string, keyboard tgBotAPI.InlineKeyboardMarkup) error
	SetMyCommands(commands []tgBotAPI.BotCommand) error
	Close()
}

type FinanceService interface {
	GetDepositsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.Deposit, error)
	CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*domain.Deposit, error)
	GetBrokerageAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.BrokerageAccount, error)
	CreateBrokerageAccount(ctx context.Context, req *models.CreateBrokerageAccountRequest) (*domain.BrokerageAccount, error)
	GetSavingAccountsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.SavingAccount, error)
	CreateSavingAccount(ctx context.Context, req *models.CreateSavingAccountRequest) (*domain.SavingAccount, error)
	GetCashHoldingsByUserID(ctx context.Context, UserID domain.UserID) ([]domain.CashHolding, error)
	CreateCashHolding(ctx context.Context, req *models.CreateCashHoldingRequest) (*domain.CashHolding, error)
	EnsureUserExists(ctx context.Context, UserID domain.UserID, username string) (bool, error)
	GetTotalBalance(ctx context.Context, UserID domain.UserID) (float64, error)
	GetFinancialAdvice(ctx context.Context, userID domain.UserID, userMessage string, previousMessages []chatgpt.ChatMessage) (string, error)
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
	interceptor     *middleware.Interceptor
	errorHandler    *errors.ErrorHandler
	appName         string
}

type WorkerPool struct {
	workers    int64
	jobChannel chan Job
	wg         sync.WaitGroup
}

type Job struct {
	ctx     context.Context
	message *domain.Message
	bot     *Bot
	span    trace.Span
}

func NewBot(botAPI BotAPI, financeService FinanceService, currencyService CurrencyService, appName string) *Bot {
	interceptor := middleware.NewInterceptor(middleware.DefaultConfig(appName+"-bot"), appName)
	errorHandler := errors.DefaultErrorHandler()

	return &Bot{
		botAPI:          botAPI,
		financeService:  financeService,
		currencyService: currencyService,
		workerPool:      NewWorkerPool(DefaultWorkerPoolSize),
		sessionManager:  NewUserSessionManager(financeService),
		interceptor:     interceptor,
		errorHandler:    errorHandler,
		appName:         appName,
	}
}

func NewBotFromToken(token string, financeService FinanceService, currencyService CurrencyService, appName string) (*Bot, error) {
	botAPI, err := NewTelegramBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot := NewBot(botAPI, financeService, currencyService, appName)

	if err = bot.setupBotCommands(); err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.workerPool.Start(ctx)
	defer b.workerPool.Stop()

	updates := b.botAPI.GetLastEvents()
	log.Printf("Bot started with %s workers", utils.FormatInteger(b.workerPool.workers))

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopping...")
			return ctx.Err()
		case update := <-updates:
			if update.Message != nil {
				msg := &domain.Message{
					ChatID:   update.Message.Chat.ID,
					UserID:   update.Message.From.ID,
					Username: update.Message.From.UserName,
					Text:     update.Message.Text,
				}

				tracer := otel.Tracer(b.appName)
				msgCtx, span := tracer.Start(ctx, "Bot.ProcessTelegramMessage")
				span.SetAttributes(
					attribute.Int64(tracing.UserID, msg.UserID),
					attribute.Int64(tracing.ChatID, msg.ChatID),
					attribute.String(tracing.MessageText, msg.Text),
				)

				job := Job{
					ctx:     msgCtx,
					message: msg,
					bot:     b,
					span:    span,
				}

				select {
				case b.workerPool.jobChannel <- job:
				case <-ctx.Done():
					return ctx.Err()
				default:
					log.Printf("Worker pool is full, dropping message from chat %s", utils.FormatInteger(update.Message.Chat.ID))
				}
			} else if update.CallbackQuery != nil {
				b.handleCallbackQuery(ctx, update.CallbackQuery)
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

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgBotAPI.InlineKeyboardMarkup) {
	if err := b.botAPI.SendMessageWithKeyboard(chatID, text, keyboard); err != nil {
		log.Printf("Error sending message with keyboard to chat %d: %v", chatID, err)
	}
}

func (b *Bot) sendMainMenu(chatID int64) {
	text := `🏠 Главное меню

	Используйте кнопки ниже для навигации:
	• 💰 Просмотр балансов и счетов
	• 💡 ИИ советы по финансам
	• ➕ Создание новых счетов
	• 💱 Актуальные курсы валют`

	keyboard := CreateMainMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) handleMenuCallback(ctx context.Context, message *domain.Message) bool {
	input := message.Text
	chatID := message.ChatID
	UserID := domain.UserID(message.UserID)

	switch input {
	case domain.CallbackMenuTotal:
		b.sendTotalBalanceCommand(ctx, chatID, UserID)
		return true
	case domain.CallbackMenuDeposits:
		b.sendDepositsCommand(ctx, chatID, UserID)
		return true
	case domain.CallbackMenuCreateDeposit:
		b.startSession(ctx, message, SessionCreateDeposit)
		return true
	case domain.CallbackMenuBrokerage:
		b.sendBrokerageAccountsCommand(ctx, chatID, UserID)
		return true
	case domain.CallbackMenuCreateBrokerage:
		b.startSession(ctx, message, SessionCreateBrokerageAccountSession)
		return true
	case domain.CallbackMenuSaving:
		b.sendSavingAccountsCommand(ctx, chatID, UserID)
		return true
	case domain.CallbackMenuCreateSaving:
		b.startSession(ctx, message, SessionCreateSavingAccount)
		return true
	case domain.CallbackMenuCash:
		b.sendCashHoldingsCommand(ctx, chatID, UserID)
		return true
	case domain.CallbackMenuCreateCash:
		b.startSession(ctx, message, SessionCreateCashHolding)
		return true
	case domain.CallbackMenuRates:
		b.sendCurrencyRatesCommand(ctx, chatID)
		return true
	case domain.CallbackMenuFinancialAdvice:
		b.sendFinancialAdviceCommand(ctx, message)
		return true
	}
	return false
}

func CreateConfirmationKeyboard(confirmAction, cancelAction string) tgBotAPI.InlineKeyboardMarkup {
	return tgBotAPI.NewInlineKeyboardMarkup(
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("✅ Да", confirmAction),
			tgBotAPI.NewInlineKeyboardButtonData("❌ Нет", cancelAction),
		),
	)
}

func CreateMainMenuKeyboard() tgBotAPI.InlineKeyboardMarkup {
	return tgBotAPI.NewInlineKeyboardMarkup(
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("💰 Общий баланс", domain.CallbackMenuTotal),
			tgBotAPI.NewInlineKeyboardButtonData("💱 Курсы валют", domain.CallbackMenuRates),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("💡 Финансовые советы", domain.CallbackMenuFinancialAdvice),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("💳 Депозиты", domain.CallbackMenuDeposits),
			tgBotAPI.NewInlineKeyboardButtonData("➕ Создать депозит", domain.CallbackMenuCreateDeposit),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("📈 Брокерские счета", domain.CallbackMenuBrokerage),
			tgBotAPI.NewInlineKeyboardButtonData("➕ Создать брокерский", domain.CallbackMenuCreateBrokerage),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("🏦 Накопительные", domain.CallbackMenuSaving),
			tgBotAPI.NewInlineKeyboardButtonData("➕ Создать накопительный", domain.CallbackMenuCreateSaving),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("💵 Наличные", domain.CallbackMenuCash),
			tgBotAPI.NewInlineKeyboardButtonData("➕ Создать наличные", domain.CallbackMenuCreateCash),
		),
	)
}

func CreateCurrencySelectionKeyboard() tgBotAPI.InlineKeyboardMarkup {
	return tgBotAPI.NewInlineKeyboardMarkup(
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("RUB - Рубль", domain.CallbackCurrencyRUB),
			tgBotAPI.NewInlineKeyboardButtonData("USD - Доллар", domain.CallbackCurrencyUSD),
			tgBotAPI.NewInlineKeyboardButtonData("EUR - Евро", domain.CallbackCurrencyEUR),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("GBP - Фунт", domain.CallbackCurrencyGBP),
			tgBotAPI.NewInlineKeyboardButtonData("JPY - Иена", domain.CallbackCurrencyJPY),
			tgBotAPI.NewInlineKeyboardButtonData("CNY - Юань", domain.CallbackCurrencyCNY),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("RSD - Динар", domain.CallbackCurrencyRSD),
			tgBotAPI.NewInlineKeyboardButtonData("XBT - Биткоин", domain.CallbackCurrencyXBT),
			tgBotAPI.NewInlineKeyboardButtonData("KZT - Тенге", domain.CallbackCurrencyKZT),
		),
		tgBotAPI.NewInlineKeyboardRow(
			tgBotAPI.NewInlineKeyboardButtonData("Пропустить (RUB)", domain.CallbackCurrencySkip),
		),
	)
}

func (b *Bot) sendErrorMessage(ctx context.Context, chatID int64, err error, operation string) {
	userMsg := b.errorHandler.Handle(ctx, err, operation)
	b.sendMessage(chatID, userMsg)
}

func (b *Bot) handleCallbackQuery(ctx context.Context, callbackQuery *tgBotAPI.CallbackQuery) {
	UserID := domain.UserID(callbackQuery.From.ID)
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	msg := &domain.Message{
		ChatID:   chatID,
		UserID:   int64(UserID),
		Username: callbackQuery.From.UserName,
		Text:     data,
	}

	tracer := otel.Tracer(b.appName)
	msgCtx, span := tracer.Start(ctx, "Bot.ProcessCallbackQuery")
	span.SetAttributes(
		attribute.Int64(tracing.UserID, int64(UserID)),
		attribute.Int64(tracing.ChatID, chatID),
		attribute.String("callback.data", data),
	)

	job := Job{
		ctx:     msgCtx,
		message: msg,
		bot:     b,
		span:    span,
	}

	select {
	case b.workerPool.jobChannel <- job:
	default:
		log.Printf("Worker pool is full, dropping callback query from chat %s", utils.FormatInteger(chatID))
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *domain.Message) {
	handler := func(ctx context.Context, msg interface{}) error {
		if message, ok := msg.(*domain.Message); ok {
			b.processMessage(ctx, message)
		}
		return nil
	}

	middlewareHandler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, handler(ctx, input)
	}
	wrappedHandler := b.interceptor.Chain(middlewareHandler, "Bot.handleMessage")
	_, err := wrappedHandler(ctx, message)
	if err != nil {
		log.Printf("Error processing message: %v", err)
	}
}

func (b *Bot) processMessage(ctx context.Context, message *domain.Message) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	chatID := message.ChatID
	UserID := domain.UserID(message.UserID)
	telegramUsername := message.Username

	isNewUser, err := b.financeService.EnsureUserExists(ctx, UserID, telegramUsername)
	if err != nil {
		log.Printf("Failed to ensure user exists for UserID %s: %v", utils.FormatInteger(int64(UserID)), err)
		b.sendMessage(chatID, "❌ Ошибка инициализации пользователя. Попробуйте позже.")
		return
	}

	if isNewUser {
		b.sendMessage(chatID, "🎉 Добро пожаловать! Вы зарегистрированы в системе.")
	}

	if session := b.sessionManager.GetSession(UserID); session != nil {
		log.Printf("Found active session for user %d, type: %s", UserID, session.Type)
		b.handleSessionMessage(ctx, message, session)
		return
	}

	if b.handleMenuCallback(ctx, message) {
		return
	}

	b.sendMainMenu(chatID)
}

func (b *Bot) startSession(ctx context.Context, msg *domain.Message, sessionType SessionType) {
	UserID := domain.UserID(msg.UserID)
	chatID := msg.ChatID

	sessionHandler := func(ctx context.Context, input interface{}) (interface{}, error) {
		b.createAndStartSession(ctx, UserID, chatID, msg, sessionType)
		return nil, nil
	}
	wrappedHandler := b.interceptor.Chain(sessionHandler, "Bot.startSession")
	_, err := wrappedHandler(ctx, msg)

	if err != nil {
		log.Printf("Error starting session: %v", err)
	}
}

func (b *Bot) createAndStartSession(ctx context.Context, UserID domain.UserID, chatID int64, msg *domain.Message, sessionType SessionType) {
	session, err := b.sessionManager.StartSession(UserID, chatID, sessionType)
	if err != nil {
		b.sendErrorMessage(ctx, chatID, err, "start_session")
		return
	}

	metrics.IncrementActiveSessions()

	if err = session.Handler.HandleStep(ctx, b, session, msg); err != nil {
		b.sendErrorMessage(ctx, chatID, err, "session_step")
		b.sessionManager.ClearSession(UserID)
	}
}

func (b *Bot) handleSessionMessage(ctx context.Context, msg *domain.Message, session *UserSession) {
	handler := func(ctx context.Context, message interface{}) error {
		b.sessionManager.UpdateLastActivity(session.UserID)
		b.processSessionMessage(ctx, msg, session)
		return nil
	}

	middlewareHandler := func(ctx context.Context, input interface{}) (interface{}, error) {
		return nil, handler(ctx, input)
	}
	wrappedHandler := b.interceptor.Chain(middlewareHandler, "Bot.handleSessionMessage")
	_, err := wrappedHandler(ctx, msg)
	if err != nil {
		log.Printf("Error in session message: %v", err)
	}
}

func (b *Bot) processSessionMessage(ctx context.Context, msg *domain.Message, session *UserSession) {
	if strings.ToLower(msg.Text) == "/cancel" || strings.ToLower(msg.Text) == "отмена" {
		b.cancelSession(session.UserID, session.ChatID)
		return
	}

	if err := session.Handler.HandleStep(ctx, b, session, msg); err != nil {
		b.sendMessage(session.ChatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
	}
}

func (b *Bot) cancelSession(UserID domain.UserID, chatID int64) {
	b.sessionManager.ClearSession(UserID)
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
	log.Printf("Started %s workers", utils.FormatInteger(int64(wp.workers)))
}

func (wp *WorkerPool) Stop() {
	close(wp.jobChannel)
	wp.wg.Wait()
	log.Println("All workers stopped")
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int64) {
	defer wp.wg.Done()

	log.Printf("Worker %s started", utils.FormatInteger(workerID))

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %s stopping due to context cancellation", utils.FormatInteger(workerID))
			return
		case job, ok := <-wp.jobChannel:
			if !ok {
				log.Printf("Worker %s stopping due to closed channel", utils.FormatInteger(workerID))
				return
			}

			job.bot.handleMessage(job.ctx, job.message)
			job.span.End()
		}
	}
}

func (b *Bot) setupBotCommands() error {
	commands := []tgBotAPI.BotCommand{
		{
			Command:     "start",
			Description: "🏠 Начать работу с ботом",
		},
		{
			Command:     "total",
			Description: "💰 Показать общий баланс",
		},
		{
			Command:     "deposits",
			Description: "💳 Показать депозиты",
		},
		{
			Command:     "create_deposit",
			Description: "➕ Добавить депозит",
		},
		{
			Command:     "brokerage_accounts",
			Description: "📈 Показать брокерские счета",
		},
		{
			Command:     "create_brokerage_account",
			Description: "➕ Добавить брокерский счет",
		},
		{
			Command:     "saving_accounts",
			Description: "🏦 Показать накопительные счета",
		},
		{
			Command:     "create_saving_account",
			Description: "➕ Добавить накопительный счет",
		},
		{
			Command:     "cash_holdings",
			Description: "💵 Показать наличные",
		},
		{
			Command:     "create_cash_holding",
			Description: "➕ Добавить наличные",
		},
		{
			Command:     "rates",
			Description: "💱 Курсы валют",
		},
		{
			Command:     "financial_advice",
			Description: "💡 Получить финансовые советы",
		},
	}

	return b.botAPI.SetMyCommands(commands)
}
