package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"toks/config/database"
)

func SendTokens(c *fiber.Ctx) error {
	send := "UPDATE users SET balance = balance - $1 WHERE id = $2;"
	get := "UPDATE users SET balance = balance + $1 WHERE id = $2;"

	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(context.Background(), send, c.Params("amount"), c.Locals("userId")); err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := tx.Exec(context.Background(), get, c.Params("amount"), c.Params("to_user")); err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Send tokens"})
}

func GenerateTokens(c *fiber.Ctx) error {
	authAdmin(c)
	query := "UPDATE users SET balance=$1 WHERE id=$2"

	if _, err := database.DB.Exec(context.Background(), query, c.Params("amount"), c.Locals("userId")); err == nil {
		return c.JSON(fiber.Map{"message": "Generated tokens"})
	} else {
		return c.JSON(fiber.Map{"message": err.Error()})
	}
}
