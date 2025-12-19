package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewFiber() *fiber.App {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	return app
}