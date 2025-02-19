package main

import (
	"log"
	"net/http"
)

func main() {

	// servemux is like a server assistant
	// - remember which request should go where
	serveMux := http.NewServeMux()

	// FileServer give a handler
	// http.Dir give us a string that is FileSystem interface
	serveMux.Handle("/", http.FileServer(http.Dir("./")))

	// don't put the the fisrt / before path name, except it's a root
	serveMux.Handle("assets/", http.FileServer(http.Dir("assets/")))

	server := &http.Server{Handler: serveMux, Addr: ":8080"}

	log.Printf("Listen to port 8080\n")
	log.Fatal(server.ListenAndServe())
}
