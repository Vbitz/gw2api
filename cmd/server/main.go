package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"j5.nz/gw2/internal/cache"
	"j5.nz/gw2/internal/gw2api"
	"j5.nz/gw2/internal/web"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, don't fail if it doesn't exist
		log.Printf("Note: .env file not found or could not be loaded: %v", err)
	}

	// Parse command line flags
	verbose := flag.Bool("verbose", false, "Enable verbose API request logging")
	addr := flag.String("addr", ":9090", "HTTP server address")
	flag.Parse()

	// Get API key from environment
	apiKey := os.Getenv("GW2_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GW2_API_KEY not set in environment or .env file")
		log.Println("Some features requiring authentication will not work")
	}

	// Create GW2 API client with optional verbose logging
	var clientOptions []gw2api.ClientOption
	clientOptions = append(clientOptions, gw2api.WithDataCache("data"))

	if apiKey != "" {
		clientOptions = append(clientOptions, gw2api.WithAPIKey(apiKey))
		log.Println("API key loaded from environment")
	}

	if *verbose {
		clientOptions = append(clientOptions, gw2api.WithVerboseLogging())
		log.Println("Verbose API logging enabled")
	}

	client := gw2api.NewClient(clientOptions...)

	// Create cache for trading post prices (3 hour TTL)
	priceCache := cache.NewLRUCache(10000)

	// Create web server
	server := web.NewServer(client, priceCache)

	// Setup HTTP server
	srv := &http.Server{
		Addr:    *addr,
		Handler: server,

		// Good practice timeouts
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		l, err := net.Listen("tcp", *addr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", *addr, err)
		}
		defer l.Close()

		fmt.Printf("Starting server on http://%s\n", l.Addr().String())
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("Server exited")
}
