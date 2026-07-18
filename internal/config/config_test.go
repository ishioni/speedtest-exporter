package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SPEEDTEST_PORT", "")
	t.Setenv("SPEEDTEST_LISTEN_ADDRESS", "")
	t.Setenv("SPEEDTEST_TIMEOUT", "")
	t.Setenv("SPEEDTEST_CACHE_FOR", "")
	t.Setenv("SPEEDTEST_LOG_LEVEL", "")
	t.Setenv("SPEEDTEST_LOG_FORMAT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ListenAddress != ":9798" || cfg.Timeout != 90*time.Second || cfg.CacheFor != 0 {
		t.Fatalf("Load() = %#v, want defaults", cfg)
	}
}

func TestLoadSupportsGoDurationValues(t *testing.T) {
	t.Setenv("SPEEDTEST_TIMEOUT", "42s")
	t.Setenv("SPEEDTEST_CACHE_FOR", "5s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Timeout != 42*time.Second || cfg.CacheFor != 5*time.Second {
		t.Fatalf("Load() = %#v", cfg)
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("SPEEDTEST_TIMEOUT", "42")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}
