package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/limbara/stock-watcher/customerrors"
	"github.com/limbara/stock-watcher/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSectors(w http.ResponseWriter, r *http.Request) {
	sectorRepo := models.Db.GetRepo("SectorRepo").(*models.SectorRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := sectorRepo.Collection().Find(ctx, bson.D{})
	if err != nil {
		panic(err)
	}
	var sectors []models.Sector
	if err = cursor.All(ctx, &sectors); err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sectors)
}

func GetSector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		responseNotFound := customerrors.ResponseNotFound.SetMessage("Sector Not Found!")
		panic(responseNotFound)
	}

	sectorRepo := models.Db.GetRepo("SectorRepo").(*models.SectorRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sector models.Sector
	if err := sectorRepo.Collection().FindOne(ctx, bson.D{{Key: "_id", Value: objectId}}).Decode(&sector); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			responseNotFound := customerrors.ResponseNotFound.SetMessage("Sector Not Found!")
			panic(responseNotFound)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sector)
}
