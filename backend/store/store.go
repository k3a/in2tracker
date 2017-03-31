// Package store implements persistent data storage with clean API.
// Please note that time.Time values are always stored in UTC in the database.
package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
	"k3a.me/money/backend/store/ddl"

	// drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func e(format string, args ...interface{}) error {
	return fmt.Errorf("store: "+format, args...)
}

// Store implements the persistent storage
type Store struct {
	db *sql.DB
}

// From creates the store from an existing db connection
func From(db *sql.DB) *Store {
	return &Store{db}
}

// New creates the store using the specified driver and db connection config
func New(driver, config string) *Store {
	return From(open(driver, config))
}

// NewTest creates the store from a in-memory test database
func NewTest() *Store {
	return From(openTest())
}

func open(driver, config string) *sql.DB {
	db, err := sql.Open(driver, config)
	if err != nil {
		log.Fatalln(err)
	}

	if driver == "mysql" {
		// see  https://github.com/go-sql-driver/mysql/issues/257
		// uncomment if you experience drops/timeouts
		// db.SetMaxIdleConns(0)
	}

	if err := pingDatabase(db); err != nil {
		log.Fatalf("database ping failed - %s", err.Error())
	}

	setupMeddler(driver)

	if err := migrateDatabase(driver, db); err != nil {
		log.Fatalf("database migration failed - %s", err.Error())
	}
	return db
}

// openTest opens a in-memory test database
func openTest() *sql.DB {
	return open("sqlite3", ":memory:")
}

func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		log.Printf("database ping failed, retrying in 1s...\n")
		time.Sleep(time.Second)
	}
	return
}

func migrateDatabase(driver string, db *sql.DB) error {
	var migrations = &migrate.AssetMigrationSource{
		Asset:    ddl.Asset,
		AssetDir: ddl.AssetDir,
		Dir:      "mysql", // load from mysql directory by default
	}
	_, err := migrate.Exec(db, driver, migrations, migrate.Up)
	return err
}

// helper function to setup the meddler default driver
// based on the selected driver name.
func setupMeddler(driver string) {
	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}
}
