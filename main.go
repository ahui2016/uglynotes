package main

import (
	"log"

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
	htmlPage.Get("/history", historyPage)
	htmlPage.Get("/note/history", noteHistoryPage)
	htmlPage.Get("/tag", tagPage)

	api := app.Group("/api", checkLoginJSON)
	api.Get("/notes/all", allNotesHandler)
	api.Get("/notes/size", notesSizeHandler)

	api.Get("/note/:id", getNoteHandler)
	api.Post("/note", newNoteHandler)
	api.Delete("/note/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	api.Put("/note/type", changeType)
	api.Put("/note/tags", updateNoteTags)
	api.Put("/note/contents", updateNoteContents)

	api.Get("/note/:id/history", noteHistory)
	api.Get("/history/:id", getHistoryHandler)
	api.Put("/history/protected", setProtected)

	api.Get("/tag/:name/notes", getNotesByTag)
	api.Put("/tag", renameTag)

	log.Fatal(app.Listen(defaultAddress))
}
