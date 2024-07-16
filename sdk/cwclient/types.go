package cwclient

type L2Block struct {
	BlockHash      string `mapstructure:"block-hash"`
	BlockHeight    uint64 `mapstructure:"block-height"`
	BlockTimestamp uint64 `mapstructure:"block-timestamp"`
}
