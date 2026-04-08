package main

import (
	"log"
	"net/http"
	"os"

	"browser-sync-api/internal/auth"
	"browser-sync-api/internal/store"
	"browser-sync-api/internal/sync"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()
	// 1. Connect to Database (Change user/password/dbname to match your local Postgres)
	// Format: postgres://username:password@host:port/database_name?sslmode=disable
	dsn := os.Getenv("DB_DSN")
	db, err := store.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Handlers
	authHandler := &auth.Handler{DB: db}
	syncHandler := &sync.Handler{DB: db}

	// 3. Setup Routes (Using Go 1.22+ standard HTTP multiplexer)
	mux := http.NewServeMux()

	// Ensure we only accept POST requests for these endpoints
	mux.HandleFunc("POST /api/v1/auth/request-code", authHandler.RequestCode)
	mux.HandleFunc("POST /api/v1/auth/verify-code", authHandler.VerifyCode)

	// Protected Sync Routes (Wrapped in RequireAuth)
	mux.HandleFunc("POST /api/v1/sync/push", auth.RequireAuth(syncHandler.PushData))
	mux.HandleFunc("GET /api/v1/sync/pull", auth.RequireAuth(syncHandler.PullData))
	mux.HandleFunc("DELETE /api/v1/sync/account", auth.RequireAuth(syncHandler.DeleteAccount))

	// 4. Start Server
	log.Println("🚀 Server starting on port :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
