package dto

import "time"

type AltcoinSeasonDTO struct {
	Index        float64 // % альткоинов, обогнавших BTC за 7д (0-100)
	Total        int     // сколько монет участвовало в расчёте
	Outperformed int     // сколько из них обогнали BTC
	IsAltSeason  bool    // true если Index >= 75
	CapturedAt   time.Time
}

type CMC20DTO struct {
	Value        float64 // суммарная капитализация топ-20 в USD
	Change24hPct float64 // % изменения за 24ч (может быть отрицательным)
	CapturedAt   time.Time
}

type FearGreedDTO struct {
	Value      int // 0-100
	Label      string
	CapturedAt time.Time
}

type MarketCapDTO struct {
	Value        float64 // в USD
	Change24hPct float64 // % изменения за 24ч
	CapturedAt   time.Time
}
