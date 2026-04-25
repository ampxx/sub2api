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
	// Bind to localhost by default instead of all interfaces — safer for local dev
	flag.StringVar(&cfg.Host, "host", envOrDefault("HOST", "127.0.0.1"), "Host to bind to (env: HOST)")
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
//	url   — the upstream subscription URL (required)
//	type  — output format, e.g. "clash", "v2ray" (optional, defaults to raw)
func handleSub(c *gin.Context) {
	subURL := c.Query("url")
	if subURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	// TODO: fetch and convert subscription content
	c.JSON(http.StatusNotImplemented, gin.H{"message": "conversion not yet implemented", "url": subURL})
}
