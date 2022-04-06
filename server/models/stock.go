package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Stock struct {
	ID    primitive.ObjectID `bson:"_id" json:"_id"`
	Code  string             `bson:"code" json:"code"`
	Name  string             `bson:"name,omitempty" json:"name"`
	Open  int                `bson:"open,omitempty" json:"open"`
	Close int                `bson:"close,omitempty" json:"close"`
	High  int                `bson:"high,omitempty" json:"high"`
	Low   int                `bson:"low,omitempty" json:"low"`
}

type StockRepo struct {
	collection *mongo.Collection
}

func (r *StockRepo) RepoName() string {
	return "StockRepo"
}

func (r *StockRepo) Collection() *mongo.Collection {
	return r.collection
}
