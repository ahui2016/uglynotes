package main

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ahui2016/uglynotes/model"
	"github.com/gofiber/fiber/v2"
)

type (
	Note = model.Note
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

func noteEditPage(c *fiber.Ctx) error {
	return c.SendFile("./static/note-edit.html")
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
	note, err := db.GetByID(c.FormValue("id"))
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
	return c.JSON(note.ID)
}

func createNote(c *fiber.Ctx) (*Note, error) {
	noteType := model.NewNoteType(c.FormValue("note-type"))
	contents := strings.TrimSpace(c.FormValue("contents"))
	if contents == "" {
		return nil, errors.New("contents is empty")
	}

	note := db.NewNote(noteType)
	if err := note.SetContents(contents); err != nil {
		return nil, err
	}

	var tags []string
	if err := json.Unmarshal([]byte(c.FormValue("tags")), &tags); err != nil {
		return nil, err
	}
	note.Tags = tags
	return note, nil
}
