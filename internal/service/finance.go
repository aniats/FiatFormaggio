package service

import (
	"context"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

type FinanceService struct {
	repo repository.Repository
	cbr  *CBRService
}

func NewFinanceService(repo repository.Repository, cbr *CBRService) *FinanceService {
	return &FinanceService{
		repo: repo,
		cbr:  cbr,
	}
}

func (s *FinanceService) GetDepositsByUserID(ctx context.Context, userId domain.UserId) ([]domain.Deposit, error) {
	result, err := s.repo.GetDepositsByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return result, nil
}
