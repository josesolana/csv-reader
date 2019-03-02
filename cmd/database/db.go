package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/josesolana/csv-reader/cmd/models"
	c "github.com/josesolana/csv-reader/constants"
)

// var once sync.Once

// Db Database Handler & Wrapper
type Db struct {
	db     *sql.DB
	insert *sql.Stmt
	read   *sql.Stmt
}

// NewDb Set up the environment.
func NewDb() Db {
	db, err := sql.Open(c.DbDriver, fmt.Sprintf("user=%s dbname=%s password =%s sslmode=require", c.DbUser, c.DbName, c.DbPass))
	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not establish a connection with the database. Error: %s", err.Error())
	}

	// once.Do(func() { createTable(db) })

	//TODO: Check amount of column to add # values -> $1, $2...
	insert, err := db.Prepare(`
		INSERT INTO ` + models.CustomerTableName + ` (` + models.CustomerColumns + `)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s", err.Error())
	}

	read, err := db.Prepare(`
		Select ` + models.CustomerColumns + `
		FROM ` + models.CustomerTableName + `
		ORDER BY ID
		LIMIT $1
		OFFSET $2
		`)

	if err != nil {
		log.Fatalf("Couldn't connect to Database. Error: %s", err.Error())
	}

	return Db{
		db:     db,
		insert: insert,
		read:   read,
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

// Reads from DB
func (d *Db) Read(limit, offset int64) (*sql.Rows, error) {
	return d.read.Query(limit, offset)
}

//Close returns the connection to the connection pool.
func (d *Db) Close() {
	d.insert.Close()
	d.db.Close()
}

// createTable Just to make easy to work in this Quiz.
// func createTable(db *sql.DB) {

// 	_, err := db.Exec(`DROP TABLE IF EXISTS ` + models.CustomerTableName + ``)

// 	if err != nil {
// 		log.Fatalf("Cannot drop "+models.CustomerTableName+"table. Error: %s", err.Error())
// 		return
// 	}

// 	_, err = db.Exec(`CREATE TABLE ` + models.CustomerTableName + ` (
// 		id bigint PRIMARY KEY,
// 		first_name varchar(255) NOT NULL,
// 		last_name varchar(255) NOT NULL,
// 		email varchar(50) NOT NULL,
// 		phone varchar(50) NOT NULL)`)

// 	if err != nil {
// 		log.Fatalf("Cannot create the Customer Table. Error: %s", err.Error())
// 	}
// 	return
// }
