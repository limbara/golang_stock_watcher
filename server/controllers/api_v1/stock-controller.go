package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/limbara/stock-watcher/models"
	"go.mongodb.org/mongo-driver/bson"
)

func GetStocks(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()

	stockRepo := models.GetDB().GetRepo("StockRepo").(*models.StockRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}

	if searches, exists := urlParams["search"]; exists {
		searchValue := searches[0]

		if searchValue != "" {
			filter["$or"] = bson.A{
				bson.M{"code": bson.M{"$regex": fmt.Sprintf("%s", searchValue), "$options": "i"}},
				bson.M{"name": bson.M{"$regex": fmt.Sprintf("%s", searchValue), "$options": "i"}},
			}
		}
	}

	cursor, err := stockRepo.Collection().Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	var stocks []models.Stock
	if err = cursor.All(ctx, &stocks); err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}
