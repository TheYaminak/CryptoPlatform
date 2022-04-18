package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	b "github.com/ekzjuperi/binance-trading-bot/internal/bot"
)

// GetParams func print bot parameters.
func GetParams(bot *b.Bot) func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		buf := new(bytes.Buffer)

		_, err := buf.WriteString(fmt.Sprintf("Bot ticker: %v \n", bot.GetSymbol()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.symbol) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Sum of open orders for last 24 hours: %v \n", bot.GetSumOfOpenOrders()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetSumOfOpenOrders()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Max daily sum of open orders: %v \n\n", bot.GetStopSumOfOpenOrdersForLastDay()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetStopSumOfOpenOrdersForLastDay()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Last trade price: %v \n", bot.GetLastTradePrice()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetLastTradePrice()) err: %v\n", err)
		}

		lastTimeTrade := bot.GetLastTimeTrade()
		if lastTimeTrade == 0 {
			_, err = buf.WriteString("Last time trade: 0\n\n")
			if err != nil {
				log.Printf("GetParams() buf.WriteString(bot.GetLastTimeTrade()) err: %v\n", err)
			}
		} else {
			_, err = buf.WriteString(fmt.Sprintf("Last time trade: %v \n\n", time.Unix(bot.GetLastTimeTrade(), 0)))
			if err != nil {
				log.Printf("GetParams() buf.WriteString(bot.GetLastTimeTrade()) err: %v\n", err)
			}
		}

		_, err = buf.WriteString(fmt.Sprintf("Calculated daily stop price: %v \n", bot.GetStopPrice()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetStopPrice()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Calculated weekly stop price: %v \n", bot.GetWeeklyStopPrice()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetWeeklyStopPrice()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Daily ratio for stop price: %v \n", bot.GetDailyRatioForStopPrice()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetDailyRatioForStopPrice()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Weekly ratio for stop price: %v \n\n", bot.GetWeeklyRatioForStopPrice()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetWeeklyRatioForStopPrice()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Minimal trade amount: %v \n", bot.GetTradeAmount()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetTradeAmount()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Profit in percent: %v \n", bot.GetProfitInPercent()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetProfitInPercent()) err: %v\n", err)
		}

		_, err = buf.WriteString(fmt.Sprintf("Time until last trade price will reset: %vs \n", bot.GetTimeUntilLastTradePriceWillReset()))
		if err != nil {
			log.Printf("GetParams() buf.WriteString(bot.GetTimeUntilLastTradePriceWillReset()) err: %v\n", err)
		}

		_, err = resWriter.Write(buf.Bytes())
		if err != nil {
			log.Printf("GetParams() resWriter.Write(buf.Bytes()) err: %v\n", err)
		}
	}
}
