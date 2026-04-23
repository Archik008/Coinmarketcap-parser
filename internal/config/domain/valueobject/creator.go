package valueobject

type CreatorCfg struct {
	TgId int
}

func NewCreatorCfg(tgId int) CreatorCfg {
	return CreatorCfg{TgId: tgId}
}
