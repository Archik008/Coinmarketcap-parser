// internal/domain/altseason.go
package domain

import (
	"crypto_parser/internal/parser/domain/app_errors"
	"time"
)

type AltcoinSeason struct {
	Index        float64 // % монет, которые BTC обогнал за 7д (0-100); низкий = больше альткоинов бьют BTC
	Total        int     // сколько монет участвовало в расчёте
	Outperformed int     // сколько из них BTC обогнал
	IsAltSeason  bool    // true если Index <= 25 (BTC обогнал менее 25% → 75%+ бьют BTC)
	CapturedAt   time.Time
}

// NewAltcoinSeason вычисляет IsAltSeason сам — это бизнес-правило домена,
// не адаптера.
func NewAltcoinSeason(index float64, total, outperformed int) (AltcoinSeason, error) {
	if index < 0 || index > 100 {
		return AltcoinSeason{}, app_errors.ErrValueOutOfRange
	}
	if total <= 0 {
		return AltcoinSeason{}, app_errors.ErrNonPositiveValue
	}
	if outperformed < 0 || outperformed > total {
		return AltcoinSeason{}, app_errors.ErrValueOutOfRange
	}
	return AltcoinSeason{
		Index:        index,
		Total:        total,
		Outperformed: outperformed,
		IsAltSeason:  index <= 25,
		CapturedAt:   time.Now().UTC(),
	}, nil
}
