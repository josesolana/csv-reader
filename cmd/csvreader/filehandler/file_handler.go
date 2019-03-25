package filehandler

import (
	"bufio"
	"encoding/csv"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/josesolana/csv-reader/constants"
)

// FileHandler Wrappeer to read files.
type FileHandler struct {
	reader *csv.Reader
	file   *os.File
}

// NewFileHandler Filer Handler
func NewFileHandler(name string) (Readable, error) {
	filePath, err := GetFilePath(name)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Cannot open file: %s", filePath)
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = 0
	//reader.ReuseRecord = true //TODO: PROBARLO PUESTO

	return &FileHandler{
		file:   file,
		reader: reader,
	}, nil
}

// Close closes the File, rendering it unusable for I/O.
// It returns an error, if any.
func (f *FileHandler) Close() error {
	return f.file.Close()
}

// Close closes the File, rendering it unusable for I/O.
// It returns an error, if any.
func (f *FileHandler) Read() ([]string, error) {
	return f.reader.Read()
}

//GetFilePath Fetch the path for a file.
func GetFilePath(fileName string) (string, error) {
	if filepath.Ext(fileName) != constants.AcceptedExt {
		log.Printf("File type not accepted: %s\n", filepath.Ext(fileName))
		return "", errors.New(constants.ErrExtensionFile)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Println(constants.ErrWorkingDirectory)
		return "", err
	}
	return filepath.Join(wd, fileName), nil
}
