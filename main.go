package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

type User struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

var db *pgxpool.Pool

func main() {
	connectDatabase()
	createUsersTableWithAdmin()
	app := fiber.New()

	// users //
	users := app.Group("/users", authUser)
	users.Post("", createUser)
	users.Delete("/:id", deleteUser)
	users.Put("/:id", updateUser)
	// users.Get("", getUsersNicknames)

	// tokens generate/send //
	tokens := app.Group("/tokens", authUser)
	tokens.Post("/generate/:amount", generateTokens)
	tokens.Post("/send/:to_user/:amount", sendTokens)

	log.Fatal(app.Listen(":8080"))
}

func authUser(c *fiber.Ctx) error {
	var admin bool
	var id string
	db.QueryRow(context.Background(), "select id, admin from users where nickname = $1 and password = $2",
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

func authAdmin(c *fiber.Ctx) error {
	if c.Locals("userAdmin") == "true" {
		return nil
	} else {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
}

// func getUsersNicknames(c *fiber.Ctx) error {
// 	arr := make([]string, 0)
// 	rows, _ := db.Query(context.Background(), "select nickname from users where admin is null;")
// 	for rows.Next() {
// 		var nickname string
// 		err := rows.Scan(&nickname)
// 		if err != nil {
// 			return err
// 		}
// 		arr = append(arr, nickname)
// 	}
// 	return c.JSON(fiber.Map{"message": "Users list", "list": arr, "error": rows.Err()})
// }

func createUser(c *fiber.Ctx) error {
	authAdmin(c)

	u := new(User)

	if err := c.BodyParser(u); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	if _, err := db.Exec(context.Background(), "insert into users(nickname, password) values ($1, $2)", u.Nickname, u.Password); err == nil {
		return c.Status(201).JSON(fiber.Map{"message": "User Is Created"})
	} else {
		return c.JSON(fiber.Map{"error": err.Error()})
	}
}

func updateUser(c *fiber.Ctx) error {
	u := new(User)

	if err := c.BodyParser(u); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	if c.Locals("userId") == c.Params("id") {
		if _, err := db.Exec(context.Background(), "UPDATE users SET nickname=$1, password=$2 WHERE id=$3", u.Nickname, u.Password, c.Locals("userId")); err == nil {
			return c.Status(201).JSON(u)
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	} else if c.Locals("userAdmin") == true && c.Locals("userId") != c.Params("id") {
		if _, err := db.Exec(context.Background(), "UPDATE users SET nickname=$1, password=$2 WHERE id=$3", u.Nickname, u.Password, c.Params("id")); err == nil {
			return c.Status(201).JSON(u)
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	}
	return nil
}

func deleteUser(c *fiber.Ctx) error {
	authAdmin(c)
	if c.Locals("userId") != c.Params("id") {
		if _, err := db.Exec(context.Background(), "delete from users where id=$1", c.Params("id")); err == nil {
			return c.JSON(fiber.Map{"message": "User deleted"})
		} else {
			return c.JSON(fiber.Map{"error": err.Error()})
		}
	}
	return c.JSON(fiber.Map{"message": "You can't do this"})
}

func generateTokens(c *fiber.Ctx) error {
	authAdmin(c)
	query := "UPDATE users SET balance=$1 WHERE id=$2"

	if _, err := db.Exec(context.Background(), query, c.Params("amount"), c.Locals("userId")); err == nil {
		return c.JSON(fiber.Map{"message": "Generated tokens"})
	} else {
		return c.JSON(fiber.Map{"message": err.Error()})
	}
}

// func sendTokens(c *fiber.Ctx) error {
// 	lock := "UPDATE users SET balance = balance - $1 WHERE id = $2;"
// 	send := "UPDATE users SET balance = balance + $1 WHERE id = $2;"
// 	// amount, from_user, to_user
// 	db.Exec(context.Background(), "BEGIN;")
// 	if _, err := db.Exec(context.Background(), lock, c.Params("amount"), c.Locals("userId")); err == nil {
// 		if _, err := db.Exec(context.Background(), send, c.Params("amount"), c.Params("to_user")); err == nil {
// 			db.Exec(context.Background(), "COMMIT;")
// 			return c.JSON(fiber.Map{"message": "Send tokens"})
// 		} else {
// 			db.Exec(context.Background(), "ROLLBACK;")
// 			return c.JSON(fiber.Map{"error": err.Error()})
// 		}
// 	} else {
// 		db.Exec(context.Background(), "ROLLBACK;")
// 		return c.JSON(fiber.Map{"error": err.Error()})
// 	}
// }

func sendTokens(c *fiber.Ctx) error {
	send := "UPDATE users SET balance = balance - $1 WHERE id = $2;"
	get := "UPDATE users SET balance = balance + $1 WHERE id = $2;"

	tx, err := db.Begin(context.Background())
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

func connectDatabase() {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL")) //'postgres://user:pass@localhost/dbname'
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL:", err)
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}
}

func createUsersTableWithAdmin() {
	query := `CREATE TABLE IF NOT EXISTS users (
		id serial PRIMARY KEY,

		nickname VARCHAR ( 50 ) UNIQUE NOT NULL CHECK (length(nickname) > 3),
		password VARCHAR ( 50 ) NOT NULL,
		balance bigint NOT NULL DEFAULT 0 CHECK (balance >= 0),
		admin bool,

		CONSTRAINT admin_true_or_null CHECK (admin),
		CONSTRAINT admin_only_1_true UNIQUE (admin)
	);
	
	INSERT INTO users(nickname, password, admin, balance) 
		VALUES ('admin', 'admin', true, 999999)
		ON CONFLICT DO NOTHING
	;`

	if _, err := db.Exec(context.Background(), query); err != nil {
		panic(err)
	}
}
