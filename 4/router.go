package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
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

	//init new mux server
	midd := http.NewServeMux()
	//add multipe handlers
	midd.Handle("/", router)

	//add autCorsHeaderMiddleware
	midd.Handle("/api/v1", negroni.New(
		negroni.HandlerFunc(CorsHeadersMiddleware),
		negroni.HandlerFunc(CheckAuth),
		negroni.Wrap(apiV1),
	))

	n := negroni.Classic()

	//add all handers under negroni managment
	n.UseHandler(midd)

	log.Debug("api: starting api server")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", n))

}
