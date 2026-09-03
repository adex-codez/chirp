package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"backend/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down" && os.Args[1] != "status") {
		log.Fatal("usage: go run ./cmd/migrate [up|down|status]")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.Database.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	var migrationErr error
	if os.Args[1] == "up" {
		migrationErr = goose.Up(db, "db/migrations")
	} else if os.Args[1] == "down" {
		migrationErr = goose.Down(db, "db/migrations")
	} else {
		migrationErr = goose.Status(db, "db/migrations")
	}
	if migrationErr != nil {
		log.Fatal(migrationErr)
	}
	fmt.Printf("migrations %s completed\n", os.Args[1])
}
