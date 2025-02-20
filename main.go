package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	// atomic type used when keeping track something across go routine
	fileServerHits atomic.Int32
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

func (cfg *apiConfig) resetFileServerHits(w http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits.Store(0)
	w.WriteHeader(200)
	w.Write([]byte("file server hits reset"))
}

func main() {

	// servemux is like a server assistant
	// - remember which request should go where
	serveMux := http.NewServeMux()
	state := apiConfig{}

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
	serveMux.HandleFunc("POST /admin/reset", state.resetFileServerHits)

	serveMux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server := &http.Server{Handler: serveMux, Addr: ":8080"}

	log.Printf("Listen to port 8080\n")
	log.Fatal(server.ListenAndServe())
}
