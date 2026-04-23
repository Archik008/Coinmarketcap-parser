// internal/domain/cmc20.go
package domain

import (
	"crypto_parser/internal/parser/domain/app_errors"
	"time"
)

type CMC20 struct {
	Value        float64 // суммарная капитализация топ-20 в USD
	Change24hPct float64 // % изменения за 24ч (может быть отрицательным)
	CapturedAt   time.Time
}

func NewCMC20(value, change float64) (CMC20, error) {
	if value <= 0 {
		return CMC20{}, app_errors.ErrNonPositiveValue
	}
	return CMC20{
		Value:        value,
		Change24hPct: change,
		CapturedAt:   time.Now().UTC(),
	}, nil
}
