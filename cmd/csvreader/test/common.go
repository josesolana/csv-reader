package test

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/josesolana/csv-reader/cmd/csvreader/filehandler"
)

func lineCounter(name string) (int, error) {
	name, err := filehandler.GetFilePath(name)
	if err != nil {
		return 0, err
	}

	file, err := os.Open(name)
	if err != nil {
		return 0, err
	}

	r := bufio.NewReader(file)

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return 0, err
		}
	}
}
