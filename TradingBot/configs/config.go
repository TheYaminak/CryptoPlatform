package configs

import (
	"os"

	"gopkg.in/yaml.v2"
)

type BotConfig struct {
	APIKey     string `yaml:"api-key"`
	SecretKey  string `yaml:"secret-key"`
	UseTestNet bool   `yaml:"use-test-net"`
	Symbol     string `yaml:"symbol"`
	Port       string `yaml:"port"`

	TradeAmount                      float64 `yaml:"trade-amount"` // minimal trade amount
	ProfitInPercent                  float64 `yaml:"profit-in-percent"`
	StopSumOfOpenOrdersForLastDay    float64 `yaml:"stop-sum-of-open-orders-for-last-day"`
	DailyRatioForStopPrice           float64 `yaml:"daily-ratio-for-stop-price"`
	WeeklyRatioForStopPrice          float64 `yaml:"weekly-ratio-for-stop-price"`
	TimeUntilLastTradePriceWillReset int64   `yaml:"time-until-last-trade-price-will-reset"`
}

func GetConfig() (*BotConfig, error) {
	f, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg BotConfig

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
