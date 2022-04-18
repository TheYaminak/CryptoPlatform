package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"

	"github.com/ekzjuperi/binance-trading-bot/configs"
	"github.com/ekzjuperi/binance-trading-bot/internal/cache"
	"github.com/ekzjuperi/binance-trading-bot/internal/models"
	"github.com/ekzjuperi/binance-trading-bot/internal/utils"
)

const (
	timeOutForTimer                        = 180
	pauseAfterTrade                        = 180
	timeOutToCheckEntryOrder               = 3
	timeOutInMinForGetStopPrice            = 5
	timeOutForGetSumOfOpenOrdersForLastDay = 5
	timeOutInSecForCheckLimitOrders        = 15
	sizeChan                               = 100

	millisecondInDay  = int64(1000 * 24 * 60 * 60)
	millisecondInWeek = millisecondInDay * 7
	bitSize32         = 32
	interval15Min     = "15m"
	interval1Hour     = "1h"
)

type Bot struct {
	client                           *binance.Client
	cache                            *cache.Cache
	analysisChan                     chan *binance.WsAggTradeEvent
	orderChan                        chan *models.Order
	fullTradeChan                    chan *models.FullTrade
	symbol                           string // trading pair
	profitInPercent                  float64
	tradeAmount                      float64
	stopSumOfOpenOrdersForLastDay    float64
	dailyRatioForStopPrice           float64
	weeklyRatioForStopPrice          float64
	timeUntilLastTradePriceWillReset int64

	sumOfOpenOrdersForLastDay float64
	lastTimeTrade             int64 // unix time from last trade
	dayStopPrice              float64
	weekStopPrice             float64
	lastTradePrice            float64

	wg  *sync.WaitGroup
	rwm *sync.RWMutex
}

// NewBot func initializes the bot.
func NewBot(client *binance.Client, cfg *configs.BotConfig) *Bot {
	bot := Bot{
		client:                           client,
		cache:                            cache.NewCache(),
		analysisChan:                     make(chan *binance.WsAggTradeEvent, sizeChan),
		orderChan:                        make(chan *models.Order, sizeChan),
		fullTradeChan:                    make(chan *models.FullTrade, sizeChan),
		symbol:                           cfg.Symbol,
		profitInPercent:                  cfg.ProfitInPercent,
		tradeAmount:                      cfg.TradeAmount,
		stopSumOfOpenOrdersForLastDay:    cfg.StopSumOfOpenOrdersForLastDay,
		dailyRatioForStopPrice:           cfg.DailyRatioForStopPrice,
		weeklyRatioForStopPrice:          cfg.WeeklyRatioForStopPrice,
		timeUntilLastTradePriceWillReset: cfg.TimeUntilLastTradePriceWillReset,

		wg:  &sync.WaitGroup{},
		rwm: &sync.RWMutex{},
	}

	return &bot
}

// Start func start bot.
func (o *Bot) Start() {
	go o.getStopPrice(&o.dayStopPrice, millisecondInDay, interval15Min, o.dailyRatioForStopPrice)

	go o.getStopPrice(&o.weekStopPrice, millisecondInWeek, interval1Hour, o.weeklyRatioForStopPrice)

	go o.getSumOfOpenTradesForLastDay()

	o.wg.Add(1)

	go o.analyze()

	o.wg.Add(1)

	go o.trade()

	o.wg.Add(1)

	go o.checkLimitOrders()

	o.wg.Wait()

	log.Println("Bot stop work")
}

// StartPricesStream func stream prices from binance.
func (o *Bot) StartPricesStream() (chan struct{}, error) {
	wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
		o.analysisChan <- event
	}
	errHandler := func(err error) {
		log.Println(err)
	}

	doneC, _, err := binance.WsAggTradeServe(o.symbol, wsAggTradeHandler, errHandler)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return doneC, nil
}

