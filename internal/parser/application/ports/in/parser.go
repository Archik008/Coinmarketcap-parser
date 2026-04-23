// internal/application/ports/in/parser.go
package in

import (
	"context"
	"crypto_parser/internal/parser/domain"
)

// Snapshot — агрегированный результат одного прохода парсинга.
type Snapshot struct {
	MarketCap     domain.MarketCap
	CMC20         domain.CMC20
	FearGreed     domain.FearGreed
	AltcoinSeason domain.AltcoinSeason
}

// ParserUseCase — входящий порт. Каждый метод — отдельный use case,
// ParseAll — агрегатор для удобства клиентов, которым нужен полный снимок.
type ParserUseCase interface {
	ParseMarketCap(ctx context.Context) (domain.MarketCap, error)
	ParseCMC20(ctx context.Context) (domain.CMC20, error)
	ParseFearGreed(ctx context.Context) (domain.FearGreed, error)
	ParseAltcoinSeason(ctx context.Context) (domain.AltcoinSeason, error)
	ParseAll(ctx context.Context) (Snapshot, error)
}
