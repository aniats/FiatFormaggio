package currency

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"github.com/aniats/FiatFormaggio/internal/repository"
)

const (
	// CacheUpdateInterval - how often to update currency rates
	CacheUpdateInterval = 24 * time.Hour
	// MinorUnitsMultiplier - conversion factor for minor units (kopecks to rubles)
	MinorUnitsMultiplier = 100
)

type ExternalCurrencyService interface {
	GetCurrencyRates(ctx context.Context, date time.Time) ([]*domain.CurrencyRateCBR, error)
}

type CachedCurrencyService struct {
	repo            repository.CurrencyRateRepository
	externalService ExternalCurrencyService
	mu              sync.RWMutex
	stopChan        chan struct{}
	updateTicker    *time.Ticker
	interceptor     *middleware.UnifiedInterceptor
}

func NewCachedCurrencyService(
	repo repository.CurrencyRateRepository,
	externalService ExternalCurrencyService,
) *CachedCurrencyService {
	return &CachedCurrencyService{
		repo:            repo,
		externalService: externalService,
		stopChan:        make(chan struct{}),
		interceptor:     middleware.NewUnifiedInterceptor(middleware.DefaultConfig("CurrencyService")),
	}
}

func (s *CachedCurrencyService) Start(ctx context.Context) error {
	if err := s.updateIfNeeded(ctx); err != nil {
		log.Printf("Initial currency rates update failed: %v", err)
	}

	s.updateTicker = time.NewTicker(CacheUpdateInterval)
	go s.periodicUpdate(ctx)

	log.Println("Currency rate caching service started")
	return nil
}

func (s *CachedCurrencyService) Stop() {
	if s.updateTicker != nil {
		s.updateTicker.Stop()
	}
	close(s.stopChan)
	log.Println("Currency rate caching service stopped")
}

func (s *CachedCurrencyService) GetCurrencyRates(ctx context.Context) ([]domain.CurrencyRate, error) {
	var result []domain.CurrencyRate
	var err error
	
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		rates, repoErr := s.repo.GetCurrencyRates(ctx)
		if repoErr != nil {
			return nil, fmt.Errorf("failed to get cached currency rates: %w", repoErr)
		}

		if len(rates) == 0 {
			s.mu.RUnlock()
			if updateErr := s.updateCurrencyRates(ctx); updateErr != nil {
				s.mu.RLock()
				return nil, fmt.Errorf("no cached rates and update failed: %w", updateErr)
			}
			s.mu.RLock()

			rates, repoErr = s.repo.GetCurrencyRates(ctx)
			if repoErr != nil {
				return nil, fmt.Errorf("failed to get currency rates after update: %w", repoErr)
			}
		}

		return rates, nil
	}
	
	wrappedHandler := s.interceptor.Chain(handler, "CachedCurrencyService.GetCurrencyRates")
	resultInterface, err := wrappedHandler(ctx, nil)
	if err != nil {
		return nil, err
	}
	
	if resultInterface != nil {
		result = resultInterface.([]domain.CurrencyRate)
	}
	
	return result, err
}

func (s *CachedCurrencyService) ForceUpdate(ctx context.Context) error {
	return s.updateCurrencyRates(ctx)
}

func (s *CachedCurrencyService) periodicUpdate(ctx context.Context) {
	for {
		select {
		case <-s.stopChan:
			return
		case <-s.updateTicker.C:
			if err := s.updateIfNeeded(ctx); err != nil {
				log.Printf("Periodic currency rates update failed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *CachedCurrencyService) updateIfNeeded(ctx context.Context) error {
	lastUpdate, err := s.repo.GetLastUpdateTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to check last update time: %w", err)
	}

	if lastUpdate == nil || time.Since(*lastUpdate) > CacheUpdateInterval {
		log.Printf("Currency rates cache is outdated, updating...")
		return s.updateCurrencyRates(ctx)
	}

	log.Printf("Currency rates cache is up to date (last update: %v)", lastUpdate.Format(time.RFC3339))
	return nil
}

func (s *CachedCurrencyService) updateCurrencyRates(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Fetching fresh currency rates from external service...")

	externalRates, err := s.externalService.GetCurrencyRates(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to fetch currency rates from external service: %w", err)
	}

	now := time.Now()
	for _, externalRate := range externalRates {
		currency, err := domain.CurrencyFromHuman(externalRate.CharCode)
		if err != nil {
			log.Printf("Skipping unknown currency %s: %v", externalRate.CharCode, err)
			continue
		}

		rateMinorUnits := int64(externalRate.Value / float64(externalRate.Nominal) * MinorUnitsMultiplier)

		rate := &domain.CurrencyRate{
			Currency:       currency,
			RateMinorUnits: rateMinorUnits,
			BaseCurrency:   domain.RUB,
			Source:         "CBR",
			UpdatedAt:      now,
		}

		if err := s.repo.UpsertCurrencyRate(ctx, rate); err != nil {
			log.Printf("Failed to store currency rate for %s: %v", currency, err)
			continue
		}
	}

	log.Printf("Successfully updated %d currency rates", len(externalRates))
	return nil
}
