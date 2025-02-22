package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/WaronLimsakul/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	// atomic type used when keeping track something across go routine
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	tokenSecret    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// http.HandlerFunc is a function type that has method ServeHTTP -> call itself, lol
	// the trick is called self-promotion = you are a functio but wanna be interface ?
	// -> call yourself!
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) reportFileServerHits(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	// need to do .Load() to get the stored value
	body := []byte(fmt.Sprintf(`
        <html>
            <body>
                <h1>Welcome, Chirpy Admin</h1>
                <p>Chirpy has been visited %d times!</p>
            </body>
        </html>
        `, cfg.fileServerHits.Load()))
	w.Write(body)
}

// 0. check if server platform is 'dev'
// 1. reset file server hits
// 2. reset users data
func (cfg *apiConfig) resetServer(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		log.Println("not development env")
		w.WriteHeader(403)
		return
	}

	cfg.fileServerHits.Store(0)

	err := cfg.dbQueries.ResetUser(req.Context())
	if err != nil {
		log.Println("error reseting user data")
		w.WriteHeader(500)
		return
	}

	err = cfg.dbQueries.ResetChirp(req.Context())
	if err != nil {
		log.Println("error reseting chirp data")
		w.WriteHeader(500)
		return
	}

	err = cfg.dbQueries.ResetRefreshTokens(req.Context())
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Server reset"))
}

func main() {
	state := apiConfig{}
	godotenv.Load() // load first so os can access.

	envPlatform := os.Getenv("PLATFORM")
	state.platform = envPlatform

	envTokenSecret := os.Getenv("TOKEN_SECRET")
	state.tokenSecret = envTokenSecret

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
	serveMux.HandleFunc("GET /api/chirps/{chirp_id}", state.getChirpByID) // {?} is a wildcard

	serveMux.HandleFunc("POST /api/users", state.createUser)
	serveMux.HandleFunc("POST /api/login", state.loginUser)
	serveMux.HandleFunc("POST /api/refresh", state.refreshUser)
	serveMux.HandleFunc("POST /api/revoke", state.revokeToken)

	server := &http.Server{Handler: serveMux, Addr: ":8080"}

	log.Printf("Listen to port 8080\n")
	log.Fatal(server.ListenAndServe())
}
