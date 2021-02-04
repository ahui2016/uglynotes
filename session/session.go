package session

import "github.com/gofiber/fiber/v2"

type Manager struct {
	store  map[string]bool
	name   string
	maxAge int
}

func NewManager(name string, maxAge int) *Manager {
	return &Manager{
		store:  make(map[string]bool),
		name:   name,
		maxAge: maxAge,
	}
}

func (manager *Manager) NewSession(sid string) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = manager.name
	cookie.Value = sid
	cookie.Path = "/" // important!
	cookie.MaxAge = manager.maxAge
	cookie.HTTPOnly = true
	return cookie
}

func (manager *Manager) Add(c *fiber.Ctx, sid string) {
	session := manager.NewSession(sid)
	c.Cookie(session)
	manager.store[sid] = true
}

func (manager *Manager) Check(c *fiber.Ctx) bool {
	cookieValue := c.Cookies(manager.name)
	return manager.store[cookieValue]
}

func (manager *Manager) Delete(c *fiber.Ctx) {
	cookieValue := c.Cookies(manager.name)
	manager.store[cookieValue] = false
	cookie := new(fiber.Cookie)
	cookie.Name = manager.name
	cookie.Value = ""
	cookie.Path = "/" // important!
	cookie.MaxAge = -1
	cookie.HTTPOnly = true
	c.Cookie(cookie)
}

