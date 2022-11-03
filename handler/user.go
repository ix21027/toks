package handler

import (
	"context"
	"toks/config/database"

	"github.com/gofiber/fiber/v2"
)

type User struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func authAdmin(c *fiber.Ctx) error {
	if c.Locals("userAdmin") == "true" {
		return nil
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
}

func CreateUser(c *fiber.Ctx) error {
	authAdmin(c)

	u := new(User)

	if err := c.BodyParser(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := database.DB.Exec(context.Background(), "insert into users(nickname, password) values ($1, $2)", u.Nickname, u.Password); err == nil {
		return c.Status(201).JSON(fiber.Map{"message": "User Is Created"})
	} else {
		return c.JSON(fiber.Map{"error": err.Error()})
	}
}

func UpdateUser(c *fiber.Ctx) error {
	u := new(User)

	if err := c.BodyParser(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if c.Locals("userId") == c.Params("id") {
		if _, err := database.DB.Exec(context.Background(), "UPDATE users SET nickname=$1, password=$2 WHERE id=$3", u.Nickname, u.Password, c.Locals("userId")); err == nil {
			return c.Status(201).JSON(u)
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	} else if c.Locals("userAdmin") == true && c.Locals("userId") != c.Params("id") {
		if _, err := database.DB.Exec(context.Background(), "UPDATE users SET nickname=$1, password=$2 WHERE id=$3", u.Nickname, u.Password, c.Params("id")); err == nil {
			return c.Status(201).JSON(u)
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	}
	return nil
}

func DeleteUser(c *fiber.Ctx) error {
	authAdmin(c)
	if c.Locals("userId") != c.Params("id") {
		if _, err := database.DB.Exec(context.Background(), "delete from users where id=$1", c.Params("id")); err == nil {
			return c.JSON(fiber.Map{"message": "User deleted"})
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "You can't do this"})
}
