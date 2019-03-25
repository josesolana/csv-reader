package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/josesolana/csv-reader/cmd/crmintegrator/integrator"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("Table to be migrated should be provided")
	}

	// To interrupt the executable
	runCh := make(chan os.Signal, 1)
	signal.Notify(runCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	i := integrator.NewIntegrator(os.Args[1], runCh)
	i.Migrate()
}
