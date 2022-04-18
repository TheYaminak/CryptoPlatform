package api

import (
	"fmt"
	"net/http"

	b "github.com/ekzjuperi/binance-trading-bot/internal/bot"
	"github.com/ekzjuperi/binance-trading-bot/internal/handlers"
)

type API struct {
	bot  *b.Bot
	port string
}

// NewAPI creates new Api instance.
func NewAPI(
	bot *b.Bot,
	port string,
) *API {
	return &API{
		bot:  bot,
		port: port,
	}
}

// Start runs API.
func (o *API) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/info", handlers.GetAccountInfo(o.bot.GetClient()))
	mux.HandleFunc("/orders", handlers.GetListOpenOrders(o.bot.GetClient(), o.bot.GetSymbol()))
	mux.HandleFunc("/profit", handlers.GetProfitStatictics())
	mux.HandleFunc("/get-logs", handlers.GetLogs())
	mux.HandleFunc("/get-params", handlers.GetParams(o.bot))
	mux.HandleFunc("/set-stop-price", o.bot.SetStopPrice())

	return http.ListenAndServe(fmt.Sprintf(":%v", o.port), mux)
}
