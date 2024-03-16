package middleware

import "github.com/gofiber/fiber/v2"

func CustomMiddleware404() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil && err == fiber.ErrMethodNotAllowed {
			err = fiber.ErrNotFound
		}

		return err
	}
}
