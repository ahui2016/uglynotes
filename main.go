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

	app.Get("/", redirectToHome)
	app.Use("/home", checkLoginHTML)
	app.Get("/home", homePage)
	app.Get("/login", loginPage)
	app.Post("/login", loginHandler)
	app.Get("/check", checkLogin)
	app.Get("/converter", converterPage)

	app.Get("/reset-all-tags", func(c *fiber.Ctx) error {
		return db.ResetAllTags()
	})

	htmlPage := app.Group("/html", checkLoginHTML)
	htmlPage.Get("/note", notePage)
	htmlPage.Get("/note/new", noteNewPage)
	htmlPage.Get("/note/new2", noteNewPage2)
	htmlPage.Get("/note/edit", noteEditPage)
	htmlPage.Get("/history", historyPage)
	htmlPage.Get("/history2", historyPage2)
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

	api.Delete("/note/:id/history", deleteNoteHistories)

	api.Get("/tag/all", getAllTags)
	api.Get("/tag/all-by-date", allTagsByDate)
	api.Get("/tag/:name/notes", getNotesByTag)
	api.Put("/tag", renameTag)
	api.Delete("/tag/:name", deleteTag)
	api.Get("/tag/group/all", allTagGroups)
	api.Post("/tag/group", addTagGroup)
	api.Delete("/tag/group/:id", deleteTagGroup)
	api.Put("/tag/group/protected", setTagGroupProtected)

	api.Get("/search/tags/:tags", searchTagGroup)
	api.Get("/search/title/:pattern", searchTitle)

	// 导出全部笔记
	api.Get("/export", exportAllNotes)

	log.Fatal(app.Listen(config.Address))
}
