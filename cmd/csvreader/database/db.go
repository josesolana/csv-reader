package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	c "github.com/josesolana/csv-reader/constants"
)

var (
	once sync.Once
)

// Db Database Handler & Wrapper
type Db struct {
	db     *sql.DB
	insert *sql.Stmt
}

// NewDb Set up the environment.
func NewDb() Db {
	db, err := sql.Open(c.DbDriver, fmt.Sprintf("user=%s dbname=%s password =%s sslmode=require", c.DbUser, c.DbName, c.DbPass))
	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s", err.Error())
	}

	once.Do(func() { createTable(db) })

	insert, err := db.Prepare(`
		INSERT INTO customers (id, first_name, last_name, email, phone)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s", err.Error())
	}
	return Db{
		db:     db,
		insert: insert,
	}
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
func (d *Db) Close() {
	d.insert.Close()
	d.db.Close()
}

func createTable(db *sql.DB) {

	_, err := db.Exec(`DROP TABLE IF EXISTS customers`)

	if err != nil {
		log.Fatalf("Cannot drop customer table. Error: %s", err.Error())
		return
	}

	_, err = db.Exec(`CREATE TABLE customers (
		id int,
		first_name varchar(255),
		last_name varchar(255),
		email varchar(50),
		phone varchar(50),
		PRIMARY KEY (id))`)

	if err != nil {
		log.Fatalf("Cannot create the Customer Table. Error: %s", err.Error())
	}
	return
}
