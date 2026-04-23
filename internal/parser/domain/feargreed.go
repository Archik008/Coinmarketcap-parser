// internal/domain/feargreed.go
package domain

import (
	"crypto_parser/internal/parser/domain/app_errors"
	"time"
)

type FearGreedLabel string

const (
	ExtremeFear  FearGreedLabel = "Extreme Fear"
	Fear         FearGreedLabel = "Fear"
	Neutral      FearGreedLabel = "Neutral"
	Greed        FearGreedLabel = "Greed"
	ExtremeGreed FearGreedLabel = "Extreme Greed"
)

type FearGreed struct {
	Value      int // 0-100
	Label      FearGreedLabel
	CapturedAt time.Time
}

func NewFearGreed(value int, label string) (FearGreed, error) {
	if value < 0 || value > 100 {
		return FearGreed{}, app_errors.ErrValueOutOfRange
	}
	l := FearGreedLabel(label)
	switch l {
	case ExtremeFear, Fear, Neutral, Greed, ExtremeGreed:
	default:
		return FearGreed{}, app_errors.ErrInvalidLabel
	}
	return FearGreed{
		Value:      value,
		Label:      l,
		CapturedAt: time.Now().UTC(),
	}, nil
}
