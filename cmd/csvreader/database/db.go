package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	c "github.com/josesolana/csv-reader/constants"
)

var once sync.Once

// Db Database Handler & Wrapper
type Db struct {
	db     *sql.DB
	insert *sql.Stmt
}

// NewDB Set up the environment.
func NewDB(name string, row []string) DB {
	if len(row) == 0 {
		return nil
	}

	db := &Db{
		db: ConnectDb(),
	}

	name = path.Base(name)                 // Filename & Extension
	name = name[:strings.Index(name, ".")] // Without Extension

	for i, r := range row {
		row[i] = name + "_" + r
	}

	once.Do(func() { createTable(db.db, row, name) })

	db.createInsert(name, row)
	return db
}

// ConnectDb Set Driver, user, pass & database name
func ConnectDb() *sql.DB {

	db, err := sql.Open(c.DbDriver, fmt.Sprintf("user=%s dbname=%s password =%s sslmode=disable", c.DbUser, getDBName(), c.DbPass))
	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s\n", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not establish a connection with the database. Error: %s\n", err)
	}
	return db
}

// Insert into DB
func (d *Db) Insert(row ...string) error {
	interfaceRow := make([]interface{}, len(row))
	for i, s := range row {
		interfaceRow[i] = s
	}
	_, err := d.insert.Exec(interfaceRow...)
	return err
}

//Close returns the connection to the connection pool.
func (d *Db) Close() error {
	if err := d.insert.Close(); err != nil {
		return err
	}

	if err := d.db.Close(); err != nil {
		return err
	}
	return nil
}

func (d *Db) createInsert(name string, row []string) {
	values := ""
	var i int
	for i, _ = range row {
		values += fmt.Sprintf("$%d%s", i+1, ", ")
	}
	values = values[:len(values)-2]

	query := `
	INSERT INTO %s (%s)
	VALUES (%s)
	ON CONFLICT DO NOTHING`
	query = fmt.Sprintf(query, name, strings.Join(row, ", "), values)

	insert, err := d.db.Prepare(query)
	if err != nil {
		log.Fatalf("Couldn't create Insert. Error: %s\n", err)
	}
	d.insert = insert
}

func createTable(db *sql.DB, row []string, name string) {
	query := `CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			is_processed boolean DEFAULT FALSE,
			retry int DEFAULT 0,
			%s VARCHAR(255) NOT NULL,
			UNIQUE(%s)
			)`

	typeCol := strings.Join(row, " varchar(255) NOT NULL,\n")
	unqCol := strings.Join(row, ", ")

	query = fmt.Sprintf(query, name, typeCol, unqCol)

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Cannot create the %s Table. Error: %s\n", name, err)
	}
}

func getDBName() string {
	if rm := os.Getenv(c.RunMode); strings.ToUpper(rm) == c.Test {
		return c.DbNameTest
	}
	return c.DbName
}
