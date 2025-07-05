package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var (
		flags = flag.NewFlagSet("migrate", flag.ExitOnError)
		dir   = flags.String("dir", "migrations", "directory with migration files")
	)

	flags.Usage = func() {
		fmt.Println("Usage: migrate [OPTIONS] DRIVER DBSTRING COMMAND")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("    up                   Migrate the DB to the most recent version available")
		fmt.Println("    up-by-one            Migrate the DB up by 1")
		fmt.Println("    up-to VERSION        Migrate the DB to a specific VERSION")
		fmt.Println("    down                 Roll back the version by 1")
		fmt.Println("    down-to VERSION      Roll back to a specific VERSION")
		fmt.Println("    redo                 Re-run the latest migration")
		fmt.Println("    reset                Roll back all migrations")
		fmt.Println("    status               Dump the migration status for the current DB")
		fmt.Println("    version              Print the current version of the database")
		fmt.Println("    create NAME [sql|go] Creates new migration file with the current timestamp")
		fmt.Println("    fix                  Apply sequential ordering to migrations")
		flags.PrintDefaults()
	}

	flags.Parse(os.Args[1:])
	args := flags.Args()

	if len(args) < 3 {
		flags.Usage()
		return
	}

	driver, dbstring, command := args[0], args[1], args[2]

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		log.Fatalf("migrate: failed to open DB: %v\n", err)
	}
	defer db.Close()

	if err := goose.SetDialect(driver); err != nil {
		log.Fatalf("migrate: failed to set dialect: %v\n", err)
	}

	arguments := []string{}
	if len(args) > 3 {
		arguments = args[3:]
	}

	if err := goose.Run(command, db, *dir, arguments...); err != nil {
		log.Fatalf("migrate %v: %v", command, err)
	}
}
