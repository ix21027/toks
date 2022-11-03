package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"toks/config"
)

// DB is the underlying database connection
var DB *pgxpool.Pool

func Connect() {
	poolConfig, err := pgxpool.ParseConfig(config.DATABASE_URL)
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL:", err)
	}

	DB, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}

	fmt.Println("[DATABASE]::CONNECTED")
}

func CreateUsersTableWithAdmin() {
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

	if _, err := DB.Exec(context.Background(), query); err != nil {
		panic(err)
	}
}
