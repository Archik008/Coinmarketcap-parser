// internal/application/service/parser_service.go
package service

import (
	"context"
	"crypto_parser/internal/parser/application/ports/in"
	"crypto_parser/internal/parser/application/ports/out"
	"crypto_parser/internal/parser/domain"
	"fmt"
)

type ParserService struct {
	scraper out.MarketDataScraper
}

func NewParserService(scraper out.MarketDataScraper) *ParserService {
	return &ParserService{scraper: scraper}
}

func (s *ParserService) ParseMarketCap(ctx context.Context) (domain.MarketCap, error) {
	mc, err := s.scraper.FetchMarketCap(ctx)
	if err != nil {
		return domain.MarketCap{}, fmt.Errorf("parse market cap: %w", err)
	}
	return mc, nil
}

func (s *ParserService) ParseCMC20(ctx context.Context) (domain.CMC20, error) {
	v, err := s.scraper.FetchCMC20(ctx)
	if err != nil {
		return domain.CMC20{}, fmt.Errorf("parse cmc20: %w", err)
	}
	return v, nil
}

func (s *ParserService) ParseFearGreed(ctx context.Context) (domain.FearGreed, error) {
	v, err := s.scraper.FetchFearGreed(ctx)
	if err != nil {
		return domain.FearGreed{}, fmt.Errorf("parse fear&greed: %w", err)
	}
	return v, nil
}

func (s *ParserService) ParseAltcoinSeason(ctx context.Context) (domain.AltcoinSeason, error) {
	v, err := s.scraper.FetchAltcoinSeason(ctx)
	if err != nil {
		return domain.AltcoinSeason{}, fmt.Errorf("parse altcoin season: %w", err)
	}
	return v, nil
}

func (s *ParserService) ParseAll(ctx context.Context) (in.Snapshot, error) {
	var snap in.Snapshot

	mc, err := s.ParseMarketCap(ctx)
	if err != nil {
		return snap, err
	}
	cmc, err := s.ParseCMC20(ctx)
	if err != nil {
		return snap, err
	}
	fg, err := s.ParseFearGreed(ctx)
	if err != nil {
		return snap, err
	}
	alt, err := s.ParseAltcoinSeason(ctx)
	if err != nil {
		return snap, err
	}

	snap.MarketCap = mc
	snap.CMC20 = cmc
	snap.FearGreed = fg
	snap.AltcoinSeason = alt
	return snap, nil
}
