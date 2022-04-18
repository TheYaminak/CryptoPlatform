package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adshao/go-binance/v2"
)

func GetAccountInfo(client *binance.Client) func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		res, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			log.Printf("o.client.NewGetAccountService() err: %v\n", err)
			return
		}

		var notZeroAssets []binance.Balance

		for _, ass := range res.Balances {
			if ass.Asset == "BUSD" || ass.Asset == "USDT" || ass.Asset == "BTC" || ass.Asset == "AXS" || ass.Asset == "BNB" || ass.Asset == "BTTC" {
				notZeroAssets = append(notZeroAssets, ass)
			}
		}

		bb, err := json.Marshal(notZeroAssets)
		if err != nil {
			log.Printf("json.Marshal(model) err: %v\n", err)
		}

		log.Printf("account info: %v\n", string(bb))

		_, err = resWriter.Write(bb)
		if err != nil {
			log.Printf("resWriter.Write() err: %v\n", err)
		}
	}
}
