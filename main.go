package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin")

// Version information, injected at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Config holds the application configuration
type Config struct {
	Port    int
	Host    string
	Token   string
	Debug   bool
}

func main() {
	cfg := parseConfig()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := setupRouter(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("sub2api %s starting on %s", Version, addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// parseConfig reads configuration from flags and environment variables.
// Flags take precedence over environment variables.
func parseConfig() Config {
	cfg := Config{}

	// Default to port 8080, or 8080 if PORT env is not set
	defaultPort := 8080
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		defaultPort = p
	}

	flag.IntVar(&cfg.Port, "port", defaultPort, "Port to listen on (env: PORT)")
	// Changed default host to 0.0.0.0 so it's reachable on my local network without extra flags
	flag.StringVar(&cfg.Host, "host", envOrDefault("HOST", "0.0.0.0"), "Host to bind to (env: HOST)")
	flag.StringVar(&cfg.Token, "token", os.Getenv("API_TOKEN"), "API authentication token (env: API_TOKEN)")
	flag.BoolVar(&cfg.Debug, "debug", os.Getenv("DEBUG") == "true", "Enable debug mode (env: DEBUG)")
	flag.Parse()

	return cfg
}

// envOrDefault returns the value of the named environment variable, or
// the provided fallback if the variable is unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// setupRouter initialises the Gin engine and registers all routes.
func setupRouter(cfg Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health / readiness endpoint — no auth required
	router.GET("/health", handleHealth)
	router.GET("/version", handleVersion)

	// API routes
	api := router.Group("/api")
	if cfg.Token != "" {
		api.Use(tokenAuthMiddleware(cfg.Token))
	}

	api.GET("/sub", handleSub)

	return router
}

// handleHealth returns a simple 200 OK to indicate the service is alive.
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleVersion returns the build version and timestamp.
func handleVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    Version,
		"build_time": BuildTime,
	})
}

// handleSub is the main subscription conversion endpoint.
// Query params:
//
//	url     — the upstream subscription URL (required)
//	type    — output format, e.g. "clash", "v2ray" (optional, defaults to raw)
//	timeout — request timeout in seconds (optional, defaults to 10)
//
// Note: reduced default timeout from 30s to 10s — my upstreams are all
// low-latency; a tighter timeout catches hangs faster and keeps the UX snappy.
//
// Note: also bumped default timeout floor to 5s minimum — occasionally my
// home connection has a brief spike and sub-5s was causing spurious failures.
func handleSub(c *gin.Context) {
	// default fetch timeout in seconds; 10s is plenty for my use case
	const defaultTimeout = 10
	// minimum allowed timeout — anything below 5s trips too often on slow days
	const minTimeout = 5

	timeoutSec := defaultTimeout
	if t, err := strconv.Atoi(c.Query("timeout")); err == nil {
		if t < minTimeout {
			t = minTimeout
		}
		timeoutSec = t
	}

	_ = timeoutSec // used by fetch logic (not yet wired up in this file)

	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
