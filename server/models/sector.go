package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Sector struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name string             `bson:"name,omitempty" json:"name"`
}

type SectorRepo struct {
	collection *mongo.Collection
}

func (r *SectorRepo) RepoName() string {
	return "SectorRepo"
}

func (r *SectorRepo) Collection() *mongo.Collection {
	return r.collection
}

func (r *SectorRepo) CreateSectors(documents []interface{}) ([]Sector, error) {
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.InsertMany(context, documents)
	if err != nil {
		err = fmt.Errorf("Create Sectors Error : %w", err)
		return nil, err
	}

	var sectors []Sector
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: result.InsertedIDs}}}}
	cursor, err := r.collection.Find(context, filter)
	if err != nil {
		err = fmt.Errorf("Create Sectors Error : %w", err)
		return nil, err
	}

	if err = cursor.All(context, &sectors); err != nil {
		err = fmt.Errorf("Create Sectors Error : %w", err)
		return nil, err
	}

	return sectors, nil
}
