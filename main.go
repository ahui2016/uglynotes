package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	defer db.Close()

	app := fiber.New(fiber.Config{
		BodyLimit:    config.MaxBodySize,
		Concurrency:  10,
		ErrorHandler: errorHandler,
	})

	// app.Use(responseNoCache)
	app.Use(limiter.New(limiter.Config{
		Max: 300,
	}))

	app.Static("/public", "./public")

	// app.Use(favicon.New(favicon.Config{File: "public/icons/favicon.ico"}))

	app.Use("/static", checkLoginHTML)
	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/home") })
	app.Use("/home", checkLoginHTML)
	app.Get("/home", homePage)
	app.Get("/login", loginPage)
	app.Post("/login", loginHandler)
	app.Get("/check", checkLogin)
	app.Get("/converter", converterPage)

	app.Get("/import-notes", importNotes)

	htmlPage := app.Group("/html", checkLoginHTML)
	htmlPage.Get("/index", indexPage)
	htmlPage.Get("/note", notePage)
	htmlPage.Get("/note/new", noteNewPage)
	htmlPage.Get("/note/edit", noteEditPage)
	htmlPage.Get("/history", historyPage)
	htmlPage.Get("/note/history", noteHistoryPage)
	htmlPage.Get("/tag", tagPage)
	htmlPage.Get("/tags", tagsPage)
	htmlPage.Get("/search", searchPage)
	htmlPage.Get("/tag/groups", tagGroupsPage)

	api := app.Group("/api", checkLoginJSON)
	api.Get("/note/all", getAllNotes)
	api.Get("/note/deleted", getDeletedNotes)
	api.Get("/note/all/size", notesSizeHandler)

	api.Post("/note", newNoteHandler)
	api.Get("/note/:id", getNoteHandler)
	api.Patch("/note/:id", patchNoteHandler)
	api.Put("/note/:id/deleted", setNoteDeleted)
	api.Delete("/note/:id", deleteNoteForever)
	api.Put("/note/:id/type", changeType)
	api.Put("/note/:id/tags", updateNoteTags)

	api.Get("/tag/all/:sortby", allTagsSorted)
	api.Get("/tag/:id/notes", getNotesByTag)
	api.Get("/tag/name/:name", getTagByName)
	api.Get("/tag/:id", getTagByID)
	api.Put("/tag/:id", renameTag)
	api.Delete("/tag/:id", deleteTag)
	api.Get("/tag/group/all", allTagGroups)
	api.Post("/tag/group", addTagGroup)
	api.Delete("/tag/group/:id", deleteTagGroup)
	api.Put("/tag/group/:id/protected", setTagGroupProtected)

	api.Get("/search/tags/:tags", searchTagGroup)
	api.Get("/search/title/:pattern", searchTitle)

	api.Get("/backup/db", downloadDatabase)
	api.Get("/backup/export", exportAllNotes)
	api.Get("/backup/json", downloadDatabaseJSON)

	log.Fatal(app.Listen(config.Address))
}
