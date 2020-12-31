package main

import (
	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/util"
	"github.com/gofiber/fiber/v2"
)

type (
	Note     = model.Note
	NoteType = model.NoteType
)

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	err = c.Status(code).JSON(fiber.Map{"message": err.Error()})
	if err != nil {
		// In case the c.JSON fails
		return c.Status(500).SendString("Internal Server Error")
	}
	return nil
}

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/home")
}

func homePage(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
}

func notePage(c *fiber.Ctx) error {
	return c.SendFile("./static/note.html")
}

func noteNewPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit.html")
}

func noteEditPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit.html")
}

func historyPage(c *fiber.Ctx) error {
	return c.SendFile("./static/history.html")
}

func loginHandler(c *fiber.Ctx) error {
	if isLoggedIn(c) {
		return jsonMessage(c, "already logged in")
	}

	if c.FormValue("password") != defaultPassword {
		passwordTry++
		if err := checkPasswordTry(c); err != nil {
			return err
		}
		return jsonError(c, "Wrong Password", 400)
	}

	passwordTry = 0
	return db.SessionSet(c)
}

func allNotesHandler(c *fiber.Ctx) error {
	notes, err := db.AllNotes()
	if err != nil {
		return nil
	}
	return c.JSON(notes)
}

func getNoteHandler(c *fiber.Ctx) error {
	note, err := db.GetByID(c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(note)
}

func newNoteHandler(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	note, err := createNote(c)
	if err != nil {
		return jsonError(c, err.Error(), 400)
	}
	if err := db.Insert(note); err != nil {
		return err
	}
	return jsonMessage(c, note.ID)
}

func createNote(c *fiber.Ctx) (*Note, error) {
	noteType, err1 := getNoteType(c)
	contents, err2 := getFormValue(c, "contents")
	tags, err3 := getTags(c)

	if err := util.WrapErrors(err1, err2, err3); err != nil {
		return nil, err
	}

	note := db.NewNote(noteType)
	if err := note.SetContents(contents); err != nil {
		return nil, err
	}
	note.Tags = tags
	return note, nil
}

func changeType(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err1 := getID(c)
	noteType, err2 := getNoteType(c)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return db.ChangeType(id, noteType)
}

func updateNoteTags(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err1 := getID(c)
	tags, err2 := getTags(c)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return db.UpdateTags(id, tags)
}

func updateNoteContents(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	id, err1 := getID(c)
	contents, err2 := getFormValue(c, "contents")
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	historyID, err := db.UpdateNoteContents(id, contents)
	if err != nil {
		return err
	}
	return jsonMessage(c, historyID)
}

func notesSizeHandler(c *fiber.Ctx) error {
	size, err := db.GetTotalSize()
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"totalSize": size,
		"capacity":  databaseCapacity,
	})
}

func getHistoryHandler(c *fiber.Ctx) error {
	history, err := db.GetHistory(c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(history)
}

func setProtected(c *fiber.Ctx) error {
	db.Lock()
	defer db.Unlock()

	historyID, err1 := getID(c)
	protected, err2 := getProtected(c)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return db.SetProtected(historyID, protected)
}