// analyze func analyzes prices from the exchange stream.
func (o *Bot) analyze() {
	defer o.wg.Done()

	doneC, err := o.StartPricesStream()
	if err != nil {
		log.Printf("o.StartPricesStream() err: %v\n", err)

		return
	}

	oldEvent := <-o.analysisChan
	oldEvent2 := <-o.analysisChan
	oldEvent3 := <-o.analysisChan
	oldEvent4 := <-o.analysisChan
	oldEvent5 := <-o.analysisChan

	price, _ := strconv.ParseFloat(oldEvent.Price, bitSize32)
	o.lastTradePrice = price
	o.lastTimeTrade = time.Now().Unix()

	timer := time.NewTimer(time.Second * 120)
	timer2 := time.NewTimer(time.Second * 90)
	timer3 := time.NewTimer(time.Second * 60)
	timer4 := time.NewTimer(time.Second * 30)
	timer5 := time.NewTimer(time.Second * 0)

	log.Println("Start trading")

	for {
		select {
		case <-timer.C:
			o.makeDecision(oldEvent)
			timer.Reset(time.Second * timeOutForTimer)

		case <-timer2.C:
			o.makeDecision(oldEvent2)
			timer2.Reset(time.Second * timeOutForTimer)

		case <-timer3.C:
			o.makeDecision(oldEvent3)
			timer3.Reset(time.Second * timeOutForTimer)

		case <-timer4.C:
			o.makeDecision(oldEvent4)
			timer4.Reset(time.Second * timeOutForTimer)

		case <-timer5.C:
			o.makeDecision(oldEvent5)
			timer5.Reset(time.Second * timeOutForTimer)

		default:
			select {
			case <-doneC:
				// if doneC send event, restart o.Analyze()
				log.Println("doneC send event")

				o.wg.Add(1)

				go o.analyze()

				return

			case <-o.analysisChan:
				continue
			}
		}
	}
}

func (o *Bot) makeDecision(oldEvent *binance.WsAggTradeEvent) {
	newEvent := <-o.analysisChan
	newEventPrice, _ := strconv.ParseFloat(newEvent.Price, bitSize32)
	oldEventPrice, _ := strconv.ParseFloat(oldEvent.Price, bitSize32)

	difference := newEventPrice / oldEventPrice * 100

	var order *models.Order
	if difference < 99.55 {
		order = &models.Order{
			Symbol:   newEvent.Symbol,
			Price:    newEventPrice,
			Quantity: 2 * o.tradeAmount / newEventPrice,
		}
	} else if difference < 99.70 {
		order = &models.Order{
			Symbol:   newEvent.Symbol,
			Price:    newEventPrice,
			Quantity: 1.5 * o.tradeAmount / newEventPrice,
		}
	} else if difference < 99.80 {
		order = &models.Order{
			Symbol:   newEvent.Symbol,
			Price:    newEventPrice,
			Quantity: o.tradeAmount / newEventPrice,
		}
	}

	*oldEvent = *newEvent

	if order == nil {
		return
	}

	timeFromLastTrade := (time.Now().Unix() - o.lastTimeTrade)

	// if enough time has passed since the last trade, reset lastTradePrice and lastTimeTrade.
	if o.lastTimeTrade != 0 && timeFromLastTrade > o.timeUntilLastTradePriceWillReset {
		o.rwm.Lock()
		o.lastTradePrice = 0
		o.lastTimeTrade = 0
		o.rwm.Unlock()
	}

	// if order price >= stop price, skip trade.
	if newEventPrice >= o.dayStopPrice {
		log.Printf("order %v skip, price(%v) > o.dayStopPrice(%v)\n", order, order.Price, o.dayStopPrice)

		return
	}

	// if order price >= stop price, skip trade.
	if newEventPrice >= o.weekStopPrice {
		log.Printf("order %v skip, price(%v) > o.weekStopPrice(%v)\n", order, order.Price, o.weekStopPrice)

		return
	}

	// if sumOfOpenOrdersForLastDay > stopSumOfOpenOrdersForLastDay skip trade.
	if o.sumOfOpenOrdersForLastDay > o.stopSumOfOpenOrdersForLastDay {
		log.Printf("order: %v skip, the worth of open trades: %v > day limit: %v\n",
			order,
			o.sumOfOpenOrdersForLastDay,
			o.stopSumOfOpenOrdersForLastDay,
		)

		return
	}

	order.Price = math.Round((order.Price)*100) / 100
	order.Quantity = math.Round((order.Quantity)*10000) / 10000

	o.orderChan <- order

	log.Println("price difference = ", difference)
}

