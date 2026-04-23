package main

import (
	"fmt"
	"log"
	"time"

	"crypto_parser/internal/parser/domain/dto"
	"crypto_parser/internal/reporting/adapters/out"
	"crypto_parser/internal/reporting/domain/valueobject"
)

func main() {
	scenarios := []struct {
		name string
		data valueobject.ReportData
	}{
		{
			// Index=19: BTC обогнал лишь 19 из 99 монет → 80+ монет бьют BTC → Alt Season
			name: "alt-season / extreme greed",
			data: valueobject.ReportData{
				GeneratedAt: time.Now().UTC(),
				MarketCap: dto.MarketCapDTO{
					Value:        2_850_000_000_000,
					Change24hPct: 3.42,
				},
				CMC20: dto.CMC20DTO{
					Value:        1_120_000_000_000,
					Change24hPct: 5.17,
				},
				FearGreed: dto.FearGreedDTO{
					Value: 88,
					Label: "Extreme Greed",
				},
				AltcoinSeason: dto.AltcoinSeasonDTO{
					Index:        19.0,
					Total:        99,
					Outperformed: 19,
					IsAltSeason:  true,
				},
			},
		},
		{
			// Index=66: BTC обогнал 64 из 97 монет → только 33 бьют BTC → Bitcoin Season
			name: "bitcoin-season / fear",
			data: valueobject.ReportData{
				GeneratedAt: time.Now().UTC(),
				MarketCap: dto.MarketCapDTO{
					Value:        1_540_000_000_000,
					Change24hPct: -2.85,
				},
				CMC20: dto.CMC20DTO{
					Value:        620_000_000_000,
					Change24hPct: -4.10,
				},
				FearGreed: dto.FearGreedDTO{
					Value: 28,
					Label: "Fear",
				},
				AltcoinSeason: dto.AltcoinSeasonDTO{
					Index:        66.0,
					Total:        97,
					Outperformed: 64,
					IsAltSeason:  false,
				},
			},
		},
		{
			// Index=42: BTC обогнал 42 из 100 монет → 58 монет бьют BTC, но < 75% → Bitcoin Season
			name: "neutral market",
			data: valueobject.ReportData{
				GeneratedAt: time.Now().UTC(),
				MarketCap: dto.MarketCapDTO{
					Value:        2_100_000_000_000,
					Change24hPct: 0.31,
				},
				CMC20: dto.CMC20DTO{
					Value:        840_000_000_000,
					Change24hPct: -0.55,
				},
				FearGreed: dto.FearGreedDTO{
					Value: 51,
					Label: "Neutral",
				},
				AltcoinSeason: dto.AltcoinSeasonDTO{
					Index:        42.0,
					Total:        100,
					Outperformed: 42,
					IsAltSeason:  false,
				},
			},
		},
	}

	reportService := out.NewPdfGenerator("temp_files")

	for _, s := range scenarios {
		path, err := reportService.Generate(s.data)
		if err != nil {
			log.Fatalf("[%s] generate failed: %v", s.name, err)
		}
		fmt.Printf("✓ %-35s → %s\n", s.name, path)
	}
}
