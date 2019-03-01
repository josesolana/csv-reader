package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/josesolana/csv-reader/cmd/crmintegrator/integrator"
)

func main() {
	// To interrupt the executable
	runCh := make(chan os.Signal, 1)
	signal.Notify(runCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	i := integrator.NewIntegrator()
	i.Migrate(&runCh)
}