func (o *Bot) trade() {
	defer o.wg.Done()

	// get order from orderChan.
	for order := range o.orderChan {
		timeFromLastTrade := (time.Now().Unix() - o.lastTimeTrade)

		// if not enough time has passed since the last trade, skip trade.
		if timeFromLastTrade < pauseAfterTrade {
			log.Printf("order: %v skip, %v s. has passed since the last trade s\n", order, timeFromLastTrade)

			continue
		}

		// if last trade price >= new price, skip trade.
		if (o.lastTradePrice != 0) && order.Price >= o.lastTradePrice {
			log.Printf("order: %v skip, order.Price(%v) >= o.lastTradePrice(%v)\n", order, order.Price, o.lastTradePrice)

			continue
		}

		log.Printf("order: %v\n", order)

		// get the highest price bid in the order book.
		depth, err := o.client.NewDepthService().Symbol(order.Symbol).
			Do(context.Background())
		if err != nil {
			log.Printf("o.client.NewDepthService(%v) err: %v\n", order.Symbol, err)
			continue
		}

		price, err := strconv.ParseFloat(depth.Bids[0].Price, bitSize32)
		if err != nil {
			log.Printf("strconv.ParseFloat(depth.Bids[0].Price) err: %v\n", err)
			continue
		}

		order.Price = price

		// create order
		firstOrderResolve, err := o.createOrder(order, binance.SideTypeBuy, binance.OrderTypeLimitMaker, binance.TimeInForceTypeGTC)
		if err != nil && err.Error() == "<APIError> code=-2010, msg=Account has insufficient balance for requested action." {
			log.Printf("Account has insufficient balance for buy %v\n", order)

			continue
		}

		if err != nil {
			log.Printf("An error occurred during order execution, order: %v, err: %v\n", order, err)

			// if  api response have err, retry create order
			firstOrderResolve, order, err = o.retryCreateOrder(order)
			if err != nil {
				log.Println(err)

				continue
			}
		}

		o.rwm.Lock()
		o.lastTimeTrade = time.Now().Unix()
		o.lastTradePrice = order.Price - (order.Price * o.profitInPercent / 2)
		o.rwm.Unlock()

		go o.checkEntryOrder(firstOrderResolve, order)
	}
}

