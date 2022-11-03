package main

import (
	"context"
	"toks/config/database"
	"toks/handler"

	"github.com/gofiber/fiber/v2"

	"log"
)

func main() {
	database.Connect()
	database.CreateUsersTableWithAdmin()
	app := fiber.New()

	// Users //
	users := app.Group("/users", authUser)
	users.Post("", handler.CreateUser)
	users.Delete("/:id", handler.DeleteUser)
	users.Put("/:id", handler.UpdateUser)

	// Tokens //
	tokens := app.Group("/tokens", authUser)
	tokens.Post("/generate/:amount", handler.GenerateTokens)
	tokens.Post("/send/:to_user/:amount", handler.SendTokens)

	log.Fatal(app.Listen(":8080"))
}

func authUser(c *fiber.Ctx) error {
	var admin bool
	var id string
	database.DB.QueryRow(context.Background(), "select id, admin from users where nickname = $1 and password = $2",
		c.Get("Nickname"), c.Get("Password"),
	).Scan(&id, &admin)

	if id != "" {
		c.Locals("userId", id)
		c.Locals("userAdmin", admin)
		return c.Next()
	} else {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
}
