package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adshao/go-binance/v2"
)

func GetListOpenOrders(client *binance.Client, symbol string) func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		openOrders, err := client.NewListOpenOrdersService().Symbol(symbol).
			Do(context.Background())
		if err != nil {
			log.Printf("client.NewListOpenOrdersService() err: %v\n", err)
			return
		}

		log.Printf("open orders: %v\n", openOrders)

		bb, err := json.Marshal(openOrders)
		if err != nil {
			log.Printf("json.Marshal(model) err: %v\n", err)
		}

		_, err = resWriter.Write(bb)
		if err != nil {
			log.Printf("resWriter.Write() err: %v\n", err)
		}
	}
}
