package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
)

/**
 * Starts the API endpoint webservice
 */
func StartHttpApi() error {
	log.Printf("API service available on port %s.", HARE_API_PORT_PUBLISHING)

	http.HandleFunc("/", apiRequestPublish)
	http.HandleFunc("/stats", apiRequestStats)
	http.HandleFunc("/alive", apiRequestAlive)
	http.HandleFunc("/exit", apiRequestExit)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	err := http.ListenAndServe(":"+HARE_API_PORT_PUBLISHING, nil)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Exits hare upon receiving a /exit API call
 */
func apiRequestExit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Exit command received by %s.", r.RemoteAddr)
	os.Exit(2)
}

/**
 * Responds with a success health-check body, 200 HTTP status code upon receiving an /alive API call
 */
func apiRequestAlive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Alive")
	w.WriteHeader(http.StatusOK)
	return
}