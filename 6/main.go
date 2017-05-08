package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
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

func test() {

	fmt.Print("test")
}
