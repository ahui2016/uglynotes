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

	htmlPage := app.Group("/html", checkLoginHTML)
	htmlPage.Get("/note", notePage)
	htmlPage.Get("/note/new", noteNewPage)
	htmlPage.Get("/note/edit", noteEditPage)

	api := app.Group("/api", checkLoginJSON)
	api.Get("/notes/all", allNotesHandler)
	api.Post("/note", getNoteHandler)
	api.Post("/note/new", newNoteHandler)
	api.Post("/note/delete", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	api.Post("/note/type/update", changeType)
	api.Post("/note/tags/update", noteTagsUpdate)
	api.Post("/note/contents/update", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"id": model.RandomID()}) // history_id
	})
	api.Get("/tag/:name", func(c *fiber.Ctx) error {
		tag, err := db.GetTag(c.Params("name"))
		if err != nil {
			return err
		}
		return c.JSON(tag)
	})

	log.Fatal(app.Listen(defaultAddress))
}
