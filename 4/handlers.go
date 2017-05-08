package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const NotFound = "Not found"

//Health endpoit for app server
func Health(w http.ResponseWriter, r *http.Request) {
	log.Debug("/health endpoint called")
	w.Write([]byte("OK"))
}

//Index endpoit for app server
func Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("/ endpoint called")
	w.Write([]byte("OK"))
}

func CheckAuth(resp http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	log.Info("Check auth Middleware")
	user, pass, _ := req.BasicAuth()
	enabled := true
	log.WithFields(log.Fields{
		"auth": enabled,
	}).Debug("handler")
	if enabled && !checkPass(user, pass) {
		reason := "Unauthorized"
		resp.WriteHeader(http.StatusUnauthorized)
		response(reason, true, nil, resp, req)
		return
	}
	next(resp, req)
}

func checkPass(user, pass string) bool {
	log.Info(fmt.Sprintf("User [%s] and Pass [%s]", user, pass))
	if user == "admin" && pass == "admin" {
		log.Info("Pass OK")
		return true
	} else {
		log.Info("Pass Error")
		return false
	}
	return false
}

func CorsHeadersMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	log.Info("Cors Middleware")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	rw.Header().Set("Access-Control-Expose-Headers", "Authorization")
	rw.Header().Set("Access-Control-Request-Headers", "Authorization")

	if r.Method == "OPTIONS" {
		rw.WriteHeader(200)
		return
	}

	next(rw, r)
}

func response(obj interface{}, prettyPrint bool, err error, resp http.ResponseWriter, req *http.Request) {
	// Check for an error
HAS_ERR:
	if err != nil {

		if err.Error() == NotFound {
			resp.WriteHeader(http.StatusNotFound)
			return
		}

		log.WithFields(log.Fields{
			"error":  err,
			"method": req.Method,
			"url":    req.URL,
		}).Error("request error")

		code := 500
		errMsg := err.Error()
		if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "ACL not found") {
			code = 403
		}
		resp.WriteHeader(code)
		resp.Write([]byte(err.Error()))
		return
	}

	// Write out the JSON object
	if obj != nil {
		buf, err := marshall(obj, true)
		if err != nil {
			goto HAS_ERR
		}
		resp.Header().Set("Content-Type", "application/json")

		// encoding/json library has a specific bug(feature) to turn empty slices into json null object,
		// let's make an empty array instead
		if string(buf) == "null" {
			buf = []byte("[]")
		}
		resp.Write(buf)
	}
}

// marshall returns a json byte slice, leaving existing json untouched.
func marshall(obj interface{}, pretty bool) ([]byte, error) {

	var js interface{}
	var buf []byte

	// Only check objects that byte slices and strings for valid json
	switch v := obj.(type) {
	case []byte:
		buf = []byte(v)
	case string:
		buf = []byte(v)
	}

	// If we were given a valid json object, return it as-is
	if buf != nil && json.Unmarshal(buf, &js) == nil {
		return buf, nil
	}

	// Otherwise marshall the object into json
	if pretty {
		return json.MarshalIndent(obj, "", "    ")
	}
	return json.Marshal(obj)
}
