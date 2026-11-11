// Command alert-receiver is the agent's ingress point: an Alertmanager webhook
// receiver. In the environment phase it only records notifications so the
// experiment can measure when an alert reaches the agent; in the prototype
// phase the same endpoint hands each firing alert to the orchestrator.
//
// Endpoints:
//
//	POST /alerts   Alertmanager webhook (payload version "4")
//	GET  /alerts   received notifications, oldest first (?since=<unix ts> to filter)
//	GET  /healthz  liveness
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nojussta/k8s-diag-agent/internal/receiver"
)

func main() {
	port := envInt("PORT", 8080)
	store := receiver.NewStore(envInt("MAX_STORED", 500))
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           receiver.NewHandler(store, os.Stdout),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "started", "port": port})
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}
	return def
}
