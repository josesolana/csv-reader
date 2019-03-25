package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	c "github.com/josesolana/csv-reader/constants"
	"github.com/pkg/errors"
)

// Db Database Handler & Wrapper
type Db struct {
	db                               *sql.DB
	read, isProcessed, increaseRetry *sql.Stmt
	tx                               *sql.Tx
}

// NewDB Set up the environment.
func NewDB(name string) DB {
	db := &Db{
		db: ConnectDb(),
	}

	if err := db.checkTableExist(name); err != nil {
		log.Fatalln(err)
	}
	db.createRead(name)
	db.createUpdateIsProcessed(name)
	db.createUpdateIncreaseRetry(name)
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

// Begin Begin a transaction
func (d *Db) Begin() error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	d.tx = tx
	return nil
}

// Commit Commit a transaction
func (d *Db) Commit() error {
	return d.tx.Commit()
}

// Reads from DB
func (d *Db) Read() (*sql.Rows, error) {
	rows, err := d.read.Query()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SetAsProcessed Set a row as processed to not be taking into accout
// in the next lap.
func (d *Db) SetAsProcessed(id int) error {
	if _, err := d.isProcessed.Exec(id); err != nil {
		return err
	}
	return nil
}

// IncreaseRetry Increase by one retry value.
func (d *Db) IncreaseRetry(id int) error {
	if _, err := d.increaseRetry.Exec(id); err != nil {
		return err
	}
	return nil
}

//Close returns the connection to the connection pool.
func (d *Db) Close() []error {
	errs := make([]error, 0)

	if err := d.increaseRetry.Close(); err != nil {
		log.Println("Cannot Close increaseRetry")
		errs = append(errs, err)
	}

	if err := d.isProcessed.Close(); err != nil {
		log.Println("Cannot Close isProcessed")
		errs = append(errs, err)
	}

	if err := d.read.Close(); err != nil {
		log.Println("Cannot Close Read")
		errs = append(errs, err)
	}

	if err := d.db.Close(); err != nil {
		log.Println("Cannot Close DB Connection")
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs
	}
	return errs
}

func (d *Db) rollbackAndWrapErr(err error) error {
	rerr := d.tx.Rollback()
	if rerr != nil && err != nil {
		return errors.Wrap(err, rerr.Error())
	}
	return err
}

func (d *Db) createRead(name string) {
	query := `
	Select id, is_processed, retry, %s.*
	FROM %s
	WHERE NOT is_processed and retry <= %d
	LIMIT %d
	FOR UPDATE SKIP LOCKED`
	query = fmt.Sprintf(query, name, name, c.TotalRetry, c.BatchSizeRow)

	read, err := d.db.Prepare(query)
	if err != nil {
		log.Fatalf("Couldn't create Read. Error: %s\n", err)
	}
	d.read = read
}

func (d *Db) createUpdateIsProcessed(name string) {
	query := `
	UPDATE %s
	SET is_processed = true
	WHERE id = $1`
	query = fmt.Sprintf(query, name)

	isProcessed, err := d.db.Prepare(query)
	if err != nil {
		log.Fatalf("Couldn't create isProcessed. Error: %s\n", err)
	}
	d.isProcessed = isProcessed
}

func (d *Db) createUpdateIncreaseRetry(name string) {
	query := `
	UPDATE %s
	SET retry = retry + 1
	WHERE id = $1`
	query = fmt.Sprintf(query, name)

	increaseRetry, err := d.db.Prepare(query)
	if err != nil {
		log.Fatalf("Couldn't create increaseRetry. Error: %s\n", err)
	}
	d.increaseRetry = increaseRetry
}

func (d *Db) checkTableExist(name string) error {
	query := `
	SELECT id
	FROM %s
	LIMIT 1`
	row, err := d.db.Exec(fmt.Sprintf(query, name))

	if err != nil {
		log.Printf("Cannot verify if table %s exists. Error: %s\n", name, err)
		return err
	}
	count, err := row.RowsAffected()
	if err != nil {
		log.Printf("Cannot verify if table %s exists. Error: %s\n", name, err)
		return err
	}
	if int64(0) == count {
		log.Printf("Given table %s doesn't exists.\n", name)
		return errors.New(c.ErrTableNoExists)
	}
	return nil
}

func getDBName() string {
	if rm := os.Getenv(c.RunMode); strings.ToUpper(rm) == c.Test {
		return c.DbNameTest
	}
	return c.DbName
}
