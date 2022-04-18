package utils

import (
	"log"
	"math"
	"strconv"
)

// CalculateProfit convert deals from string to float, and round profit.
func CalculateProfit(firstDeal, secondDeal string) float64 {
	firstDealFloat, err := strconv.ParseFloat(firstDeal, 32)
	if err != nil {
		log.Printf("CalculateProfit() strconv.ParseFloat(%v) err: %v", firstDeal, err)
	}

	secondDealFloat, err := strconv.ParseFloat(secondDeal, 32)
	if err != nil {
		log.Printf("CalculateProfit() strconv.ParseFloat(%v) err: %v", secondDeal, err)
	}

	return math.Round((secondDealFloat-firstDealFloat)*100) / 100
}
