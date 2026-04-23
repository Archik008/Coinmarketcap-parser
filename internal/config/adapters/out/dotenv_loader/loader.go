package dotenvloader

import (
	"crypto_parser/internal/config/domain/valueobject"
	"io"
	"strconv"

	"github.com/joho/godotenv"
)

type DotEnvCfg struct {
	r             io.Reader
	CoinMarketCap valueobject.CoinMarketCapCfg
	BotCfg        valueobject.BotCfg
	CreatorCfg    valueobject.CreatorCfg
}

func NewDotEnvCfg(r io.Reader) *DotEnvCfg {
	return &DotEnvCfg{r: r}
}

func (d *DotEnvCfg) Load() error {
	env, err := godotenv.Parse(d.r)
	if err != nil {
		return err
	}

	intID, err := strconv.Atoi(env["CREATOR_USER_ID"])
	if err != nil {
		return err
	}

	d.CoinMarketCap = valueobject.NewCoinMarketCapCfg(env["COINMARKETCAP_URL"], env["COINMARKETCAP_API_KEY"])
	d.BotCfg = valueobject.NewBotCfg(env["BOT_TOKEN"])
	d.CreatorCfg = valueobject.NewCreatorCfg(intID)

	return nil
}
