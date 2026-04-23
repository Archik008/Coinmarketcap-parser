package coinmarketcap

import (
	"context"
	"crypto_parser/internal/config/domain/valueobject"
	"crypto_parser/internal/parser/domain"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Scraper struct {
	cfg     valueobject.CoinMarketCapCfg
	queueCh chan struct{}
	client  *http.Client
	wg      *sync.WaitGroup
}

func NewScraper(cfg valueobject.CoinMarketCapCfg, ctx context.Context, wg *sync.WaitGroup) *Scraper {
	scraper := &Scraper{
		client:  &http.Client{Timeout: 15 * time.Second},
		cfg:     cfg,
		queueCh: make(chan struct{}, 1),
	}

	scraper.startQueueListener(wg, ctx)

	return scraper
}

// params — вспомогательный метод: собирает fetchParams из полей Scraper.
// Так каждый FetchXxx передаёт только path — всё остальное уже здесь.
func (s *Scraper) params(path string) fetchParams {
	return fetchParams{
		path:     path,
		baseURL:  s.cfg.Url,
		apiToken: s.cfg.ApiToken,
		client:   s.client,
	}
}

func (s *Scraper) startQueueListener(wg *sync.WaitGroup, ctx context.Context) {
	wg.Go(func() {
		for {
			// сначала ждём 2s — буфер остаётся занятым, второй запрос блокируется
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				slog.Info("Queue listener stopped")
				return
			}
			// дренируем канал
			select {
			case <-s.queueCh:
			case <-ctx.Done():
				slog.Info("Queue listener stopped at 2 select")
				return
			}
		}
	})
}

func (s *Scraper) FetchMarketCap(ctx context.Context) (domain.MarketCap, error) {
	raw, err := fetch[marketCapDTO](ctx, s.params("/v1/global-metrics/quotes/latest"), s.queueCh)
	if err != nil {
		return domain.MarketCap{}, err
	}

	return domain.NewMarketCap(
		raw.Data.Quote.USD.TotalMarketCap,
		raw.Data.Quote.USD.Change24hPct,
	)
}

func (s *Scraper) FetchCMC20(ctx context.Context) (domain.CMC20, error) {
	raw, err := fetch[cmc20HistoricalResponse](ctx, s.params("/v3/index/cmc20-historical"), s.queueCh)
	if err != nil {
		return domain.CMC20{}, err
	}
	if len(raw.Data) < 2 {
		return domain.CMC20{}, fmt.Errorf("cmc20: need at least 2 data points, got %d", len(raw.Data))
	}
	last := raw.Data[len(raw.Data)-1]
	prev := raw.Data[len(raw.Data)-2]
	change := (last.Value - prev.Value) / prev.Value * 100
	return domain.NewCMC20(last.Value, change)
}

func (s *Scraper) FetchFearGreed(ctx context.Context) (domain.FearGreed, error) {
	raw, err := fetch[fearGreedResponse](ctx, s.params("/v3/fear-and-greed/historical"), s.queueCh)
	if err != nil {
		return domain.FearGreed{}, err
	}
	if len(raw.Data) == 0 {
		return domain.FearGreed{}, fmt.Errorf("fear & greed: empty response")
	}
	latest := raw.Data[0]
	return domain.NewFearGreed(latest.Value, normalizeFGLabel(latest.ValueClassification))
}

// normalizeFGLabel переводит API-кейсинг ("Extreme fear") в доменные константы ("Extreme Fear").
var fgLabelMap = map[string]string{
	"Extreme fear":  string(domain.ExtremeFear),
	"Extreme greed": string(domain.ExtremeGreed),
	// на случай если API исправит регистр:
	"Extreme Fear":  string(domain.ExtremeFear),
	"Extreme Greed": string(domain.ExtremeGreed),
}

func normalizeFGLabel(apiLabel string) string {
	if normalized, ok := fgLabelMap[apiLabel]; ok {
		return normalized
	}
	return apiLabel // Fear / Neutral / Greed совпадают, домен проверит остальное
}

func (s *Scraper) FetchAltcoinSeason(ctx context.Context) (domain.AltcoinSeason, error) {
	raw, err := fetchCryptoCurrency(ctx, s.params("/v1/cryptocurrency/listings/latest"), s.queueCh)
	if err != nil {
		return domain.AltcoinSeason{}, err
	}

	// --- конфиг ---
	const minVolume = 1_000_000.0

	stablecoins := map[string]bool{
		"USDT": true,
		"USDC": true,
		"BUSD": true,
		"DAI":  true,
		"TUSD": true,
	}

	var (
		btc7d    float64
		btcFound bool
	)

	// --- 1. найти BTC ---
	for _, c := range raw.Data {
		if c.Symbol == "BTC" {
			btc7d = c.Quote.USD.PercentChange7d
			btcFound = true
			break
		}
	}

	if !btcFound {
		return domain.AltcoinSeason{}, fmt.Errorf("BTC not found in response")
	}

	// --- 2. фильтрация и сравнение ---
	total := 0
	outperformed := 0

	for _, c := range raw.Data {

		// skip BTC
		if c.Symbol == "BTC" {
			continue
		}

		// skip стейблы
		if stablecoins[c.Symbol] {
			continue
		}

		// skip low volume
		if c.Quote.USD.Volume24h < minVolume {
			continue
		}

		total++

		// сравнение с BTC: считаем монеты, которые BTC обогнал
		if btc7d > c.Quote.USD.PercentChange7d {
			outperformed++
		}
	}

	if total == 0 {
		return domain.AltcoinSeason{}, fmt.Errorf("no coins after filtering")
	}

	// --- 3. расчёт индекса ---
	// Адаптер считает только математику из сырых данных.
	// Интерпретация (>= 75 = alt season) — в domain.NewAltcoinSeason.
	index := (float64(outperformed) / float64(total)) * 100

	return domain.NewAltcoinSeason(index, total, outperformed)
}
