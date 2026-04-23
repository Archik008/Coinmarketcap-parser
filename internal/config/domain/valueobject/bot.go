package valueobject

type BotCfg struct {
	ApiToken string
}

func NewBotCfg(apiToken string) BotCfg {
	return BotCfg{ApiToken: apiToken}
}
