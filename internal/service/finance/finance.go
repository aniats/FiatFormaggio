package finance

import (
	cbr2 "github.com/aniats/FiatFormaggio/internal/service/cbr"

	"github.com/aniats/FiatFormaggio/internal/repository"
)

type FinanceService struct {
	repo repository.Repository
	cbr  *cbr2.CBRService
}

func NewFinanceService(repo repository.Repository, cbr *cbr2.CBRService) *FinanceService {
	return &FinanceService{
		repo: repo,
		cbr:  cbr,
	}
}
