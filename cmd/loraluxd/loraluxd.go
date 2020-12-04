package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/22arw/loralux/cmd/loraluxd/config"
	"github.com/22arw/loralux/cmd/loraluxd/scrape"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// flagEnvFile is a variable that is collected from the -envFile="" flag upon running the
// lorafication binary.
var flagEnvFile string

// flagVerbose is a variable that is collected from the -verbose flag upon running the
// lorafication binary.
var flagVerbose bool

func init() {
	// Register the collection of the -envFile="" flag if it was passed to the binary.
	flag.StringVar(&flagEnvFile, "envFile", "", "the path of a JSON or YAML file that contains configuration variables, if not supplied the configuration will be collected from the environment.")

	// Register the collection of the -verbose flag if it was passed to the binary.
	flag.BoolVar(&flagVerbose, "verbose", false, "display verbose information")

	// Collect all registered CLI flags.
	flag.Parse()
}

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	var cfg config.Config
	var err error

	// Process configuration variables from either a file or the environment
	// depending on whether or not the -envFile="" flag was specified.
	if flagEnvFile != "" {
		if cfg, err = config.FromFile(flagEnvFile); err != nil {
			log.Printf("collect config from file: %v", err)
			exitCode = 1
			return
		}
	} else {
		if cfg, err = config.FromEnvironment(); err != nil {
			log.Printf("collect config from environment: %v", err)
			exitCode = 1
			return
		}
	}

	// Configure the logger.
	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(cfg.LogLevel))
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Build the logger.
	logger, err := zCfg.Build()
	if err != nil {
		log.Printf("build logger: %v", err)
		exitCode = 1
		return
	}

	// Add a service key with the value of loralux to all logs emitted from this
	// daemon to namespace the logs.
	logger = logger.With(zap.String("service", "loraluxd"))

	// Display verbose information if it was requested.
	if flagVerbose {
		// Log the parsed CLI flags.
		logger.Info("values of CLI flags",
			zap.String("envFile", flagEnvFile),
			zap.Bool("verbose", flagVerbose))

		// Log the parsed configuration values.
		logger.Info("values of configuration",
			zap.Int("logLevel", cfg.LogLevel),
			zap.String("serverAddress", cfg.ServerAddress),
			zap.String("scrapeEndpoint", cfg.ScrapeEndpoint),
			zap.Duration("scapeInterval", cfg.ScrapeInterval.Duration),
			zap.Duration("readTimeout", cfg.ReadTimeout.Duration))
	}

	// Configure the scraper used to scrape the LoRaWAN server for lumen sensor
	// readings.
	scraper := scrape.NewScraper(cfg.ServerAddress, cfg.ScrapeEndpoint, cfg.ReadTimeout.Duration)

	// Create a channel for interrupt and termination signals to be caught on
	// to possibly facilitate graceful shutdowns.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Set up an interval to scrape the lorawan server on.
	interval := time.NewTicker(cfg.ScrapeInterval.Duration)
	defer interval.Stop()

	logger.Info("starting to scrape lorawan server",
		zap.Duration("interval", cfg.ScrapeInterval.Duration),
		zap.String("address", cfg.ServerAddress))

	var exit bool
	for {
		select {
		case <-interval.C:
			resp, err := scraper.Scrape()
			if err != nil {
				logger.Warn("error encountered while scraping server", zap.Error(err))
				continue
			}

			if flagVerbose {
				logger.Info("successfully scraped server", zap.Int("points", len(resp.Data)))
			}
		case <-shutdown:
			logger.Info("shutdown signal received, attempting to shutdown gracefully")
			exit = true
		}

		if exit {
			break
		}
	}
}
