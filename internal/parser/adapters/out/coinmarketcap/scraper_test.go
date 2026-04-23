package coinmarketcap_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"crypto_parser/internal/config/domain/valueobject"
	"crypto_parser/internal/parser/adapters/out/coinmarketcap"
	"crypto_parser/internal/parser/domain/app_errors"
)

// newScraper поднимает httptest-сервер и возвращает Scraper, указывающий на него.
// Сервер закрывается автоматически после теста.
func newScraper(t *testing.T, handler http.HandlerFunc) *coinmarketcap.Scraper {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	var wg sync.WaitGroup

	scraper := coinmarketcap.NewScraper(valueobject.CoinMarketCapCfg{
		Url:      srv.URL,
		ApiToken: "test-key",
	}, context.Background(), &wg)
	return scraper
}

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

// ─── FetchMarketCap ────────────────────────────────────────────────────────────

func TestFetchMarketCap_ParsesResponse(t *testing.T) {
	s := newScraper(t, jsonHandler(`{
		"data": {
			"quote": {
				"USD": {
					"total_market_cap": 2500000000000,
					"total_market_cap_yesterday_percentage_change": 1.5
				}
			}
		}
	}`))

	mc, err := s.FetchMarketCap(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.Value != 2_500_000_000_000 {
		t.Errorf("Value: got %v, want 2500000000000", mc.Value)
	}
	if mc.Change24hPct != 1.5 {
		t.Errorf("Change24hPct: got %v, want 1.5", mc.Change24hPct)
	}
	if mc.CapturedAt.IsZero() {
		t.Error("CapturedAt must not be zero")
	}
}

func TestFetchMarketCap_NegativeValueFromAPI(t *testing.T) {
	s := newScraper(t, jsonHandler(`{
		"data": {
			"quote": {
				"USD": {
					"total_market_cap": -1,
					"total_market_cap_yesterday_percentage_change": 0
				}
			}
		}
	}`))

	_, err := s.FetchMarketCap(context.Background())
	if !errors.Is(err, app_errors.ErrNonPositiveValue) {
		t.Fatalf("expected ErrNonPositiveValue, got: %v", err)
	}
}

func TestFetchMarketCap_MalformedJSON(t *testing.T) {
	s := newScraper(t, jsonHandler(`not json`))

	_, err := s.FetchMarketCap(context.Background())
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}

func TestFetchMarketCap_ContextCancelled(t *testing.T) {
	s := newScraper(t, jsonHandler(`{}`))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.FetchMarketCap(ctx)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

// ─── FetchFearGreed ────────────────────────────────────────────────────────────

func TestFetchFearGreed_ParsesResponse(t *testing.T) {
	s := newScraper(t, jsonHandler(`{"data":[{"value":72,"value_classification":"Greed"}]}`))

	fg, err := s.FetchFearGreed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fg.Value != 72 {
		t.Errorf("Value: got %v, want 72", fg.Value)
	}
	if string(fg.Label) != "Greed" {
		t.Errorf("Label: got %v, want Greed", fg.Label)
	}
}

func TestFetchFearGreed_NormalizesExtremeFear(t *testing.T) {
	// API шлёт "Extreme fear" (строчная f) — адаптер должен нормализовать в доменную константу
	s := newScraper(t, jsonHandler(`{"data":[{"value":15,"value_classification":"Extreme fear"}]}`))

	fg, err := s.FetchFearGreed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(fg.Label) != "Extreme Fear" {
		t.Errorf("Label: got %q, want %q", fg.Label, "Extreme Fear")
	}
}

func TestFetchFearGreed_InvalidLabelFromAPI(t *testing.T) {
	s := newScraper(t, jsonHandler(`{"data":[{"value":50,"value_classification":"Unknown Mood"}]}`))

	_, err := s.FetchFearGreed(context.Background())
	if !errors.Is(err, app_errors.ErrInvalidLabel) {
		t.Fatalf("expected ErrInvalidLabel, got: %v", err)
	}
}

func TestFetchFearGreed_OutOfRangeValue(t *testing.T) {
	s := newScraper(t, jsonHandler(`{"data":[{"value":150,"value_classification":"Extreme Greed"}]}`))

	_, err := s.FetchFearGreed(context.Background())
	if !errors.Is(err, app_errors.ErrValueOutOfRange) {
		t.Fatalf("expected ErrValueOutOfRange, got: %v", err)
	}
}

func TestFetchFearGreed_EmptyData(t *testing.T) {
	s := newScraper(t, jsonHandler(`{"data":[]}`))

	_, err := s.FetchFearGreed(context.Background())
	if err == nil {
		t.Fatal("expected error on empty data array")
	}
}

// ─── FetchCMC20 ────────────────────────────────────────────────────────────────

func TestFetchCMC20_ParsesResponse(t *testing.T) {
	// prev=150, current=160 → change = +6.666...%
	s := newScraper(t, jsonHandler(`{"data":[{"value":150},{"value":160}]}`))

	cmc, err := s.FetchCMC20(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmc.Value != 160 {
		t.Errorf("Value: got %v, want 160", cmc.Value)
	}
	want := (160.0 - 150.0) / 150.0 * 100
	if cmc.Change24hPct != want {
		t.Errorf("Change24hPct: got %v, want %v", cmc.Change24hPct, want)
	}
}

func TestFetchCMC20_NegativeChangeIsValid(t *testing.T) {
	// рынок упал: prev=160, current=150 → change отрицательный, ошибки не должно быть
	s := newScraper(t, jsonHandler(`{"data":[{"value":160},{"value":150}]}`))

	cmc, err := s.FetchCMC20(context.Background())
	if err != nil {
		t.Fatalf("negative change must not cause error, got: %v", err)
	}
	if cmc.Change24hPct >= 0 {
		t.Errorf("expected negative change, got %v", cmc.Change24hPct)
	}
}

func TestFetchCMC20_TooFewDataPoints(t *testing.T) {
	s := newScraper(t, jsonHandler(`{"data":[{"value":150}]}`))

	_, err := s.FetchCMC20(context.Background())
	if err == nil {
		t.Fatal("expected error when fewer than 2 data points")
	}
}

// ─── FetchAltcoinSeason ────────────────────────────────────────────────────────

// altListingJSON строит JSON для /cryptocurrency/listings/latest.
// BTC и монеты с нужными параметрами.
const (
	altSeasonJSON = `{"data":[
		{"symbol":"BTC","cmc_rank":1,"quote":{"USD":{"percent_change_7d":5.0,"volume_24h":50000000000,"market_cap":1000000000000}}},
		{"symbol":"ETH","cmc_rank":2,"quote":{"USD":{"percent_change_7d":10.0,"volume_24h":5000000000,"market_cap":500000000000}}},
		{"symbol":"SOL","cmc_rank":5,"quote":{"USD":{"percent_change_7d":8.0,"volume_24h":2000000000,"market_cap":100000000000}}}
	]}`
	// BTC 7d=5%; ETH=10%, SOL=8% — оба обгоняют → index=100%, IsAltSeason=true

	bitcoinSeasonJSON = `{"data":[
		{"symbol":"BTC","cmc_rank":1,"quote":{"USD":{"percent_change_7d":20.0,"volume_24h":50000000000,"market_cap":1000000000000}}},
		{"symbol":"ETH","cmc_rank":2,"quote":{"USD":{"percent_change_7d":5.0,"volume_24h":5000000000,"market_cap":500000000000}}},
		{"symbol":"SOL","cmc_rank":5,"quote":{"USD":{"percent_change_7d":10.0,"volume_24h":2000000000,"market_cap":100000000000}}}
	]}`
	// BTC 7d=20%; ETH=5%, SOL=10% — оба уступают → index=0%, IsAltSeason=false

	stablecoinJSON = `{"data":[
		{"symbol":"BTC","cmc_rank":1,"quote":{"USD":{"percent_change_7d":5.0,"volume_24h":50000000000,"market_cap":1000000000000}}},
		{"symbol":"USDT","cmc_rank":3,"quote":{"USD":{"percent_change_7d":20.0,"volume_24h":100000000000,"market_cap":80000000000}}},
		{"symbol":"ETH","cmc_rank":2,"quote":{"USD":{"percent_change_7d":10.0,"volume_24h":5000000000,"market_cap":500000000000}}}
	]}`
	// USDT — стейбл, исключается; ETH обгоняет BTC → total=1, index=100%

	lowVolumeJSON = `{"data":[
		{"symbol":"BTC","cmc_rank":1,"quote":{"USD":{"percent_change_7d":5.0,"volume_24h":50000000000,"market_cap":1000000000000}}},
		{"symbol":"ETH","cmc_rank":2,"quote":{"USD":{"percent_change_7d":10.0,"volume_24h":500000,"market_cap":500000000000}}},
		{"symbol":"SOL","cmc_rank":5,"quote":{"USD":{"percent_change_7d":3.0,"volume_24h":2000000,"market_cap":100000000000}}}
	]}`
	// ETH volume=500k < 1M → исключается; SOL(3%) < BTC(5%) → total=1, outperformed=0, index=0%

	noBTCJSON = `{"data":[
		{"symbol":"ETH","cmc_rank":2,"quote":{"USD":{"percent_change_7d":10.0,"volume_24h":5000000000,"market_cap":500000000000}}}
	]}`

	onlyBTCJSON = `{"data":[
		{"symbol":"BTC","cmc_rank":1,"quote":{"USD":{"percent_change_7d":5.0,"volume_24h":50000000000,"market_cap":1000000000000}}}
	]}`
)

func TestFetchAltcoinSeason_AltSeason(t *testing.T) {
	s := newScraper(t, jsonHandler(altSeasonJSON))

	alt, err := s.FetchAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// BTC(5%) не обогнал ни ETH(10%), ни SOL(8%) → outperformed=0, index=0% → alt season
	if alt.Total != 2 {
		t.Errorf("Total: got %v, want 2", alt.Total)
	}
	if alt.Outperformed != 0 {
		t.Errorf("Outperformed: got %v, want 0", alt.Outperformed)
	}
	if alt.Index != 0 {
		t.Errorf("Index: got %v, want 0", alt.Index)
	}
	if !alt.IsAltSeason {
		t.Error("IsAltSeason: want true")
	}
}

func TestFetchAltcoinSeason_BitcoinSeason(t *testing.T) {
	s := newScraper(t, jsonHandler(bitcoinSeasonJSON))

	alt, err := s.FetchAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// BTC(20%) обогнал и ETH(5%), и SOL(10%) → outperformed=2, index=100% → bitcoin season
	if alt.Outperformed != 2 {
		t.Errorf("Outperformed: got %v, want 2", alt.Outperformed)
	}
	if alt.IsAltSeason {
		t.Error("IsAltSeason: want false")
	}
}

func TestFetchAltcoinSeason_StablecoinExcluded(t *testing.T) {
	s := newScraper(t, jsonHandler(stablecoinJSON))

	alt, err := s.FetchAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// USDT не должен попасть в total
	if alt.Total != 1 {
		t.Errorf("Total: got %v, want 1 (stablecoin must be excluded)", alt.Total)
	}
}

func TestFetchAltcoinSeason_LowVolumeExcluded(t *testing.T) {
	s := newScraper(t, jsonHandler(lowVolumeJSON))

	alt, err := s.FetchAltcoinSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ETH с volume=500k исключается; остаётся только SOL
	// ETH исключён (volume<1M); SOL(3%) < BTC(5%) → BTC обогнал SOL → outperformed=1
	if alt.Total != 1 {
		t.Errorf("Total: got %v, want 1 (low-volume coin must be excluded)", alt.Total)
	}
	if alt.Outperformed != 1 {
		t.Errorf("Outperformed: got %v, want 1", alt.Outperformed)
	}
}

func TestFetchAltcoinSeason_BTCNotFound(t *testing.T) {
	s := newScraper(t, jsonHandler(noBTCJSON))

	_, err := s.FetchAltcoinSeason(context.Background())
	if err == nil {
		t.Fatal("expected error when BTC is absent")
	}
}

func TestFetchAltcoinSeason_NoCoinsAfterFilter(t *testing.T) {
	s := newScraper(t, jsonHandler(onlyBTCJSON))

	_, err := s.FetchAltcoinSeason(context.Background())
	if err == nil {
		t.Fatal("expected error when no coins remain after filtering")
	}
}

// ─── Queue / rate-limiting ─────────────────────────────────────────────────────

const marketCapJSON = `{"data":{"quote":{"USD":{"total_market_cap":1000000000000,"total_market_cap_yesterday_percentage_change":1.5}}}}`

// TestQueue_FirstRequestIsImmediate проверяет, что первый запрос не блокируется
// очередью: буфер канала = 1, поэтому запись в него происходит без ожидания.
func TestQueue_FirstRequestIsImmediate(t *testing.T) {
	s := newScraper(t, jsonHandler(marketCapJSON))

	start := time.Now()
	if _, err := s.FetchMarketCap(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 2*time.Second {
		t.Errorf("first request must not be delayed by queue, took %v", elapsed)
	}
}

// TestQueue_ConsecutiveRequestsThrottled проверяет, что второй последовательный
// запрос задерживается минимум на 2 секунды: горутина дренирует канал раз в 2s,
// поэтому второй send блокируется до следующего тика.
func TestQueue_ConsecutiveRequestsThrottled(t *testing.T) {
	s := newScraper(t, jsonHandler(marketCapJSON))

	start := time.Now()

	if _, err := s.FetchMarketCap(context.Background()); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if _, err := s.FetchMarketCap(context.Background()); err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	if elapsed := time.Since(start); elapsed < 2*time.Second {
		t.Errorf("two consecutive requests must take >= 2s due to queue, took %v", elapsed)
	}
}

// TestQueue_DifferentEndpointsShareQueue проверяет, что разные Fetch-методы
// делят один и тот же канал-очередь: два разных вызова тоже дают задержку >= 2s.
func TestQueue_DifferentEndpointsShareQueue(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/global-metrics/quotes/latest":
			_, _ = w.Write([]byte(marketCapJSON))
		case r.URL.Path == "/v3/fear-and-greed/historical":
			_, _ = w.Write([]byte(`{"data":[{"value":72,"value_classification":"Greed"}]}`))
		}
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	var wg sync.WaitGroup

	s := coinmarketcap.NewScraper(valueobject.CoinMarketCapCfg{Url: srv.URL, ApiToken: "test-key"}, context.Background(), &wg)

	start := time.Now()

	if _, err := s.FetchMarketCap(context.Background()); err != nil {
		t.Fatalf("FetchMarketCap failed: %v", err)
	}
	if _, err := s.FetchFearGreed(context.Background()); err != nil {
		t.Fatalf("FetchFearGreed failed: %v", err)
	}

	if elapsed := time.Since(start); elapsed < 2*time.Second {
		t.Errorf("different-endpoint consecutive requests must share queue and take >= 2s, took %v", elapsed)
	}
}
