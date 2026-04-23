//go:build integration

package coinmarketcap_test

// Запуск: go test -tags=integration ./internal/parser/adapters/out/coinmarketcap/...

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"crypto_parser/internal/config"
	"crypto_parser/internal/parser/adapters/out/coinmarketcap"
)

const apiDelay = 1 * time.Second

func newIntegrationScraper(t *testing.T) *coinmarketcap.Scraper {
	t.Helper()
	if err := godotenv.Load("testdata/.env"); err != nil {
		t.Fatalf("load testdata/.env: %v", err)
	}
	// Задержка после каждого теста — не спамим API.
	t.Cleanup(func() { time.Sleep(apiDelay) })
	return coinmarketcap.NewScraper(config.CoinMarketCfg{
		Url:      os.Getenv("COINMARKETCAP_URL"),
		ApiToken: os.Getenv("COINMARKETCAP_API_KEY"),
	})
}

func TestIntegration_FetchMarketCap(t *testing.T) {
	s := newIntegrationScraper(t)

	mc, err := s.FetchMarketCap(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.Value <= 0 {
		t.Errorf("expected positive market cap, got %v", mc.Value)
	}
	if mc.CapturedAt.IsZero() {
		t.Error("CapturedAt must not be zero")
	}
}

func TestIntegration_FetchFearGreed(t *testing.T) {
	s := newIntegrationScraper(t)

	fg, err := s.FetchFearGreed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fg.Value < 0 || fg.Value > 100 {
		t.Errorf("Value out of range [0,100]: got %v", fg.Value)
	}
}

func TestIntegration_FetchCMC20(t *testing.T) {
	s := newIntegrationScraper(t)
	cmc, err := s.FetchCMC20(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmc.Value <= 0 {
		t.Errorf("expected positive value, got %v", cmc.Value)
	}
}

func TestIntegration_FetchAltcoinSeason(t *testing.T) {
	s := newIntegrationScraper(t)

	alt, err := s.FetchAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alt.Index < 0 || alt.Index > 100 {
		t.Errorf("Index out of range [0,100]: got %v", alt.Index)
	}
	if alt.Total <= 0 {
		t.Errorf("Total must be positive, got %v", alt.Total)
	}
	if alt.Outperformed < 0 || alt.Outperformed > alt.Total {
		t.Errorf("Outperformed=%v out of [0, Total=%v]", alt.Outperformed, alt.Total)
	}
	// проверяем что IsAltSeason соответствует доменному правилу
	wantAlt := alt.Index >= 75
	if alt.IsAltSeason != wantAlt {
		t.Errorf("IsAltSeason: got %v, want %v (index=%.1f)", alt.IsAltSeason, wantAlt, alt.Index)
	}
}