func (o *Bot) checkEntryOrder(firstOrderResolve *binance.CreateOrderResponse, order *models.Order) {
	timeNow := time.Now().Unix()

	var entryBinanceOrder *binance.Order

	var err error

	for {
		orderTimeout := (time.Now().Unix() - timeNow)

		entryBinanceOrder, err = o.client.NewGetOrderService().Symbol(firstOrderResolve.Symbol).
			OrderID(firstOrderResolve.OrderID).Do(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		if entryBinanceOrder.Status == binance.OrderStatusTypeFilled {
			log.Printf("entry order execute price: %v, quantity: %v", entryBinanceOrder.Price, entryBinanceOrder.ExecutedQuantity)
			break
		}

		if orderTimeout > 100 && entryBinanceOrder.Status == binance.OrderStatusTypeNew {
			_, err := o.client.NewCancelOrderService().Symbol(firstOrderResolve.Symbol).
				OrderID(firstOrderResolve.OrderID).Do(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			o.rwm.Lock()
			o.lastTimeTrade = time.Now().Unix() - pauseAfterTrade
			o.lastTradePrice = order.Price
			o.rwm.Unlock()

			log.Printf("order %v canceled after timeout", order)

			return
		}

		time.Sleep(time.Second * timeOutToCheckEntryOrder)
	}

	log.Printf("Order %v executed\n", order)

	order.Price += order.Price * o.profitInPercent

	for {
		secondOrderResolve, err := o.createOrder(order, binance.SideTypeSell, binance.OrderTypeLimit, binance.TimeInForceTypeGTC)
		if err != nil {
			log.Printf("An error occurred during order execution, order: %v, type: %v, err: %v\n", order, binance.OrderTypeLimit, err)
			continue
		}

		log.Printf("Order limit create %v\n", secondOrderResolve)

		exitBinanceOrder, err := o.client.NewGetOrderService().Symbol(secondOrderResolve.Symbol).
			OrderID(secondOrderResolve.OrderID).Do(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		cTrade := &models.FullTrade{
			EnterOrder: entryBinanceOrder,
			ExiteOrder: exitBinanceOrder,
		}

		o.fullTradeChan <- cTrade

		break
	}
}

func (o *Bot) createOrder(
	order *models.Order,
	side binance.SideType,
	typeOrder binance.OrderType,
	timeInForce binance.TimeInForceType,
) (*binance.CreateOrderResponse, error) {
	var orderResponse *binance.CreateOrderResponse

	var err error

	switch typeOrder {
	case binance.OrderTypeMarket:
		orderResponse, err = o.client.NewCreateOrderService().Symbol(order.Symbol).
			Side(side).
			Type(typeOrder).
			Quantity(fmt.Sprintf("%f", order.Quantity)).
			Do(context.Background())
	case binance.OrderTypeLimitMaker:
		orderResponse, err = o.client.NewCreateOrderService().Symbol(order.Symbol).
			Side(side).
			Type(typeOrder).
			Quantity(fmt.Sprintf("%f", order.Quantity)).
			Price(fmt.Sprintf("%.2f", order.Price)).
			Do(context.Background())

	case binance.OrderTypeLimit:
		orderResponse, err = o.client.NewCreateOrderService().Symbol(order.Symbol).
			Side(side).
			Type(typeOrder).
			TimeInForce(timeInForce).
			Quantity(fmt.Sprintf("%v", order.Quantity)).
			Price(fmt.Sprintf("%.2f", order.Price)).
			Do(context.Background())

	case binance.OrderTypeStopLoss:
		stopLoss := int(order.Price - order.Price*o.profitInPercent)

		orderResponse, err = o.client.NewCreateOrderService().Symbol(order.Symbol).
			Side(side).
			Type(typeOrder).
			TimeInForce(timeInForce).
			Quantity(fmt.Sprintf("%v", order.Quantity)).
			Price(fmt.Sprintf("%v", stopLoss)).
			Do(context.Background())
	}

	return orderResponse, err
}

func (o *Bot) checkLimitOrders() {
	defer o.wg.Done()

	go func() {
		for {
			time.Sleep(time.Second * timeOutInSecForCheckLimitOrders)

			ordersAwaitingCompletion := o.cache.Cache.Items()

			if len(ordersAwaitingCompletion) == 0 {
				continue
			}

			orders, err := o.client.NewListOrdersService().Symbol(o.symbol).Do(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			openOrders, err := o.client.NewListOpenOrdersService().Symbol(o.symbol).Do(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			listBinanceOrders := map[string]*binance.Order{}

			for _, openOrder := range openOrders {
				listBinanceOrders[fmt.Sprintf("%v", openOrder.OrderID)] = openOrder
			}

			for _, order := range orders {
				listBinanceOrders[fmt.Sprintf("%v", order.OrderID)] = order
			}

			for _, awaitingOrder := range ordersAwaitingCompletion {
				b, err := json.Marshal(awaitingOrder.Object)
				if err != nil {
					log.Printf("json.Marshal(awaitingOrder.Object) err: %v\n", err)

					continue
				}

				var fullTrade models.FullTrade

				err = json.Unmarshal(b, &fullTrade)
				if err != nil {
					log.Printf("json.Unmarshal(b, FullTrade) err: %v\n", err)

					continue
				}

				binanceOrder, ok := listBinanceOrders[fmt.Sprintf("%v", fullTrade.ExiteOrder.OrderID)]
				if !ok {
					log.Printf("CheckLimitOrder() didn't have awaitingOrder:%v", fullTrade.ExiteOrder.OrderID)
					continue
				}

				if binanceOrder.Status == binance.OrderStatusTypeFilled {
					fullTrade.ExiteOrder = binanceOrder

					fullTrade.Profit = utils.CalculateProfit(
						fullTrade.EnterOrder.CummulativeQuoteQuantity,
						fullTrade.ExiteOrder.CummulativeQuoteQuantity,
					)

					SaveFullTradeInFile(fullTrade)

					o.cache.Cache.Delete(fmt.Sprintf("%v", fullTrade.ExiteOrder.OrderID))

					newlastTradePrice, _ := strconv.ParseFloat(fullTrade.ExiteOrder.Price, bitSize32)

					o.rwm.Lock()
					o.cache.SaveCache()

					if o.lastTradePrice == 0 {
						o.lastTradePrice = newlastTradePrice - newlastTradePrice*o.profitInPercent
						o.lastTimeTrade = time.Now().Unix() - pauseAfterTrade
					}
					o.rwm.Unlock()

					log.Printf("exite order execute price: %v, quantity: %v profit %v",
						binanceOrder.Price,
						binanceOrder.ExecutedQuantity,
						fullTrade.Profit,
					)
				}
			}
		}
	}()

	for compleatedOrder := range o.fullTradeChan {
		err := o.cache.Cache.Add(fmt.Sprintf("%v", compleatedOrder.ExiteOrder.OrderID), compleatedOrder, 0)
		if err != nil {
			log.Printf("o.cache.Cache.Add() err: %v\n", err)
		}

		o.rwm.Lock()
		o.cache.SaveCache()
		o.rwm.Unlock()
	}
}

func SaveFullTradeInFile(fullTrade models.FullTrade) {
	file, err := os.OpenFile("statistics/trade-statistics.json", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("os.OpenFile(trade-statistics.json) err: %v\n", err)
	}

	defer file.Close()

	b, err := json.Marshal(fullTrade)
	if err != nil {
		log.Printf("json.Marshal(fullTrade) err: %v\n", err)
		return
	}

	if _, err = io.WriteString(file, "\n"); err != nil {
		log.Printf("f.WriteString(/n) err: %v\n", err)
	}

	if _, err = io.WriteString(file, string(b)); err != nil {
		log.Printf("f.Write(string(b)) err: %v\n", err)
	}
}

func (o *Bot) retryCreateOrder(order *models.Order) (*binance.CreateOrderResponse, *models.Order, error) {
	firstErrTime := time.Now().Unix()

	for time.Now().Unix()-firstErrTime <= 5 {
		time.Sleep(time.Second * 1)

		depth, err := o.client.NewDepthService().Symbol(order.Symbol).
			Do(context.Background())

		if err != nil {
			log.Printf("o.client.NewDepthService(%v) err: %v\n", order.Symbol, err)

			continue
		}

		price, err := strconv.ParseFloat(depth.Bids[0].Price, bitSize32)
		if err != nil {
			log.Printf("strconv.ParseFloat(depth.Bids[0].Price) err: %v\n", err)

			continue
		}

		if price-order.Price > 50 {
			log.Printf("newPrice: %v > oldPrice: %v\n", price, order.Price)

			continue
		}

		order.Price = price

		firstOrderResolve, err := o.createOrder(order, binance.SideTypeBuy, binance.OrderTypeLimitMaker, binance.TimeInForceTypeGTC)
		if err != nil {
			log.Printf("An error occurred during order execution, order: %v, err: %v\n", order, err)

			continue
		}

		return firstOrderResolve, order, nil
	}

	return nil, nil, fmt.Errorf("number of attempts to create an order %v exceeded", order)
}

func (o *Bot) getStopPrice(stopPriceFilter *float64, tsStartInterval int64, interval string, ratioForStopPrice float64) {
	for {
		startTS := time.Now().UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)) - tsStartInterval

		klines, err := o.client.NewKlinesService().Symbol(o.symbol).StartTime(startTS).Interval("15m").Do(context.Background())
		if err != nil {
			log.Printf("o.client.NewKlinesService(%v) err: %v\n", o.symbol, err)
			continue
		}

		minPrice := float64(0)
		maxPrice := float64(0)

		for _, kline := range klines {
			priceHigh, err := strconv.ParseFloat(kline.High, bitSize32)
			if err != nil {
				log.Printf("getStopPrice() strconv.ParseFloat(%v, 32)) err: %v\n", kline.High, err)
				continue
			}

			priceLow, err := strconv.ParseFloat(kline.Low, bitSize32)
			if err != nil {
				log.Printf("getStopPrice() strconv.ParseFloat(%v, 32)) err: %v\n", kline.Low, err)
				continue
			}

			if minPrice == 0 && priceLow != 0 {
				minPrice = priceLow
			}

			if priceHigh > maxPrice {
				maxPrice = priceHigh
			}

			if priceLow < minPrice {
				minPrice = priceLow
			}
		}

		stopPrice := maxPrice - ((maxPrice - minPrice) * ratioForStopPrice)

		o.rwm.Lock()
		*stopPriceFilter = stopPrice
		o.rwm.Unlock()

		time.Sleep(time.Minute * timeOutInMinForGetStopPrice)
	}
}

func (o *Bot) getSumOfOpenTradesForLastDay() {
	for {
		last24Hours := time.Now().UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)) - millisecondInDay

		ordersAwaitingCompletion := o.cache.Cache.Items()

		if len(ordersAwaitingCompletion) == 0 {
			continue
		}

		sumOfOpenTradesForLastDay := float64(0)

		for _, order := range ordersAwaitingCompletion {
			b, err := json.Marshal(order.Object)
			if err != nil {
				log.Printf("json.Marshal(order.Object) err: %v\n", err)
				continue
			}

			var fullTrade models.FullTrade

			err = json.Unmarshal(b, &fullTrade)
			if err != nil {
				log.Printf("json.Unmarshal(b, FullTrade) err: %v\n", err)
				continue
			}

			if fullTrade.EnterOrder.Time < last24Hours {
				continue
			}

			if fullTrade.ExiteOrder.Status == binance.OrderStatusTypeFilled {
				continue
			}

			cummulativeQuoteQuantity, _ := strconv.ParseFloat(fullTrade.EnterOrder.CummulativeQuoteQuantity, bitSize32)

			sumOfOpenTradesForLastDay += cummulativeQuoteQuantity
		}

		o.rwm.Lock()
		o.sumOfOpenOrdersForLastDay = math.Round((sumOfOpenTradesForLastDay)*100) / 100
		o.rwm.Unlock()

		time.Sleep(time.Minute * timeOutForGetSumOfOpenOrdersForLastDay)
	}
}

func (o *Bot) SetStopPrice() func(http.ResponseWriter, *http.Request) {
	return func(resWriter http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()
		price := query["price"][0]

		if price == "" {
			_, err := resWriter.Write([]byte(fmt.Sprintf("incorrect stop price = %v", price)))
			if err != nil {
				log.Printf("resWriter.Write() err: %v\n", err)
			}

			return
		}

		stopPriceFloat, _ := strconv.ParseFloat(price, bitSize32)

		o.dayStopPrice = stopPriceFloat

		_, err := resWriter.Write([]byte(fmt.Sprintf("stop price now %v", stopPriceFloat)))
		if err != nil {
			log.Printf("resWriter.Write() err: %v\n", err)
		}
	}
}

func (o *Bot) GetClient() *binance.Client {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.client
}

func (o *Bot) GetSymbol() string {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.symbol
}

func (o *Bot) GetStopPrice() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.dayStopPrice
}

func (o *Bot) GetWeeklyStopPrice() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.weekStopPrice
}

func (o *Bot) GetSumOfOpenOrders() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.sumOfOpenOrdersForLastDay
}

func (o *Bot) GetLastTimeTrade() int64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.lastTimeTrade
}

func (o *Bot) GetLastTradePrice() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.lastTradePrice
}

func (o *Bot) GetProfitInPercent() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.profitInPercent
}

func (o *Bot) GetTradeAmount() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.tradeAmount
}

func (o *Bot) GetStopSumOfOpenOrdersForLastDay() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.stopSumOfOpenOrdersForLastDay
}

func (o *Bot) GetDailyRatioForStopPrice() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.dailyRatioForStopPrice
}

func (o *Bot) GetWeeklyRatioForStopPrice() float64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.weeklyRatioForStopPrice
}

func (o *Bot) GetTimeUntilLastTradePriceWillReset() int64 {
	o.rwm.RLock()
	defer o.rwm.RUnlock()

	return o.timeUntilLastTradePriceWillReset
}
