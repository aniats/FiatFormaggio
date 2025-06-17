package app

import (
	"context"
	"fmt"
	"github.com/aniats/FiatFormaggio/internal/domain"
	"time"
)

type SessionType string

const (
	SessionCreateDeposit                 SessionType = "create_deposit"
	SessionCreateSavingAccount           SessionType = "create_saving_account"
	SessionCreateBrokerageAccountSession SessionType = "create_brokerage_account_session"
	SessionCreateCashHolding             SessionType = "create_cash_holding"
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
	StepComplete     SessionStep = "complete"
)

type SessionHandler interface {
	GetSessionType() SessionType
	HandleStep(ctx context.Context, bot *Bot, session *UserSession, msg *Message) error
	GetNextStep(currentStep SessionStep, input string) (SessionStep, error)
	ValidateInput(step SessionStep, input string) error
	FormatConfirmation(session *UserSession) string
	CompleteSession(ctx context.Context, bot *Bot, session *UserSession) error
}

type UserSession struct {
	UserID       domain.UserId
	ChatID       int64
	Type         SessionType
	CurrentStep  SessionStep
	Data         map[string]interface{}
	Handler      SessionHandler
	LastActivity time.Time
	CreatedAt    time.Time
}

type UserSessionManager struct {
	sessions map[domain.UserId]*UserSession
	handlers map[SessionType]SessionHandler
}

func NewUserSessionManager() *UserSessionManager {
	manager := &UserSessionManager{
		sessions: make(map[domain.UserId]*UserSession),
		handlers: make(map[SessionType]SessionHandler),
	}

	manager.RegisterHandler(&DepositCreationHandler{})
	manager.RegisterHandler(&BrokerageAccountCreationHandler{})
	manager.RegisterHandler(&SavingAccountCreationHandler{})
	manager.RegisterHandler(&CashHoldingCreationHandler{})

	return manager
}

func (usm *UserSessionManager) RegisterHandler(handler SessionHandler) {
	usm.handlers[handler.GetSessionType()] = handler
}

func (usm *UserSessionManager) StartSession(userID domain.UserId, chatID int64, sessionType SessionType) (*UserSession, error) {
	handler, exists := usm.handlers[sessionType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for session type: %s", sessionType)
	}

	session := &UserSession{
		UserID:       userID,
		ChatID:       chatID,
		Type:         sessionType,
		CurrentStep:  StepStart,
		Data:         make(map[string]interface{}),
		Handler:      handler,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
	}

	usm.sessions[userID] = session
	return session, nil
}

func (usm *UserSessionManager) GetSession(userID domain.UserId) *UserSession {
	return usm.sessions[userID]
}

func (usm *UserSessionManager) ClearSession(userID domain.UserId) {
	delete(usm.sessions, userID)
}

func (usm *UserSessionManager) UpdateLastActivity(userID domain.UserId) {
	if session := usm.sessions[userID]; session != nil {
		session.LastActivity = time.Now()
	}
}

func (usm *UserSessionManager) CleanupExpiredSessions(maxInactivity time.Duration) {
	cutoff := time.Now().Add(-maxInactivity)

	for userID, session := range usm.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(usm.sessions, userID)
		}
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
