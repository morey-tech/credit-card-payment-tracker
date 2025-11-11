package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/morey-tech/credit-card-payment-tracker/pkg/config"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/database"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/handlers"
)

func main() {
	// Load application configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	if cfg.DiscordWebhookURL != "" {
		log.Printf("Discord webhook configured")
	} else {
		log.Printf("Discord webhook not configured (notifications disabled)")
	}

	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./credit_cards.db"
	}

	// Initialize database
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Set up HTTP routes using ServeMux
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", handlers.HealthCheck)
	mux.HandleFunc("/api/v1/cards", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateCard(w, r)
		} else {
			handlers.GetCards(w, r)
		}
	})
	mux.HandleFunc("/api/v1/cards/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetCardByID(w, r)
		case http.MethodPut:
			handlers.UpdateCard(w, r)
		case http.MethodDelete:
			handlers.DeleteCard(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/statements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateStatement(w, r)
		} else {
			handlers.GetStatements(w, r)
		}
	})
	mux.HandleFunc("/api/v1/statements/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a schedule request
		if len(r.URL.Path) > len("/api/v1/statements/") {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) >= 6 && pathParts[5] == "schedule" {
				handlers.SchedulePayment(w, r)
				return
			}
		}
		handlers.UpdateStatement(w, r)
	})
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			handlers.UpdateSettings(w, r)
		} else {
			handlers.GetSettings(w, r)
		}
	})

	// Serve static files at /static/ path
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	// Serve index.html at root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// corsMiddleware adds CORS headers for local development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
