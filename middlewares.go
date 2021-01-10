package main

import (
	"errors"

	"github.com/ahui2016/uglynotes/settings"
	"github.com/gofiber/fiber/v2"
)

func jsonMessage(c *fiber.Ctx, msg string) error {
	return c.Status(200).JSON(fiber.Map{"message": msg})
}

func jsonMsgOK(c *fiber.Ctx) error {
	return jsonMessage(c, "OK")
}

func jsonError(c *fiber.Ctx, msg string, status int) error {
	return c.Status(status).JSON(fiber.Map{"message": msg})
}

func responseNoCache(c *fiber.Ctx) error {
	c.Response().Header.Set(
		fiber.HeaderCacheControl,
		"no-store, no-cache",
	)
	return c.Next()
}

func checkLoginHTML(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		if err := checkPasswordTry(c); err != nil {
			return err
		}
		return c.SendFile("./public/login.html")
	}
	return c.Next()
}

func checkLoginJSON(c *fiber.Ctx) error {
	if isLoggedOut(c) {
		return jsonError(c, "Require Login", fiber.StatusUnauthorized)
	}
	return c.Next()
}

func isLoggedIn(c *fiber.Ctx) bool {
	return db.SessionCheck(c)
}

func isLoggedOut(c *fiber.Ctx) bool {
	return !isLoggedIn(c)
}

func checkPasswordTry(c *fiber.Ctx) error {
	if passwordTry >= settings.PasswordMaxTry {
		_ = db.Close()
		msg := "No more try. Input wrong password too many times."
		return errors.New(msg)
	}
	return nil
}
