package service

import (
	"context"
	"encoding/xml"

	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aniats/FiatFormaggio/internal/domain"
	"golang.org/x/net/html/charset"
)

type CBRService struct {
	client *http.Client
}

func NewCBRService(client *http.Client) *CBRService {
	return &CBRService{client: client}
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
	url := fmt.Sprintf("https://www.cbr.ru/scripts/XML_daily.asp?date_req=%s", date.Format("02/01/2006"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	var valCurs ValCurs
	if err := decoder.Decode(&valCurs); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}

	var rates []*domain.CurrencyRateCBR
	for _, valute := range valCurs.Valutes {
		valueStr := strings.Replace(valute.Value, ",", ".", -1)
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		nominal, err := strconv.Atoi(valute.Nominal)
		if err != nil {
			continue
		}

		rates = append(rates, &domain.CurrencyRateCBR{
			CharCode: valute.CharCode,
			Name:     valute.Name,
			Value:    value,
			Nominal:  nominal,
		})
	}

	return rates, nil
}
