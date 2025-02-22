package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

// need to be a method of *apiConfig so we can update
// value of the original struct (pass by reference)
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
