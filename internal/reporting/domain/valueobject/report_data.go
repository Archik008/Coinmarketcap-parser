package valueobject

import (
	"crypto_parser/internal/parser/domain/dto"
	"time"
)

type ReportData struct {
	MarketCap     dto.MarketCapDTO
	CMC20         dto.CMC20DTO
	FearGreed     dto.FearGreedDTO
	AltcoinSeason dto.AltcoinSeasonDTO
	GeneratedAt   time.Time
}
