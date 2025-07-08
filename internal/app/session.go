package app

import (
	"context"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/metrics"
)

type SessionType string

const (
	SessionCreateDeposit                 SessionType = "create_deposit"
	SessionCreateSavingAccount           SessionType = "create_saving_account"
	SessionCreateBrokerageAccountSession SessionType = "create_brokerage_account_session"
	SessionCreateCashHolding             SessionType = "create_cash_holding"
	SessionFinancialAdvice               SessionType = "financial_advice"
)

type SessionStep string

const (
	StepStart        SessionStep = "start"
	StepName         SessionStep = "name"
	StepAmount       SessionStep = "amount"
	StepCurrency     SessionStep = "currency"
	StepInterestRate SessionStep = "interest_rate"
	StepDate         SessionStep = "date"
	StepCategory     SessionStep = "category"
	StepAccount      SessionStep = "account"
	StepConfirmation SessionStep = "confirmation"
)

type SessionHandler interface {
	GetSessionType() SessionType
	HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *domain.Message) error
	FormatConfirmation(session *UserSession) string
	CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error
}

type UserSession struct {
	UserID       domain.UserID
	ChatID       int64
	Type         SessionType
	CurrentStep  SessionStep
	Data         map[string]interface{}
	Handler      SessionHandler
	LastActivity time.Time
	CreatedAt    time.Time
}

type UserSessionManager struct {
	sessions map[domain.UserID]*UserSession
	handlers map[SessionType]SessionHandler
}

func NewUserSessionManager(financeService FinanceService) *UserSessionManager {
	manager := &UserSessionManager{
		sessions: make(map[domain.UserID]*UserSession),
		handlers: make(map[SessionType]SessionHandler),
	}

	manager.RegisterHandler(&DepositCreationHandler{})
	manager.RegisterHandler(&BrokerageAccountCreationHandler{})
	manager.RegisterHandler(&SavingAccountCreationHandler{})
	manager.RegisterHandler(&CashHoldingCreationHandler{})
	manager.RegisterHandler(&FinancialAdviceHandler{})

	return manager
}

func (usm *UserSessionManager) RegisterHandler(handler SessionHandler) {
	usm.handlers[handler.GetSessionType()] = handler
}

func (usm *UserSessionManager) StartSession(UserID domain.UserID, chatID int64, sessionType SessionType) (*UserSession, error) {
	handler, exists := usm.handlers[sessionType]
	if !exists {
		return nil, errors.NewUnknownSessionHandlerError(string(sessionType))
	}

	session := &UserSession{
		UserID:       UserID,
		ChatID:       chatID,
		Type:         sessionType,
		CurrentStep:  StepStart,
		Data:         make(map[string]interface{}),
		Handler:      handler,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
	}

	usm.sessions[UserID] = session
	return session, nil
}

func (usm *UserSessionManager) GetSession(UserID domain.UserID) *UserSession {
	return usm.sessions[UserID]
}

func (usm *UserSessionManager) ClearSession(UserID domain.UserID) {
	delete(usm.sessions, UserID)
	metrics.DecrementActiveSessions()
}

func (usm *UserSessionManager) UpdateLastActivity(UserID domain.UserID) {
	if session := usm.sessions[UserID]; session != nil {
		session.LastActivity = time.Now()
	}
}

func (s *UserSession) GetData(key string) interface{} {
	return s.Data[key]
}

func (s *UserSession) SetData(key string, value interface{}) {
	s.Data[key] = value
}

func (s *UserSession) GetString(key string) string {
	if value, ok := s.Data[key].(string); ok {
		return value
	}
	return ""
}

func (s *UserSession) GetFloat(key string) float64 {
	if value, ok := s.Data[key].(float64); ok {
		return value
	}
	return 0
}

func (s *UserSession) GetTime(key string) *time.Time {
	if value, ok := s.Data[key].(*time.Time); ok {
		return value
	}
	return nil
}
