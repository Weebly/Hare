package hare

import (
	"log"
	"net/http"
)

/**
 * Starts the API endpoint webservice
 */
func StartHttpApi() error {
	log.Printf("Publishing service available on port %s.", HARE_API_PORT_PUBLISHING)
	log.Print("Statistics service available at /stats.")
	http.HandleFunc("/", apiRequestPublish)
	http.HandleFunc("/stats", apiRequestStats)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	err := http.ListenAndServe(":"+HARE_API_PORT_PUBLISHING, nil)
	if err != nil {
		return err
	}
	log.Print("shouldnt be here")
	return nil
}

/**
 * Prepares and dispatches a response via ResponseWriter for an exit request
 */
func apiRequestExit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Exit command received by %s.", r.RemoteAddr)
	//  os.Exit(2)
}
