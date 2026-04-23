// internal/domain/marketcap.go
package domain

import (
	"crypto_parser/internal/parser/domain/app_errors"
	"time"
)

type MarketCap struct {
	Value        float64 // в USD
	Change24hPct float64 // % изменения за 24ч
	CapturedAt   time.Time
}

func NewMarketCap(value, change float64) (MarketCap, error) {
	if value <= 0 {
		return MarketCap{}, app_errors.ErrNonPositiveValue
	}
	return MarketCap{
		Value:        value,
		Change24hPct: change,
		CapturedAt:   time.Now().UTC(),
	}, nil
}
