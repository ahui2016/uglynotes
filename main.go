package main

import (
	"log"

	"github.com/ahui2016/uglynotes/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	defer db.Close()

	app := fiber.New(fiber.Config{
		BodyLimit:    maxBodySize,
		Concurrency:  10,
		ErrorHandler: errorHandler,
	})

	app.Use(responseNoCache)
	app.Use(limiter.New(limiter.Config{
		Max: 300,
	}))

	app.Static("/public", "./public")

	// app.Use(favicon.New(favicon.Config{File: "public/icons/favicon.ico"}))

	app.Use("/static", checkLoginHTML)
	app.Static("/static", "./static")

	app.Get("/", redirectToHome)
	app.Use("/home", checkLoginHTML)
	app.Get("/home", homePage)
	app.Post("/login", loginHandler)

	app.Post("/new-note", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": model.RandomID()})
	})

	log.Fatal(app.Listen(defaultAddress))
}
