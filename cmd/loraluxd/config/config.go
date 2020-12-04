// Package config exposes a Config type that pulls values from configuration files for
// loralux daemon.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/22arw/loralux/internal/platform/duration"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// Constant block for Config struct field defaults.
const (
	// DefaultLogLevel is the default value of the LogLevel struct field on the Config
	// type.
	DefaultLogLevel = 0

	// DefaultServerAddress is the default value of the ServerAddress struct field on the
	// Config type.
	DefaultServerAddress = "http://localhost:8080"

	// DefaultScrapeEndpoint is the default value of the ScrapeEndpoint struct field on
	// the Config type.
	DefaultScrapeEndpoint = "/scrape"

	// DefaultScrapeInterval is the default value of the ScrapeInterval struct field on
	// the Config type.
	DefaultScrapeInterval = 5 * time.Second

	// DefaultReadTimeout is the default value of the ReadTimeout struct field on the
	// Config type.
	DefaultReadTimeout = 10 * time.Second
)

// Config is a struct that contains the struct fields necessary for running the
// loralux daemon.
type Config struct {
	LogLevel       int               `json:"logLevel" yaml:"logLevel" envconfig:"LOG_LEVEL"`
	ServerAddress  string            `json:"serverAddress" yaml:"serverAddress" envconfig:"SERVER_ADDRESS"`
	ScrapeEndpoint string            `json:"scrapeEndpoint" yaml:"scrapeEndpoint" envconfig:"SCRAPE_ENDPOINT"`
	ScrapeInterval duration.Duration `json:"scrapeInterval" yaml:"scrapeInterval" envconfig:"SCRAPE_INTERVAL"`
	ReadTimeout    duration.Duration `json:"readTimeout" yaml:"readTimeout" envconfig:"READ_TIMEOUT"`
}

// FromEnvironment gathers the configuration variables from the environment.
func FromEnvironment() (Config, error) {
	var c Config

	if err := envconfig.Process("LORALUX", &c); err != nil {
		return c, fmt.Errorf("process environment variables: %w", err)
	}
	c.Defaults()

	if err := c.Validate(); err != nil {
		return c, fmt.Errorf("validate configuration: %w", err)
	}

	return c, nil
}

// FromFile gathers the configuration variables from either a JSON or YAML file.
func FromFile(fp string) (Config, error) {
	var c Config

	f, err := os.Open(fp)
	if err != nil {
		return c, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	switch ext := filepath.Ext(fp); ext {
	case ".json":
		if err := json.NewDecoder(f).Decode(&c); err != nil {
			return c, fmt.Errorf("decode json file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.NewDecoder(f).Decode(&c); err != nil {
			return c, fmt.Errorf("decode json file: %w", err)
		}
	}
	c.Defaults()

	if err := c.Validate(); err != nil {
		return c, fmt.Errorf("validate configuration: %w", err)
	}

	return c, nil
}

// Defaults is a method on the Config pointer receiver that sets defaults on the receiver
// where values are not already set.
func (c *Config) Defaults() {
	if c.LogLevel == 0 {
		c.LogLevel = DefaultLogLevel
	}

	if c.ServerAddress == "" {
		c.ServerAddress = DefaultServerAddress
	}

	if c.ScrapeEndpoint == "" {
		c.ScrapeEndpoint = DefaultScrapeEndpoint
	}

	if c.ScrapeInterval.IsEmpty() {
		c.ScrapeInterval.Duration = DefaultScrapeInterval
	}

	if c.ReadTimeout.IsEmpty() {
		c.ReadTimeout.Duration = DefaultReadTimeout
	}
}

// Validate is a method on the Config pointer receiver that validates the values set on
// the receiver.
func (c *Config) Validate() error {
	if actual, min, max := zapcore.Level(c.LogLevel), zapcore.DebugLevel, zapcore.FatalLevel; actual < min || actual > max {
		return fmt.Errorf("log level must be [%d, %d]", min, max)
	}

	if _, err := url.ParseRequestURI(c.ServerAddress); err != nil {
		return fmt.Errorf("server address must be a valid URI: %w", err)
	}

	if c.ScrapeEndpoint == "" || !strings.HasPrefix(c.ScrapeEndpoint, "/") {
		return errors.New("scrape endpoint must be supplied and start with a /")
	}

	if c.ScrapeInterval.IsEmpty() {
		return errors.New("scrape interval must be > 0ms")
	}

	if c.ReadTimeout.IsEmpty() {
		return errors.New("read timeout must be > 0ms")
	}

	return nil
}
