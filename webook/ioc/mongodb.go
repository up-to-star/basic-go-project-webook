package ioc

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitMongoDB() *mongo.Database {
	ops := options.Client().ApplyURI("mongodb://root:example@localhost:27017")
	client, err := mongo.Connect(ops)
	if err != nil {
		panic(err)
	}
	db := client.Database("webook")
	return db
}
