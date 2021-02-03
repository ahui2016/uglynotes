package database

import "github.com/gofiber/fiber/v2"

// SessionCheck .
func (db *DB) SessionCheck(c *fiber.Ctx) bool {
	sess, err := db.Sess.Get(c)

	if err != nil || sess.Get(cookieName) == nil {
		return false
	}
	return sess.Get(cookieName).(bool)
}

// SessionSet .
func (db *DB) SessionSet(c *fiber.Ctx) error {
	sess, err := db.Sess.Get(c)
	if err != nil {
		return err
	}
	sess.Set(cookieName, true)
	return sess.Save()
}
