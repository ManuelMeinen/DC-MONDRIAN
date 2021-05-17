package db

import (
	"fmt"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Backend wraps the database backend
type Backend struct {
	db *sql.DB
}

var DB *Backend

func SetupDB(path string) {
	var err error
	
	DB, err = New(path)
	
	if err != nil {
		log.Fatal(err)
	}
	// if in memory DB is used, add some test data
	if path != ":memory:" {
		return
	}
}

func New(path string) (*Backend, error) {
	var err error

	db, err := sql.Open("sqlite3", fmt.Sprintf("%v", path))
	if err != nil {
		return nil, err
	}

	// from now on, close the sql database in case of error
	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	// prevent weird errors. (see https://stackoverflow.com/a/35805826)
	db.SetMaxOpenConns(1)

	// Make sure DB is reachable
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// set journaling to WAL
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatal(err)
	}

	// enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal(err)
	}

	// Ensure foreign keys are supported and enabled
	var enabled bool
	err = db.QueryRow("PRAGMA foreign_keys;").Scan(&enabled)
	if err == sql.ErrNoRows {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if !enabled {
		db.Close()
		log.Fatal(err)
	}

	// Check the schema version and set up new DB if necessary.
	var existingVersion int
	err = db.QueryRow("PRAGMA user_version;").Scan(&existingVersion)
	if err != nil {
		return nil, err
	}
	if existingVersion == 0 {
		if err = setup(db, Schema, SchemaVersion, path); err != nil {
			return nil, err
		}
	} else if existingVersion != SchemaVersion {
		return nil, err
	}

	return &Backend{db: db}, nil
}

func setup(db *sql.DB, schema string, schemaVersion int, path string) error {
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	// Write schema version to database.
	_, err = db.Exec(fmt.Sprintf("PRAGMA user_version = %d;", schemaVersion))
	if err != nil {
		return err
	}
	return nil
}



