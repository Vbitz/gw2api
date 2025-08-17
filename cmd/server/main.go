package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"j5.nz/gw2/internal/discord"
	"j5.nz/gw2/internal/gw2api"
	"j5.nz/gw2/internal/web"
)

func main() {
	var (
		discordToken = flag.String("discord-token", "", "Discord bot token")
		port         = flag.String("port", "8080", "Web server port")
		webOnly      = flag.Bool("web-only", false, "Run only the web server (no Discord bot)")
		discordOnly  = flag.Bool("discord-only", false, "Run only the Discord bot (no web server)")
	)
	flag.Parse()

	// Create GW2 API client
	gw2Client := gw2api.NewClient(
		gw2api.WithTimeout(30*time.Second),
		gw2api.WithUserAgent("gw2api-server/1.0"),
	)

	// Channel to handle shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start Discord bot if requested and token is provided
	var bot *discord.Bot
	if !*webOnly {
		if *discordToken == "" {
			if *discordOnly {
				log.Fatal("Discord token is required when running in discord-only mode")
			}
			log.Println("Warning: No Discord token provided, Discord bot will not start")
		} else {
			var err error
			bot, err = discord.NewBot(*discordToken, gw2Client)
			if err != nil {
				log.Fatalf("Failed to create Discord bot: %v", err)
			}

			err = bot.Start()
			if err != nil {
				log.Fatalf("Failed to start Discord bot: %v", err)
			}
			log.Println("Discord bot started successfully")
		}
	}

	// Start web server if requested
	var server *http.Server
	if !*discordOnly {
		webServer := web.NewServer(gw2Client)
		mux := webServer.SetupRoutes()

		server = &http.Server{
			Addr:    ":" + *port,
			Handler: mux,
		}

		go func() {
			log.Printf("Web server starting on port %s", *port)
			log.Printf("Visit http://localhost:%s for the web interface", *port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start web server: %v", err)
			}
		}()
	}

	// If running both services, show status
	if !*webOnly && !*discordOnly {
		if *discordToken != "" {
			log.Println("Running both Discord bot and web server")
			log.Printf("Discord bot: Active with slash commands")
			log.Printf("Web server: http://localhost:%s", *port)
		} else {
			log.Printf("Running web server only (no Discord token provided)")
			log.Printf("Web server: http://localhost:%s", *port)
		}
	}

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop Discord bot
	if bot != nil {
		log.Println("Stopping Discord bot...")
		if err := bot.Stop(); err != nil {
			log.Printf("Error stopping Discord bot: %v", err)
		} else {
			log.Println("Discord bot stopped")
		}
	}

	// Stop web server
	if server != nil {
		log.Println("Stopping web server...")
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error stopping web server: %v", err)
		} else {
			log.Println("Web server stopped")
		}
	}

	log.Println("Shutdown complete")
}