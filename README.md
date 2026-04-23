# Crypto Market Parser Bot

A Telegram bot that fetches live crypto market data from CoinMarketCap and delivers a PDF report on demand.

## Report preview

<!-- Add a screenshot: put the image at assets/report_preview.png and uncomment the line below -->
[Report preview](assets/report_preview.png)

## Features

- **PDF report** with a 24h % change chart, metrics, progress bars, and metric descriptions
- **4 market indicators** — Market Cap, CMC20 Index, Fear & Greed Index, Altcoin Season Index
- **Built-in rate limiting** — 2-second cooldown between CoinMarketCap API calls
- **Graceful shutdown** — `/quit` command with inline confirmation, clean goroutine teardown
- **Access control** — `/quit` is restricted to the configured creator Telegram ID

## Bot commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message |
| `/report` | Fetch data and send a PDF report |
| `/quit` | Stop the bot (creator only, requires confirmation) |

## Getting started

### Prerequisites

- Go 1.26+
- CoinMarketCap API key ([free tier](https://coinmarketcap.com/api/))
- Telegram bot token from [@BotFather](https://t.me/BotFather)

### Configuration

Copy the sample env file and fill in your values:

```bash
cp .env.sample .env
```

`.env.sample`:

```env
CMC_URL=https://pro-api.coinmarketcap.com
CMC_API_TOKEN=your_coinmarketcap_api_key

BOT_API_TOKEN=your_telegram_bot_token
CREATOR_TG_ID=your_telegram_user_id
```

### Run

```bash
go run ./cmd/parser
```

## Project structure

```
cmd/
  parser/         — entry point
internal/
  config/         — env config loading (dotenv)
  parser/
    domain/       — domain models (MarketCap, CMC20, FearGreed, AltcoinSeason)
    application/  — ParserService, use-case ports
    adapters/
      out/coinmarketcap/  — CoinMarketCap API scraper with rate limiting
    infra/        — Telegram bot (gotgbot/v2)
  reporting/
    domain/       — ReportData value object, ReportGenerator port
    adapters/out/ — PDF generation (fpdf + go-chart)
```

## Tech stack

| Concern | Library |
|---------|---------|
| Telegram bot | [gotgbot/v2](https://github.com/PaulSonOfLars/gotgbot) |
| PDF generation | [go-pdf/fpdf](https://github.com/go-pdf/fpdf) |
| Chart rendering | [go-chart/v2](https://github.com/wcharczuk/go-chart) |
| Env config | [godotenv](https://github.com/joho/godotenv) |
