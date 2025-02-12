package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/google/uuid"

	"github.com/Kreagentle/teya_ledger/internal/handlers"
	"github.com/Kreagentle/teya_ledger/internal/models"
)

var (
	address = flag.String("addr", ":8000", "Address to bind the HTTP ledger server")
)

func run(ctx context.Context, address string) error {
	// initialise account storage
	accountsStorage := make(map[uuid.UUID]models.LedgerAccount)

	// initialise HTTP routes
	router := handlers.NewRouter(&handlers.AccountStorage{LedgerAccountsStorage: accountsStorage})

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// start the server in a goroutine so that it doesn't block the shutdown handling below
	go func() {
		log.Printf("Starting ledger on %s...\n", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error starting ledger: %s", err)
		}
	}()

	// wait for the context to be done (canceled) and gracefully shutdown the server
	<-ctx.Done()
	log.Println("Shutting down ledger...")
	return server.Shutdown(context.Background())
}

func main() {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, *address); err != nil {
		log.Fatalf("Error: %s\n", err)
	}
}
