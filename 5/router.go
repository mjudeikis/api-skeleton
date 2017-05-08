package main

import (
	"net/http"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mangirdaz/api-skeleton/config"
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

	url := fmt.Sprintf("%s:%s", config.Get("EnvAPIIP"), config.Get("EnvAPIPort"))
	log.Debugf("api: starting api server on %s", url)
	log.Fatal(http.ListenAndServe(url, n))

}
