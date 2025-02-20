package main

import (
	"log"
	"net/http"
)

func main() {

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
	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./"))))

	// the http.ResponseWriter is a portal to communicate with the response we will send
	serveMux.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server := &http.Server{Handler: serveMux, Addr: ":8080"}

	// do the strip things here.

	log.Printf("Listen to port 8080\n")
	log.Fatal(server.ListenAndServe())
}
