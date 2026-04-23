package valueobject

type CoinMarketCapCfg struct {
	Url      string
	ApiToken string
}

func NewCoinMarketCapCfg(url, token string) CoinMarketCapCfg {
	return CoinMarketCapCfg{Url: url, ApiToken: token}
}
