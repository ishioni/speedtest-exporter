// Package config loads and validates speedtest-exporter configuration.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort    = "9798"
	defaultTimeout = 90 * time.Second
)

// Config is the runtime configuration for the exporter.
type Config struct {
	ListenAddress string
	Binary        string
	ServerID      string
	Timeout       time.Duration
	CacheFor      time.Duration
	LogLevel      string
	LogFormat     string
}

// Load reads configuration from SPEEDTEST_* environment variables.
func Load() (Config, error) {
	port := value("SPEEDTEST_PORT", defaultPort)
	listenAddress := value("SPEEDTEST_LISTEN_ADDRESS", net.JoinHostPort("", port))
	if _, _, err := net.SplitHostPort(listenAddress); err != nil {
		return Config{}, fmt.Errorf("SPEEDTEST_LISTEN_ADDRESS must be host:port: %w", err)
	}

	timeout, err := duration("SPEEDTEST_TIMEOUT", defaultTimeout, false)
	if err != nil {
		return Config{}, err
	}
	cacheFor, err := duration("SPEEDTEST_CACHE_FOR", 0, true)
	if err != nil {
		return Config{}, err
	}

	logLevel := strings.ToLower(value("SPEEDTEST_LOG_LEVEL", "info"))
	switch logLevel {
	case "debug", "info", "warn", "warning", "error":
	default:
		return Config{}, fmt.Errorf("SPEEDTEST_LOG_LEVEL must be debug, info, warn, or error")
	}
	logFormat := strings.ToLower(value("SPEEDTEST_LOG_FORMAT", "json"))
	if logFormat != "json" && logFormat != "text" {
		return Config{}, fmt.Errorf("SPEEDTEST_LOG_FORMAT must be json or text")
	}

	return Config{
		ListenAddress: listenAddress,
		Binary:        value("SPEEDTEST_BINARY", "speedtest"),
		ServerID:      os.Getenv("SPEEDTEST_SERVER"),
		Timeout:       timeout,
		CacheFor:      cacheFor,
		LogLevel:      logLevel,
		LogFormat:     logFormat,
	}, nil
}

func duration(name string, fallback time.Duration, allowZero bool) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}

	// The legacy exporter accepted a whole number of seconds. Retain that form
	// while also allowing standard Go durations such as "2m".
	if seconds, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if seconds < 0 || (seconds == 0 && !allowZero) {
			return 0, fmt.Errorf("%s must be %s", name, positiveDescription(allowZero))
		}
		return time.Duration(seconds) * time.Second, nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed < 0 || (parsed == 0 && !allowZero) {
		return 0, fmt.Errorf("%s must be %s (for example 90 or 90s)", name, positiveDescription(allowZero))
	}
	return parsed, nil
}

func positiveDescription(allowZero bool) string {
	if allowZero {
		return "zero or a positive duration"
	}
	return "a positive duration"
}

func value(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
