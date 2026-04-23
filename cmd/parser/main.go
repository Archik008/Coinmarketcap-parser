package main

import (
	"context"
	dotenvloader "crypto_parser/internal/config/adapters/out/dotenv_loader"
	"crypto_parser/internal/parser/adapters/out/coinmarketcap"
	service "crypto_parser/internal/parser/application"
	"crypto_parser/internal/parser/infra"
	"crypto_parser/internal/reporting/adapters/out"
	"log"
)

func main() {
	cfg := dotenvloader.NewDotEnvCfg(".env")
	if err := cfg.Load(); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coinMarketScraper := coinmarketcap.NewScraper(cfg.CoinMarketCap, ctx)
	parserService := service.NewParserService(coinMarketScraper)

	reportService := out.NewPdfGenerator("temp_files")

	botInstance := infra.NewBot(cfg.BotCfg.ApiToken, int64(cfg.CreatorCfg.TgId), parserService, reportService)

	botInstance.Run(ctx)
}
