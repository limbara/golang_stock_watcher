package main

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterCrons() {
	logger := utils.Logger()

	c := cron.New()

	// running scrapeStocksCodeAndName every At 00:00 on Sunday
	c.AddFunc("0 0 * * 0", scrapeStocksCodeAndName)

	// running scrapeStockPriceSummary every monday to saturday on 5pm
	c.AddFunc("0 17 * * 1-6", scrapeStockPriceSummary)

	// delete old log file
	c.AddFunc("0 0 * * *", deleteOlderLogFile)

	logger.Sugar().Info("Running Cron")
	c.Start()
}

type StockCodeAndNameDTO struct {
	Code string `validate:"required,uppercase,alpha"`
	Name string `validate:"required"`
}

func scrapeStocksCodeAndName() {
	logger := utils.Logger()

	urls := make([]string, 60)
	// we're just going to make our life easier by generating 60 page url
	for i := 1; i <= 60; i++ {
		urls[i-1] = fmt.Sprintf("https://www.idnfinancials.com/id/company/page/%d", i)
	}

	q, err := queue.New(
		2, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)
	if err != nil {
		logger.Sugar().Panic(fmt.Errorf("Error init queue in scrapeStocksCodeAndName : %w", err))
	}

	c := colly.NewCollector(
		colly.AllowedDomains("www.idnfinancials.com"),
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

	c.OnHTML("#table-companies div.table-body .table-row", func(e *colly.HTMLElement) {
		filteredChildren := e.DOM.Children().Filter(".tc-company")

		hasCode := filteredChildren.Children().Eq(0).Text() != ""
		hasName := filteredChildren.Children().Eq(1).Text() != ""

		if filteredChildren.Children().Length() >= 2 && hasCode && hasName {
			dto := StockCodeAndNameDTO{
				Code: filteredChildren.Children().Eq(0).Text(),
				Name: filteredChildren.Children().Eq(1).Text(),
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

	if len(validatedStockDTOs) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, errSave := models.GetDB().GetRepo("StockRepo").Collection().BulkWrite(ctx, validatedStockDTOs)
		if errSave != nil {
			logger.Sugar().Info("Error Saving scrapeStocksCodeAndName", errSave)
		} else {
			logger.Sugar().Infof("Finish Scrapping scrapeStocksCodeAndName, upsert %d stocks", len(validatedStockDTOs))
		}
	}
}

type StockSummary struct {
	Code  string `validate:"required,uppercase,alpha"` // also protect from warant stock ALDO-W
	Open  string `validate:"required,numeric"`
	Close string `validate:"required,numeric"`
	High  string `validate:"required,numeric"`
	Low   string `validate:"required,numeric"`
}

// scrape non-live Stocks last updated price
func scrapeStockPriceSummary() {
	logger := utils.Logger()

	// helper function to scrape stock price summary, returns true if there's anything scraped
	scrape := func(dateTime time.Time) bool {
		c := colly.NewCollector(
			colly.AllowedDomains("www.infovesta.com"),
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

		done := make(chan bool, 1)
		validate := validator.New()
		var stockDTOs []*StockSummary

		c.OnHTML("div table tbody tr", func(e *colly.HTMLElement) {
			filteredChildren := e.DOM.Children().Filter("td")

			if filteredChildren.Eq(0).Text() == "GOTO" {
				println(filteredChildren.Eq(0).Text())
			}

			hasCodeColumn := filteredChildren.Eq(0).Text() != ""
			hasOpenColumn := filteredChildren.Eq(1).Text() != ""
			hasHighColumn := filteredChildren.Eq(2).Text() != ""
			hasLowColumn := filteredChildren.Eq(3).Text() != ""
			hasCloseColumn := filteredChildren.Eq(4).Text() != ""

			if hasCodeColumn && hasOpenColumn && hasHighColumn && hasLowColumn && hasCloseColumn {
				dto := StockSummary{
					Code:  filteredChildren.Eq(0).Text(),
					Open:  strings.ReplaceAll(filteredChildren.Eq(1).Text(), ",", ""),
					Close: strings.ReplaceAll(filteredChildren.Eq(4).Text(), ",", ""),
					High:  strings.ReplaceAll(filteredChildren.Eq(2).Text(), ",", ""),
					Low:   strings.ReplaceAll(filteredChildren.Eq(3).Text(), ",", ""),
				}

				err := validate.Struct(dto)

				if err == nil {
					stockDTOs = append(stockDTOs, &dto)
				}
			}
		})

		c.OnScraped(func(r *colly.Response) {
			done <- true
		})

		c.Visit(fmt.Sprintf("http://www.infovesta.com/index/stock/ALL_=%d", dateTime.Unix()))

		<-done

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
						"$set":         bson.M{"open": openPrice, "close": closePrice, "high": highPrice, "low": lowPrice},
						"$setOnInsert": bson.M{"code": dto.Code},
					}
					validatedStockDTOs = append(validatedStockDTOs, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true))
				}
			}
		}

		if len(validatedStockDTOs) > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, errSave := models.GetDB().GetRepo("StockRepo").Collection().BulkWrite(ctx, validatedStockDTOs)
			if errSave != nil {
				logger.Sugar().Info("Error Saving scrapeStockPriceSummary", errSave)
			} else {
				logger.Sugar().Infof("Finish Scrapping scrapeStockPriceSummary, upsert %d stocks", len(validatedStockDTOs))
			}
		}

		if len(stockDTOs) > 0 {
			return true
		} else {
			return false
		}
	}

	// get the last 7 days calculated from today inclusive today skipping if day is Sunday or Saturday
	// we're going to keep scraping if the current scraping date is not returning any result for the last 7 days scrapeDateTimes
	scrapeDateTimes := []time.Time{}
	i := 0
	for len(scrapeDateTimes) <= 7 {
		dateTime := time.Now().AddDate(0, 0, -1*i)
		for dateTime.Weekday() == time.Sunday || dateTime.Weekday() == time.Saturday {
			i++
			dateTime = time.Now().AddDate(0, 0, -1*i)
		}
		scrapeDateTimes = append(scrapeDateTimes, dateTime)
		i++
	}

	for _, scrapeDateTime := range scrapeDateTimes {
		if scrape(scrapeDateTime) {
			logger.Sugar().Infof("Scrapped scrapeStockPriceSummary on %s, stop scrapping further", scrapeDateTime.String())
			break
		}
	}
}

func deleteOlderLogFile() {
	logger := utils.Logger()

	logPath := viper.Get("LOG_PATH")

	nDaysAgo := time.Now().AddDate(0, 0, -7)
	filePath := fmt.Sprintf("%s/%s.log", logPath, nDaysAgo.Format("2006-01-02"))

	if err := os.Remove(filePath); err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			logger.Sugar().Infof("deleteOlderLogFile error : %v", pathError.Error())
		} else {
			logger.Sugar().Infof("deleteOlderLogFile error : %v", err)
		}
	} else {
		logger.Sugar().Infof("deleteOlderLogFile %s deleted", filePath)
	}
}
