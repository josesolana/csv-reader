package test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/josesolana/csv-reader/cmd/crmintegrator/database"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type IntegratorTest struct {
	suite.Suite
	db *sql.DB
}

func TestIntegratorController(t *testing.T) {
	suite.Run(t, new(IntegratorTest))
}

func (it *IntegratorTest) SetupSuite() {
	it.db = database.ConnectDb()
}

func (it *IntegratorTest) TearDownSuite() {
	if err := it.db.Close(); err != nil {
		log.Fatalln(err)
	}
}

func (it *IntegratorTest) tearDown(name string) {
	_, err := it.db.Exec("DELETE FROM " + name + " WHERE id >= 0")
	if err != nil {
		log.Fatalln(err)
	}
}
