package coinmarketcap

// DTO (Data Transfer Object) — структуры, которые точно повторяют форму JSON от API.
// Они живут только в адаптере: домен про них не знает.
// json-теги говорят декодеру какой ключ JSON → какое поле Go.

type marketCapDTO struct {
	Data struct {
		Quote struct {
			USD struct {
				TotalMarketCap float64 `json:"total_market_cap"`
				Change24hPct   float64 `json:"total_market_cap_yesterday_percentage_change"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}

// fearGreedResponse — /v3/fear-and-greed/historical возвращает массив записей.
// Нас интересует data[0] — самая свежая.
type fearGreedResponse struct {
	Data []fearGreedEntry `json:"data"`
}

type fearGreedEntry struct {
	Value               int    `json:"value"`
	ValueClassification string `json:"value_classification"`
}

// cmc20HistoricalResponse — /v3/index/cmc20-historical возвращает массив точек.
// Change считаем сами из двух последних записей.
type cmc20HistoricalResponse struct {
	Data []cmc20Entry `json:"data"`
}

type cmc20Entry struct {
	Value float64 `json:"value"`
}

// For altcoin season solving

type ListingsResponse struct {
	Data   []Coin `json:"data"`
	Status Status `json:"status"`
}

type Coin struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	CmcRank int    `json:"cmc_rank"`
	Quote   Quote  `json:"quote"`
}

type Quote struct {
	USD USDQuote `json:"USD"`
}

type USDQuote struct {
	Price            float64 `json:"price"`
	Volume24h        float64 `json:"volume_24h"`
	MarketCap        float64 `json:"market_cap"`
	PercentChange1h  float64 `json:"percent_change_1h"`
	PercentChange24h float64 `json:"percent_change_24h"`
	PercentChange7d  float64 `json:"percent_change_7d"`
}

type Status struct {
	Timestamp    string `json:"timestamp"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
