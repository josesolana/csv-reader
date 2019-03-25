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

	p, err := processor.NewProcessor(os.Args[1])
	if err != nil {
		log.Fatalf("Fatal Error: %s\n", err)
	}

	if err := p.Migrate(); err != nil {
		log.Fatalf("Fatal Error: %s\n", err)
	}
}
