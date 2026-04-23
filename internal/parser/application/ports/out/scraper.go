// internal/application/ports/out/scraper.go
package out

import (
	"context"
	"crypto_parser/internal/parser/domain"
)

// MarketDataScraper — исходящий порт.
// Сервис не знает, HTML это, API или mock.
type MarketDataScraper interface {
	FetchMarketCap(ctx context.Context) (domain.MarketCap, error)
	FetchCMC20(ctx context.Context) (domain.CMC20, error)
	FetchFearGreed(ctx context.Context) (domain.FearGreed, error)
	FetchAltcoinSeason(ctx context.Context) (domain.AltcoinSeason, error)
}
