package main

import (
	"log"
	"UASBE/config"
	"UASBE/database"
	"UASBE/routes"

	_ "UASBE/docs"

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

var swaggerWrapper = fiberSwagger.WrapHandler

func main() {
	config.LoadConfig()
	cfg := config.AppConfig

	app := config.NewFiber()

	// init db
	dbpool := database.NewPostgresDB(cfg) // harus *pgxpool.Pool
	mongoClient := database.ConnectMongoDB(cfg.MongoURI)
	mongoColl := database.GetCollection(mongoClient, cfg.MongoDB, "achievements")

	routes.SetupRoutes(app, dbpool, mongoColl) // panggil SetupRoutes

	// Swagger route
	app.Get("/swagger/*", swaggerWrapper)

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", cfg.AppPort)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
