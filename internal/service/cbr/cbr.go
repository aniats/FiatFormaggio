package cbr

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"github.com/aniats/FiatFormaggio/internal/errors"
	"github.com/aniats/FiatFormaggio/internal/middleware"
	"golang.org/x/net/html/charset"
)

type CBRService struct {
	client      *http.Client
	interceptor *middleware.Interceptor
}

func NewCBRService(client *http.Client, appName string) *CBRService {
	interceptor := middleware.NewInterceptor(middleware.DefaultConfig("CBRService"), appName)
	return &CBRService{
		client:      client,
		interceptor: interceptor,
	}
}

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

type Valute struct {
	ID        string `xml:"ID,attr"`
	NumCode   string `xml:"NumCode"`
	CharCode  string `xml:"CharCode"`
	Nominal   string `xml:"Nominal"`
	Name      string `xml:"Name"`
	Value     string `xml:"Value"`
	VunitRate string `xml:"VunitRate"`
}

func (s *CBRService) GetCurrencyRates(ctx context.Context, date time.Time) ([]*domain.CurrencyRateCBR, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		resp, err := s.fetchCBRData(ctx, date)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		decoder := xml.NewDecoder(resp.Body)
		decoder.CharsetReader = charset.NewReaderLabel

		var valCurs ValCurs
		if err := decoder.Decode(&valCurs); err != nil {
			return nil, errors.WrapExternalAPIError(err)
		}

		rates := make([]*domain.CurrencyRateCBR, 0, len(valCurs.Valutes))
		for _, valute := range valCurs.Valutes {
			rate, err := s.parseValuteToRate(valute)
			if err != nil {
				continue
			}
			rates = append(rates, rate)
		}

		return rates, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "CBRService.GetCurrencyRates")
	resultInterface, err := wrappedHandler(ctx, date)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.([]*domain.CurrencyRateCBR), nil
	}
	return nil, nil
}

func (s *CBRService) fetchCBRData(ctx context.Context, date time.Time) (*http.Response, error) {
	handler := func(ctx context.Context, input interface{}) (interface{}, error) {
		url := fmt.Sprintf("https://www.cbr.ru/scripts/XML_daily.asp?date_req=%s", date.Format("02/01/2006"))

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, errors.WrapExternalAPIError(err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, errors.WrapExternalAPIError(err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, errors.NewTechnicalError(errors.CodeExternalAPIError, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		}

		return resp, nil
	}

	wrappedHandler := s.interceptor.Chain(handler, "CBRService.fetchCBRData")
	resultInterface, err := wrappedHandler(ctx, date)
	if err != nil {
		return nil, err
	}

	if resultInterface != nil {
		return resultInterface.(*http.Response), nil
	}
	return nil, nil
}

func (s *CBRService) parseValuteToRate(valute Valute) (*domain.CurrencyRateCBR, error) {
	valueStr := strings.Replace(valute.Value, ",", ".", -1)
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, errors.WrapParseError(err).WithContext("value", valute.Value)
	}

	nominal, err := strconv.Atoi(valute.Nominal)
	if err != nil {
		return nil, errors.WrapParseError(err).WithContext("nominal", valute.Nominal)
	}

	return &domain.CurrencyRateCBR{
		CharCode: valute.CharCode,
		Name:     valute.Name,
		Value:    value,
		Nominal:  nominal,
	}, nil
}
