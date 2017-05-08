package main

import (
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("Starting api-skelenot")
	NewRouter()
}

func NewRouter() {
	//create new router
	router := mux.NewRouter().StrictSlash(false)
	//api index init
	router.HandleFunc("/", Index)
	//start serving router
	http.Handle("/", router)

	srv := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	//start server and fail with error in logs if it fails
	log.Fatal(srv.ListenAndServe())

}

//Index endpoit for app server
func Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("/ endpoint called")
	w.Write([]byte("OK"))
}
