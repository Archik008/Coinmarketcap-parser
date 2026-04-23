package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"crypto_parser/internal/parser/adapters/stub"
	"crypto_parser/internal/parser/application"
	"crypto_parser/internal/parser/domain"
)

var (
	validMC, _  = domain.NewMarketCap(1_000_000_000, 2.5)
	validCMC, _ = domain.NewCMC20(500_000_000, -1.2)
	validFG, _  = domain.NewFearGreed(45, string(domain.Fear))
	validAlt, _ = domain.NewAltcoinSeason(30.0, 100, 30)

	errScraper = errors.New("scraper unavailable")
)

// --- ParseMarketCap ---

func TestParseMarketCap_Success(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchMarketCapFn: func(_ context.Context) (domain.MarketCap, error) {
			return validMC, nil
		},
	})

	got, err := svc.ParseMarketCap(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != validMC.Value {
		t.Errorf("Value: got %v, want %v", got.Value, validMC.Value)
	}
}

func TestParseMarketCap_ScraperError(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchMarketCapFn: func(_ context.Context) (domain.MarketCap, error) {
			return domain.MarketCap{}, errScraper
		},
	})

	_, err := svc.ParseMarketCap(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("error chain broken: got %v", err)
	}
	if !strings.Contains(err.Error(), "parse market cap") {
		t.Errorf("missing context in error: %v", err)
	}
}

// --- ParseCMC20 ---

func TestParseCMC20_Success(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchCMC20Fn: func(_ context.Context) (domain.CMC20, error) {
			return validCMC, nil
		},
	})

	got, err := svc.ParseCMC20(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != validCMC.Value {
		t.Errorf("Value: got %v, want %v", got.Value, validCMC.Value)
	}
}

func TestParseCMC20_ScraperError(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchCMC20Fn: func(_ context.Context) (domain.CMC20, error) {
			return domain.CMC20{}, errScraper
		},
	})

	_, err := svc.ParseCMC20(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("error chain broken: got %v", err)
	}
	if !strings.Contains(err.Error(), "parse cmc20") {
		t.Errorf("missing context in error: %v", err)
	}
}

// --- ParseFearGreed ---

func TestParseFearGreed_Success(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchFearGreedFn: func(_ context.Context) (domain.FearGreed, error) {
			return validFG, nil
		},
	})

	got, err := svc.ParseFearGreed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Value != validFG.Value {
		t.Errorf("Value: got %v, want %v", got.Value, validFG.Value)
	}
}

func TestParseFearGreed_ScraperError(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchFearGreedFn: func(_ context.Context) (domain.FearGreed, error) {
			return domain.FearGreed{}, errScraper
		},
	})

	_, err := svc.ParseFearGreed(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("error chain broken: got %v", err)
	}
	if !strings.Contains(err.Error(), "parse fear") {
		t.Errorf("missing context in error: %v", err)
	}
}

// --- ParseAltcoinSeason ---

func TestParseAltcoinSeason_Success(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchAltcoinSeasonFn: func(_ context.Context) (domain.AltcoinSeason, error) {
			return validAlt, nil
		},
	})

	got, err := svc.ParseAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Index != validAlt.Index {
		t.Errorf("Index: got %v, want %v", got.Index, validAlt.Index)
	}
}

func TestParseAltcoinSeason_ScraperError(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchAltcoinSeasonFn: func(_ context.Context) (domain.AltcoinSeason, error) {
			return domain.AltcoinSeason{}, errScraper
		},
	})

	_, err := svc.ParseAltcoinSeason(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("error chain broken: got %v", err)
	}
	if !strings.Contains(err.Error(), "parse altcoin season") {
		t.Errorf("missing context in error: %v", err)
	}
}

// --- ParseAll ---

func TestParseAll_Success(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchMarketCapFn:     func(_ context.Context) (domain.MarketCap, error) { return validMC, nil },
		FetchCMC20Fn:         func(_ context.Context) (domain.CMC20, error) { return validCMC, nil },
		FetchFearGreedFn:     func(_ context.Context) (domain.FearGreed, error) { return validFG, nil },
		FetchAltcoinSeasonFn: func(_ context.Context) (domain.AltcoinSeason, error) { return validAlt, nil },
	})

	snap, err := svc.ParseAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.MarketCap.Value != validMC.Value {
		t.Errorf("Snapshot.MarketCap: got %v, want %v", snap.MarketCap.Value, validMC.Value)
	}
	if snap.CMC20.Value != validCMC.Value {
		t.Errorf("Snapshot.CMC20: got %v, want %v", snap.CMC20.Value, validCMC.Value)
	}
	if snap.FearGreed.Value != validFG.Value {
		t.Errorf("Snapshot.FearGreed: got %v, want %v", snap.FearGreed.Value, validFG.Value)
	}
	if snap.AltcoinSeason.Index != validAlt.Index {
		t.Errorf("Snapshot.AltcoinSeason: got %v, want %v", snap.AltcoinSeason.Index, validAlt.Index)
	}
}

func TestParseAll_FirstStepFails(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchMarketCapFn: func(_ context.Context) (domain.MarketCap, error) {
			return domain.MarketCap{}, errScraper
		},
	})

	_, err := svc.ParseAll(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("expected errScraper in chain, got: %v", err)
	}
}

func TestParseAll_MiddleStepFails(t *testing.T) {
	svc := service.NewParserService(&stub.Scraper{
		FetchMarketCapFn: func(_ context.Context) (domain.MarketCap, error) { return validMC, nil },
		FetchCMC20Fn:     func(_ context.Context) (domain.CMC20, error) { return validCMC, nil },
		FetchFearGreedFn: func(_ context.Context) (domain.FearGreed, error) {
			return domain.FearGreed{}, errScraper
		},
	})

	_, err := svc.ParseAll(context.Background())
	if !errors.Is(err, errScraper) {
		t.Fatalf("expected errScraper in chain, got: %v", err)
	}
}
