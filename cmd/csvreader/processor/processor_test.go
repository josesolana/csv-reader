package processor

import (
	"testing"

	"github.com/josesolana/csv-reader/cmd/csvreader/test/testutils"

	"github.com/stretchr/testify/suite"
)

type ProcessorTest struct {
	suite.Suite
	db         *testutils.MockDB
	mockRow    *testutils.MockRows
	processor  *Processor
}

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorTest))
}

func (pt *ProcessorTest) SetupTest() {
	pt.db = testutils.NewMockDB()
	pt.mockRow = testutils.NewMockRows()
	pt.mockReader = testutils.NewMockReader()
	pt.processor = NewProcessorWithValues(pt.mockReader, pt.db)
}

func (pt *ProcessorTest) Migrate() {
	
}
