package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
	"github.com/limbara/stock-watcher/models"
	"github.com/limbara/stock-watcher/utils"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterCrons() {
	log.Default().Println("Creating Cron")

	c := cron.New()

	c.AddFunc(fmt.Sprintf("@every %s", 24*time.Hour), scrapeStocksCodeAndName)

	c.AddFunc(fmt.Sprintf("@every %s", 12*time.Hour), scrapeStockPriceSummary)

	log.Default().Println("Running Cron")
	c.Start()
}

type StockCodeAndNameDTO struct {
	Code string `validate:"required,uppercase,alpha"`
	Name string `validate:"required"`
}

func scrapeStocksCodeAndName() {
	logger, err := utils.Logger()
	if err != nil {
		panic(fmt.Errorf("Error Create Logger in scrapeStocksCodeAndName : %w", err))
	}

	urls := make([]string, 26)
	for i := 0; i < 26; i++ {
		urls[i] = fmt.Sprintf("https://www.duniainvestasi.com/bei/bulks/index/%c", rune(i+65))
	}

	q, err := queue.New(
		2, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)
	if err != nil {
		logger.Sugar().Panic(fmt.Errorf("Error init queue in scrapeStocksCodeAndName : %w", err))
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.duniainvestasi.com"),
	)
	extensions.RandomUserAgent(c)

	c.OnError(func(r *colly.Response, err error) {
		logger.Sugar().Infof("Error Scrapping %s\n : %+v", r.Request.URL, err)
	})

	c.OnRequest(func(r *colly.Request) {
		logger.Sugar().Infof("Visiting %s\n", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		logger.Sugar().Infof("Scrapping %s with Status %d\n", r.Request.URL, r.StatusCode)
	})

	validate := validator.New()
	var stockDTOs []*StockCodeAndNameDTO

	countScrapped := 0

	done := make(chan struct{}, 1)

	c.OnHTML("#CONTENT table tbody tr", func(e *colly.HTMLElement) {
		filteredChildren := e.DOM.Children().Filter("td")

		if filteredChildren.Children().Length() >= 2 && filteredChildren.Eq(0) != nil && filteredChildren.Eq(1) != nil {
			dto := StockCodeAndNameDTO{
				Code: filteredChildren.Eq(0).Text(),
				Name: filteredChildren.Eq(1).Text(),
			}
			err := validate.Struct(dto)

			if err == nil {
				stockDTOs = append(stockDTOs, &dto)
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		countScrapped++

		if countScrapped >= len(urls) {
			done <- struct{}{}
		}
	})

	for _, url := range urls {
		q.AddURL(url)
	}

	if err := q.Run(c); err != nil {
		logger.Sugar().Info("Error Run Scrapping Queue", err)
	}

	// wait to be signalled back
	<-done

	var validatedStockDTOs []mongo.WriteModel

	for _, dto := range stockDTOs {
		err := validate.Struct(dto)
		if err == nil {
			filter := bson.M{"code": dto.Code}
			update := bson.M{
				"$set":         bson.M{"name": dto.Name},
				"$setOnInsert": bson.M{"code": dto.Code},
			}
			validatedStockDTOs = append(validatedStockDTOs, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, errSave := models.Db.GetRepo("StockRepo").Collection().BulkWrite(ctx, validatedStockDTOs)
	if errSave != nil {
		logger.Sugar().Info("Error Saving scrapeStocksCodeAndName", errSave)
	} else {
		logger.Sugar().Infof("Finish Scrapping scrapeStocksCodeAndName, upsert %d stocks", len(validatedStockDTOs))
	}
}

type StockSummary struct {
	Code  string `validate:"required,uppercase,alpha"` // also protect from warant stock ALDO-W
	Open  string `validate:"required,numeric"`
	Close string `validate:"required,numeric"`
	High  string `validate:"required,numeric"`
	Low   string `validate:"required,numeric"`
}

func scrapeStockPriceSummary() {
	logger, err := utils.Logger()
	if err != nil {
		panic(fmt.Errorf("Error Create Logger in scrapeStocksCodeAndName : %w", err))
	}

	currentTime := time.Now().AddDate(0, 0, -1)

	c := colly.NewCollector(
		colly.AllowedDomains("www.duniainvestasi.com"),
	)
	extensions.RandomUserAgent(c)

	c.OnError(func(r *colly.Response, err error) {
		logger.Sugar().Infof("Error Scrapping %s\n : %+v\n", r.Request.URL, err)
	})

	c.OnRequest(func(r *colly.Request) {
		logger.Sugar().Infof("Visiting %s\n", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		logger.Sugar().Infof("Scrapping %s with Status %d\n", r.Request.URL, r.StatusCode)
	})

	hasToScrape := make(chan bool, 1)
	validate := validator.New()
	var stockDTOs []*StockSummary

	c.OnHTML("#CONTENT table tbody tr", func(e *colly.HTMLElement) {
		filteredChildren := e.DOM.Children().Filter("td")

		hasCodeColumn := filteredChildren.Eq(0) != nil
		hasPrevColumn := filteredChildren.Eq(1) != nil
		hasOpenColumn := filteredChildren.Eq(2) != nil
		hasHighColumn := filteredChildren.Eq(3) != nil
		hasLowColumn := filteredChildren.Eq(4) != nil
		hasCloseColumn := filteredChildren.Eq(5) != nil

		if filteredChildren.Children().Length() >= 6 && hasCodeColumn && hasPrevColumn && hasOpenColumn && hasHighColumn && hasLowColumn && hasCloseColumn {
			dto := StockSummary{
				Code:  filteredChildren.Eq(0).Text(),
				Open:  strings.ReplaceAll(filteredChildren.Eq(1).Text(), ",", ""),
				Close: strings.ReplaceAll(filteredChildren.Eq(5).Text(), ",", ""),
				High:  strings.ReplaceAll(filteredChildren.Eq(3).Text(), ",", ""),
				Low:   strings.ReplaceAll(filteredChildren.Eq(4).Text(), ",", ""),
			}

			err := validate.Struct(dto)

			if err == nil {
				stockDTOs = append(stockDTOs, &dto)
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		var validatedStockDTOs []mongo.WriteModel

		for _, dto := range stockDTOs {
			err := validate.Struct(dto)
			if err == nil {
				openPrice, errOpenPrice := strconv.Atoi(dto.Open)
				closePrice, errClosePrice := strconv.Atoi(dto.Close)
				highPrice, errHighPrice := strconv.Atoi(dto.High)
				lowPrice, errLowPrice := strconv.Atoi(dto.Low)

				if errOpenPrice == nil && errClosePrice == nil && errHighPrice == nil && errLowPrice == nil {
					filter := bson.M{"code": dto.Code}
					update := bson.M{
						"$set": bson.M{"open": openPrice, "close": closePrice, "high": highPrice, "low": lowPrice},
					}
					validatedStockDTOs = append(validatedStockDTOs, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
				}
			}
		}

		if len(validatedStockDTOs) > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, errSave := models.Db.GetRepo("StockRepo").Collection().BulkWrite(ctx, validatedStockDTOs)
			if errSave != nil {
				logger.Sugar().Info("Error Saving scrapeStockPriceSummary", errSave)
			} else {
				logger.Sugar().Infof("Finish Scrapping scrapeStockPriceSummary, upsert %d stocks", len(validatedStockDTOs))
			}
		}

		// continue scrapping if current result is not empty
		if len(stockDTOs) > 0 {
			stockDTOs = nil // clear slice before continuing
			hasToScrape <- true
		} else {
			hasToScrape <- false
		}
	})

	hasToScrape <- true
	i := 1
	for {
		if <-hasToScrape {
			url := fmt.Sprintf("https://www.duniainvestasi.com/bei/prices/daily/%s/page:%d", currentTime.Format("20060102"), i)
			c.Visit(url)
			i++
		} else {
			break
		}
	}
}
