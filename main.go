package main

import (
	"encoding/json"
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

	htmlPage := app.Group("/html", checkLoginHTML)
	htmlPage.Get("/note", notePage)
	htmlPage.Get("/note/edit", noteEditPage)

	api := app.Group("/api", checkLoginJSON)
	api.Get("/notes/all", allNotesHandler)
	api.Post("/note", getNoteHandler)
	api.Post("/note/new", newNoteHandler)
	api.Post("/note/delete", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	api.Post("/note/type/update", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	api.Post("/note/tags/update", func(c *fiber.Ctx) error {
		var tags []string
		err := json.Unmarshal([]byte(c.FormValue("tags")), &tags)
		log.Print(tags)
		return err
	})
	api.Post("/note/contents/update", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": model.RandomID()}) // history_id
	})

	log.Fatal(app.Listen(defaultAddress))
}
