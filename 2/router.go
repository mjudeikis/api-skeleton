package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// NewRouter start new router
func NewRouter() {

	//create new router
	router := mux.NewRouter().StrictSlash(false)

	//api health init
	router.Methods("GET").Path("/health").Name("Health endpoint").HandlerFunc(http.HandlerFunc(Health))
	//this does not work as we hit "mux" limitation for serving trafic... go to point 3

	//api backend init
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	apiV1.Methods("GET").Path("/").Name("Index").Handler(http.HandlerFunc(Index))

	srv := &http.Server{
		Handler: apiV1,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

}

func Health(w http.ResponseWriter, r *http.Request) {
	log.Debug("/health endpoint called")
	w.Write([]byte("OK"))
}

//Index endpoit for app server
func Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("/ endpoint called")
	w.Write([]byte("OK"))
}
