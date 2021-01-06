package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/ahui2016/uglynotes/model"
	"github.com/gofiber/fiber/v2"
)

// getFormValue gets the c.FormValue(key), trims its spaces,
// and checks if it is empty or not.
func getFormValue(c *fiber.Ctx, key string) (string, error) {
	value := strings.TrimSpace(c.FormValue(key))
	if value == "" {
		return "", errors.New(key + " is empty")
	}
	return value, nil
}

func getID(c *fiber.Ctx) (string, error) {
	return getFormValue(c, "id")
}

func getNoteType(c *fiber.Ctx) (NoteType, error) {
	noteTypeString, err := getFormValue(c, "note-type")
	noteType := model.NewNoteType(noteTypeString)
	return noteType, err
}

func getTags(c *fiber.Ctx) ([]string, error) {
	tagsString, err := getFormValue(c, "tags")
	if err != nil {
		return nil, err
	}
	var tags []string
	err = json.Unmarshal([]byte(tagsString), &tags)
	return tags, err
}

func getProtected(c *fiber.Ctx) (protected bool, err error) {
	s, err := getFormValue(c, "protected")
	if err != nil {
		return
	}
	if s == "true" {
		protected = true
	}
	return
}

func getParams(c *fiber.Ctx, key string) (string, error) {
	return url.QueryUnescape(c.Params(key))
}

func getTagGroup(c *fiber.Ctx) ([]string, error) {
	tagsString, err := getParams(c, "tags")
	return strings.Split(tagsString, " "), err
}
