package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/josesolana/csv-reader/cmd/csvreader/processor"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("Filename should be provided")
	}

	p := processor.NewProcessor()
	if err := p.Migrate(os.Args[1]); err != nil {
		log.Fatalf("Fatal Error: %s", err.Error())
	}
}
