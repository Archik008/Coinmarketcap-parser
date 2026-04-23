package dotenvloader

import (
	"crypto_parser/internal/config/domain/valueobject"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DotEnvCfg struct {
	path          string
	CoinMarketCap valueobject.CoinMarketCapCfg
	BotCfg        valueobject.BotCfg
	CreatorCfg    valueobject.CreatorCfg
}

func NewDotEnvCfg(path string) *DotEnvCfg {
	return &DotEnvCfg{path: path}
}

func (d *DotEnvCfg) Load() error {
	if err := godotenv.Load(d.path); err != nil {
		return err
	}

	url := os.Getenv("COINMARKETCAP_URL")
	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
	botApiKey := os.Getenv("BOT_TOKEN")
	creator_id := os.Getenv("CREATOR_USER_ID")

	intID, err := strconv.Atoi(creator_id)
	if err != nil {
		return err
	}

	d.CoinMarketCap = valueobject.NewCoinMarketCapCfg(url, apiKey)
	d.BotCfg = valueobject.NewBotCfg(botApiKey)
	d.CreatorCfg = valueobject.NewCreatorCfg(intID)

	return nil
}
