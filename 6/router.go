package main

import (
	"net/http"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mangirdaz/api-skeleton/config"
	"github.com/mangirdaz/api-skeleton/kv"
	"github.com/urfave/negroni"
)

type Server struct {
	kv *kv.LibKVBackend
}

// NewRouter start new router
func NewRouter() {

	//create new router
	router := mux.NewRouter().StrictSlash(false)

	//create storage backend
	kv := config.InitKVStorage()
	//create server instance
	s := Server{kv: kv}

	//api health init
	router.Methods("GET").Path("/health").Name("Health endpoint").Handler(http.HandlerFunc(s.Health))
	//this does not work as we hit "mux" limitation for serving trafic... go to point 3

	//api backend init
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	apiV1.Methods("GET").Path("/").Name("Index").Handler(http.HandlerFunc(s.Index))
	apiV1.Methods("GET").Path("/counter").Name("Counter").Handler(http.HandlerFunc(s.Counter))

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
