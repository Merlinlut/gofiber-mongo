package config

import (
	"context"
	"gofiber-mongo/route"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewApp() *fiber.App {
	app := fiber.New()

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	db := client.Database("alumni_db")

	route.RegisterRoutes(app, db)

	return app
}
