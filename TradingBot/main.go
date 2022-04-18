package main

import (
	"log"

	"github.com/adshao/go-binance/v2"

	"github.com/ekzjuperi/binance-trading-bot/api"
	"github.com/ekzjuperi/binance-trading-bot/configs"
	b "github.com/ekzjuperi/binance-trading-bot/internal/bot"
)

func main() {
	cfg, err := configs.GetConfig()
	if err != nil {
		log.Fatalf("configs.GetConfig() err: %v", err)
	}

	binance.UseTestnet = cfg.UseTestNet
	log.Printf("Use binance test net? %v\n", cfg.UseTestNet)

	client := binance.NewClient(cfg.APIKey, cfg.SecretKey)

	bot := b.NewBot(client, cfg)

	serviceAPI := api.NewAPI(bot, cfg.Port)

	go func() {
		err := serviceAPI.Start()

		log.Printf("api stop work %v\n", err)
	}()

	bot.Start()
}
