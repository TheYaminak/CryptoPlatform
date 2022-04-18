package models

import "github.com/adshao/go-binance/v2"

// FullTrade struct contain info about deal.
type FullTrade struct {
	EnterOrder *binance.Order
	ExiteOrder *binance.Order
	Profit     float64
}
