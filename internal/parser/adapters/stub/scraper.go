// Stub-адаптер для тестов: реализует out.MarketDataScraper.
// Каждое поле — функция; если nil — возвращает нулевое значение без ошибки.
package stub

import (
	"context"
	"crypto_parser/internal/parser/domain"
)

type Scraper struct {
	FetchMarketCapFn     func(ctx context.Context) (domain.MarketCap, error)
	FetchCMC20Fn         func(ctx context.Context) (domain.CMC20, error)
	FetchFearGreedFn     func(ctx context.Context) (domain.FearGreed, error)
	FetchAltcoinSeasonFn func(ctx context.Context) (domain.AltcoinSeason, error)
}

func (s *Scraper) FetchMarketCap(ctx context.Context) (domain.MarketCap, error) {
	if s.FetchMarketCapFn == nil {
		return domain.MarketCap{}, nil
	}
	return s.FetchMarketCapFn(ctx)
}

func (s *Scraper) FetchCMC20(ctx context.Context) (domain.CMC20, error) {
	if s.FetchCMC20Fn == nil {
		return domain.CMC20{}, nil
	}
	return s.FetchCMC20Fn(ctx)
}

func (s *Scraper) FetchFearGreed(ctx context.Context) (domain.FearGreed, error) {
	if s.FetchFearGreedFn == nil {
		return domain.FearGreed{}, nil
	}
	return s.FetchFearGreedFn(ctx)
}

func (s *Scraper) FetchAltcoinSeason(ctx context.Context) (domain.AltcoinSeason, error) {
	if s.FetchAltcoinSeasonFn == nil {
		return domain.AltcoinSeason{}, nil
	}
	return s.FetchAltcoinSeasonFn(ctx)
}
