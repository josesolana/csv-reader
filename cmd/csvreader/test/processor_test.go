package test

import (
	"io"
	"log"
	"path"
	"testing"

	_ "github.com/lib/pq"

	"database/sql"

	"github.com/josesolana/csv-reader/cmd/csvreader/database"
	p "github.com/josesolana/csv-reader/cmd/csvreader/processor"
	c "github.com/josesolana/csv-reader/constants"
	"github.com/stretchr/testify/suite"
)

type ProcessorTest struct {
	suite.Suite
	db *sql.DB
}

func TestProcessorController(t *testing.T) {
	suite.Run(t, new(ProcessorTest))
}

func (pt *ProcessorTest) SetupSuite() {
	pt.db = database.ConnectDb()
}

func (pt *ProcessorTest) TearDownSuite() {
	if err := pt.db.Close(); err != nil {
		log.Fatalln(err)
	}
}

func (pt *ProcessorTest) TestCSVBlank() {
	name := c.FileNameMockEmpty + c.AcceptedExt
	defer pt.tearDown(c.FileNameMockEmpty)

	proc, err := p.NewProcessor(name)
	pt.NotNil(err)
	pt.EqualError(err, io.EOF.Error())
	pt.Nil(proc)
}

func (pt *ProcessorTest) TestCSVErrorReading() {
	name := c.FileNameMockErrorReading + c.AcceptedExt
	defer pt.tearDown(c.FileNameMockErrorReading)

	proc, err := p.NewProcessor(name)
	pt.NotNil(err)
	pt.NotEqual(err.Error(), io.EOF.Error())
	pt.Nil(proc)
}

func (pt *ProcessorTest) TestCSVSuccess() {
	name := c.FileNameMock + c.AcceptedExt
	defer pt.tearDown(c.FileNameMock)

	fileLines, err := lineCounter(name)
	pt.Nil(err)

	proc, err := p.NewProcessor(name)
	pt.Nil(err)

	err = proc.Migrate()
	pt.Nil(err)

	var count int
	name = path.Base(c.FileNameMock)
	err = pt.db.QueryRow("SELECT COUNT(*) FROM " + name).Scan(&count)
	pt.Nil(err)

	pt.Equal(fileLines, count)

}

func (pt *ProcessorTest) tearDown(name string) {
	name = path.Base(name)
	_, err := pt.db.Exec("DROP TABLE IF EXISTS " + name)
	if err != nil {
		log.Fatalf("Teardown fail dropping %s table. Error: %s\n ", name, err)
	}
}
