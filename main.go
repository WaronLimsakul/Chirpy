package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/WaronLimsakul/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	// atomic type used when keeping track something across go routine
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	tokenSecret    string
	polkaKey       string
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type LoggedInUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func main() {
	state := apiConfig{}
	godotenv.Load() // load first so os can access.

	envPlatform := os.Getenv("PLATFORM")
	state.platform = envPlatform

	envTokenSecret := os.Getenv("TOKEN_SECRET")
	state.tokenSecret = envTokenSecret

	envPolkaKey := os.Getenv("POLKA_KEY")
	state.polkaKey = envPolkaKey

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// sqlc create this database package
	// this function just connect the db to the queries
	dbQueries := database.New(db)
	state.dbQueries = dbQueries

	// servemux is like a server assistant
	// - remember which request should go where
	serveMux := http.NewServeMux()

	// - Handle("/app/", handler) means catching a path /app (trailing / = and everything under it)
	// Note that if many paths catch, it goes to the most specific one
	// - FileServer(root) will give any file request wants according to its url,
	//   we just need to give it root where to start
	// - Dir("dir") means we will set directory "dir" as a root for serving what req wants
	// - StripPrefix("/app", handler) maens we will strip prefix "haha" before send to request
	// - e.g. req: /app/Ron.html -> strip: /Ron.html -> serve file: get dir/Ron.html
	rootFileServer := http.FileServer(http.Dir("./"))
	serveMux.Handle("/app/", http.StripPrefix("/app", state.middlewareMetricsInc(rootFileServer)))

	// the http.ResponseWriter is a portal to communicate with the response we will send
	serveMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	serveMux.HandleFunc("GET /admin/metrics", state.reportFileServerHits)
	serveMux.HandleFunc("POST /admin/reset", state.resetServer)

	// serveMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	serveMux.HandleFunc("POST /api/chirps", state.createChirp)
	serveMux.HandleFunc("GET /api/chirps", state.getAllChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirp_id}", state.getChirpByID)   // {?} is a wildcard
	serveMux.HandleFunc("DELETE /api/chirps/{chirp_id}", state.deleteChirp) // {?} is a wildcard

	serveMux.HandleFunc("POST /api/users", state.createUser)
	serveMux.HandleFunc("PUT /api/users", state.updateUser)
	serveMux.HandleFunc("POST /api/login", state.loginUser)
	serveMux.HandleFunc("POST /api/refresh", state.refreshUser)
	serveMux.HandleFunc("POST /api/revoke", state.revokeToken)

	serveMux.HandleFunc("POST /api/polka/webhooks", state.reddenUser)

	server := &http.Server{Handler: serveMux, Addr: ":8080"}

	log.Printf("Listen to port 8080\n")
	log.Fatal(server.ListenAndServe())
}
