package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/ekzjuperi/binance-trading-bot/internal/models"
)

// GetProfitStatictics func print profit statistic.
func GetProfitStatictics() func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		trades := []*models.FullTrade{}

		jsonFile, err := os.Open("statistics/trade-statistics.json")
		if err != nil {
			log.Printf("GetProfitStatictics() os.Open(statistics/trade-statistics.json) err: %v\n", err)

			return
		}
		defer jsonFile.Close()

		fileScanner := bufio.NewScanner(jsonFile)
		for fileScanner.Scan() {
			if fileScanner.Text() == "" {
				continue
			}

			trade := &models.FullTrade{}

			err = json.Unmarshal([]byte(fileScanner.Text()), &trade)
			if err != nil {
				log.Printf("GetProfitStatictics() json.Unmarshal(%v) err: %v\n", trade, err)

				return
			}

			trades = append(trades, trade)
		}

		err = fileScanner.Err()
		if err != nil {
			log.Printf("Error while reading file: %s\n", err)

			return
		}

		if len(trades) == 0 {
			_, err = resWriter.Write([]byte("No deals"))
			if err != nil {
				log.Printf("GetProfitStatictics() resWriter.Write(totalProfit) err: %v\n", err)
			}

			return
		}

		sort.SliceStable(trades, func(i, j int) bool {
			return trades[i].ExiteOrder.UpdateTime > trades[j].ExiteOrder.UpdateTime
		})

		totalProfit := float64(0)
		dayProfit := float64(0)

		firstData := time.Unix((trades[0].ExiteOrder.UpdateTime / int64(time.Microsecond)), 0)

		previousDay := firstData.Format("02-01-06")

		currentDay := ""

		buf := new(bytes.Buffer)

		_, err = buf.WriteString(fmt.Sprintf("\nDay %v \n", previousDay))
		if err != nil {
			log.Printf("GetProfitStatictics() buf.WriteString(previousDay) err: %v\n", err)
		}

		for i, trade := range trades {
			tm := time.Unix((trade.ExiteOrder.UpdateTime / int64(time.Microsecond)), 0)

			currentDay = tm.Format("02-01-06")

			if currentDay != previousDay {
				_, err = buf.WriteString(fmt.Sprintf("Day profit: %.2f\n\n", dayProfit))
				if err != nil {
					log.Printf("GetProfitStatictics() buf.WriteString(dayProfit) err: %v\n", err)
				}

				_, err = buf.WriteString(fmt.Sprintf("Day %v \n", currentDay))
				if err != nil {
					log.Printf("GetProfitStatictics() buf.WriteString(currentDay) err: %v\n", err)
				}

				previousDay = currentDay

				dayProfit = 0
			}

			totalProfit += trade.Profit
			dayProfit += trade.Profit

			trade := fmt.Sprintf("%v quantity:%s profit: %.2f\n", tm.Format("02-01-06 15:04"), trade.ExiteOrder.ExecutedQuantity[:6], trade.Profit)

			_, err = buf.WriteString(trade)
			if err != nil {
				log.Printf("GetProfitStatictics() buf.WriteString(trade) err: %v\n", err)
			}

			if i == len(trades)-1 {
				_, err = buf.WriteString(fmt.Sprintf("Day profit: %.2f\n", dayProfit))
				if err != nil {
					log.Printf("GetProfitStatictics() buf.WriteString(dayProfit) err: %v\n", err)
				}
			}
		}

		_, err = resWriter.Write([]byte(fmt.Sprintf("Total profit: %.2f\n", totalProfit)))
		if err != nil {
			log.Printf("GetProfitStatictics() resWriter.Write(totalProfit) err: %v\n", err)
		}

		_, err = resWriter.Write(buf.Bytes())
		if err != nil {
			log.Printf("GetProfitStatictics() resWriter.Write(buf.Bytes()) err: %v\n", err)
		}
	}
}
