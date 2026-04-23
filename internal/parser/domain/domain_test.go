package domain_test

import (
	"crypto_parser/internal/parser/domain"
	"crypto_parser/internal/parser/domain/app_errors"
	"errors"
	"testing"
	"time"
)

// --- MarketCap ---

func TestNewMarketCap(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		change    float64
		wantErr   error
	}{
		{"valid positive change", 1_000_000, 2.5, nil},
		{"valid negative change", 1_000_000, -5.3, nil},
		{"zero value", 0, 0, app_errors.ErrNonPositiveValue},
		{"negative value", -500, 1.0, app_errors.ErrNonPositiveValue},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mc, err := domain.NewMarketCap(tc.value, tc.change)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if mc.Value != tc.value {
				t.Errorf("Value: got %v, want %v", mc.Value, tc.value)
			}
			if mc.Change24hPct != tc.change {
				t.Errorf("Change24hPct: got %v, want %v", mc.Change24hPct, tc.change)
			}
			if mc.CapturedAt.IsZero() || mc.CapturedAt.After(time.Now().UTC()) {
				t.Errorf("CapturedAt looks wrong: %v", mc.CapturedAt)
			}
		})
	}
}

// --- CMC20 ---

func TestNewCMC20(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		change  float64
		wantErr error
	}{
		{"valid positive change", 500_000, 1.2, nil},
		{"valid negative change", 500_000, -8.7, nil}, // рынок упал — change отрицательный, это норм
		{"zero value", 0, 0, app_errors.ErrNonPositiveValue},
		{"negative value", -1, 0, app_errors.ErrNonPositiveValue},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmc, err := domain.NewCMC20(tc.value, tc.change)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if cmc.Value != tc.value {
				t.Errorf("Value: got %v, want %v", cmc.Value, tc.value)
			}
			if cmc.Change24hPct != tc.change {
				t.Errorf("Change24hPct: got %v, want %v", cmc.Change24hPct, tc.change)
			}
		})
	}
}

// --- FearGreed ---

func TestNewFearGreed(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		label   string
		wantErr error
	}{
		{"extreme fear boundary", 0, string(domain.ExtremeFear), nil},
		{"fear", 25, string(domain.Fear), nil},
		{"neutral", 50, string(domain.Neutral), nil},
		{"greed", 75, string(domain.Greed), nil},
		{"extreme greed boundary", 100, string(domain.ExtremeGreed), nil},
		{"negative value", -1, string(domain.Fear), app_errors.ErrValueOutOfRange},
		{"value over 100", 101, string(domain.Greed), app_errors.ErrValueOutOfRange},
		{"invalid label", 50, "Unknown", app_errors.ErrInvalidLabel},
		{"empty label", 50, "", app_errors.ErrInvalidLabel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fg, err := domain.NewFearGreed(tc.value, tc.label)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if fg.Value != tc.value {
				t.Errorf("Value: got %v, want %v", fg.Value, tc.value)
			}
			if string(fg.Label) != tc.label {
				t.Errorf("Label: got %v, want %v", fg.Label, tc.label)
			}
		})
	}
}

// --- AltcoinSeason ---

func TestNewAltcoinSeason(t *testing.T) {
	tests := []struct {
		name         string
		index        float64
		total        int
		outperformed int
		wantAlt      bool
		wantErr      error
	}{
		{"bitcoin season", 30.0, 100, 30, false, nil},
		{"boundary just above alt season", 25.1, 100, 25, false, nil},
		{"alt season boundary", 25.0, 100, 25, true, nil},
		{"full alt season", 0.0, 50, 0, true, nil},
		{"index negative", -1.0, 100, 0, false, app_errors.ErrValueOutOfRange},
		{"index over 100", 101.0, 100, 0, false, app_errors.ErrValueOutOfRange},
		{"total zero", 0.0, 0, 0, false, app_errors.ErrNonPositiveValue},
		{"total negative", 30.0, -1, 0, false, app_errors.ErrNonPositiveValue},
		{"outperformed negative", 30.0, 100, -1, false, app_errors.ErrValueOutOfRange},
		{"outperformed exceeds total", 30.0, 10, 11, false, app_errors.ErrValueOutOfRange},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			as, err := domain.NewAltcoinSeason(tc.index, tc.total, tc.outperformed)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err=%v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if as.Index != tc.index {
				t.Errorf("Index: got %v, want %v", as.Index, tc.index)
			}
			if as.Total != tc.total {
				t.Errorf("Total: got %v, want %v", as.Total, tc.total)
			}
			if as.Outperformed != tc.outperformed {
				t.Errorf("Outperformed: got %v, want %v", as.Outperformed, tc.outperformed)
			}
			if as.IsAltSeason != tc.wantAlt {
				t.Errorf("IsAltSeason: got %v, want %v", as.IsAltSeason, tc.wantAlt)
			}
			if as.CapturedAt.IsZero() {
				t.Error("CapturedAt must not be zero")
			}
		})
	}
}
