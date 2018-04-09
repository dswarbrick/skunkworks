package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func uploadPut(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s request, content-length: %d\n", r.Proto, r.Method, r.ContentLength)
	log.Printf("Request: %#v\n", r)
}

func uploadHead(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s request\n", r.Proto, r.Method)
	log.Printf("Request: %#v\n", r)

	w.Header().Set("X-File-Size", "12345")
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/upload", uploadPut).Methods("PUT")
	r.HandleFunc("/upload", uploadHead).Methods("HEAD")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
