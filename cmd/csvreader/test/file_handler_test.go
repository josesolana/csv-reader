package test

import (
	"io"
	"testing"

	fh "github.com/josesolana/csv-reader/cmd/csvreader/filehandler"
	c "github.com/josesolana/csv-reader/constants"
	"github.com/stretchr/testify/suite"
)

type FileHandlerTest struct {
	suite.Suite
}

func TestFileHandlerController(t *testing.T) {
	suite.Run(t, new(FileHandlerTest))
}

func (fht *FileHandlerTest) TestGetFilePathSuccess() {
	nfh, err := fh.GetFilePath(c.FileNameMock + c.AcceptedExt)
	fht.Nil(err)
	fht.NotEmpty(nfh)
}

func (fht *FileHandlerTest) TestGetFilePathWrongExtensionFile() {
	nfh, err := fh.GetFilePath("TEST." + c.AcceptedExt + "Wrong")
	fht.EqualError(err, c.ErrExtensionFile)
	fht.Empty(nfh)
}

func (fht *FileHandlerTest) TestNewFileHandlerNonExistentFile() {
	nfh, err := fh.NewFileHandler("NonExistent" + c.AcceptedExt)
	fht.Contains(err.Error(), "no such file or directory")
	fht.Nil(nfh)
}

func (fht *FileHandlerTest) TestNewFileHandlerNonExistentDirectory() {
	nfh, err := fh.NewFileHandler("NonExistentDirectory/" + c.FileNameMock + c.AcceptedExt)
	fht.Contains(err.Error(), "no such file or directory")
	fht.Nil(nfh)
}

func (fht *FileHandlerTest) TestNewFileHandlerRead() {
	name := c.FileNameMock + c.AcceptedExt
	lines, err := lineCounter(name)
	fht.Nil(err)

	nfh, err := fh.NewFileHandler(name)
	fht.Nil(err)
	fht.NotNil(nfh)
	defer nfh.Close()

	for i := 0; i <= lines; i++ {
		line, err := nfh.Read()
		fht.NotEmpty(line)
		fht.Nil(err)
	}
	_, err = nfh.Read()
	fht.NotNil(err)
	fht.Error(io.EOF, err)
}
