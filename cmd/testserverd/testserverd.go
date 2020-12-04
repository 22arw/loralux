package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// ScrapeResponse is the response that gets sent out of the /scrape endpoint of
// testserverd. The number of data points and the points themselves are randomly
// generated.
type ScrapeResponse struct {
	Data []float64 `json:"data"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// Build logger with default options.
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Printf("build logger: %v", err)
		exitCode = 1
		return
	}

	// Add a service key with the value of testserverd to all logs emitted from this
	// daemon to namespace the logs.
	logger = logger.With(zap.String("service", "testserverd"))

	// Create a mux that will server as the handler for the test server.
	mux := http.NewServeMux()
	mux.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		var resp ScrapeResponse

		// Generate somewhere between [10, 50] data points.
		pts := rand.Intn(41) + 10
		resp.Data = make([]float64, pts)

		logger.Info("/scrape invoked", zap.Int("points", pts))

		// Populate the data points with random floats.
		for i := 0; i < pts; i++ {
			resp.Data[i] = rand.Float64()
		}

		// Marshal the randomly generated response.
		b, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		// Write the header and response data.
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})

	// Create the test server.
	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 15,
	}

	// Create a channel for interrupt and termination signals to be caught on
	// to possibly facilitate graceful shutdowns.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create a channel to catch HTTP-related non-recoverable errors.
	serverErr := make(chan error, 1)

	// Start the HTTP server.
	go func() {
		logger.Info("server started", zap.String("address", server.Addr))
		serverErr <- server.ListenAndServe()
	}()

	// Block until either a shutdown signal or an HTTP-related non-recoverable error
	// is encountered.
	select {
	case <-shutdown:
		logger.Info("shutdown signal received, attempting to gracefully terminate server")
		signal.Reset(os.Interrupt, syscall.SIGTERM)
	case err := <-serverErr:
		logger.Error("fatal server error", zap.Error(err))
		exitCode = 1
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Attempt to gracefully shutdown.
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown error", zap.Error(err))
		logger.Info("attempting to forcefully shutdown server")

		if err := server.Close(); err != nil {
			logger.Error("forceful shutdown error", zap.Error(err))
			exitCode = 1
			return
		}
	}
}
