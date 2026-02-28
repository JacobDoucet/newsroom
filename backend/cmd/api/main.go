package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JacobDoucet/newsroom/internal/db"
	"github.com/JacobDoucet/newsroom/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Initialize database
	ctx := context.Background()
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")

	// Initialize handlers
	h := handlers.New(pool)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Arc management
		r.Post("/arcs", h.CreateArc)
		r.Get("/arcs", h.ListArcs)
		r.Post("/arcs/{id}/start", h.StartArc)
		r.Post("/arcs/{id}/advance", h.AdvanceArc)

		// Region initialization
		r.Post("/arcs/{id}/regions/init", h.InitRegion)

		// Packet generation
		r.Post("/arcs/{id}/packets/generate", h.GeneratePacket)
		r.Get("/packets/{id}", h.GetPacket)
		r.Get("/packets/{id}/candidates", h.GetPacketCandidates)

		// Review actions
		r.Post("/candidates/{id}/select", h.SelectCandidate)
		r.Post("/candidates/{id}/reject", h.RejectCandidate)
		r.Post("/packets/{id}/rank", h.RankCandidates)
		r.Post("/candidates/{id}/edit", h.EditCandidate)
		r.Post("/candidates/{id}/publish", h.PublishCandidate)

		// Public endpoints
		r.Get("/public/latest", h.GetLatestArticles)
		r.Get("/public/article/{id}", h.GetArticle)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
