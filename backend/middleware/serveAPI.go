package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// Serve an API using mux.Router().Host({domain}).Subrouter().
// Provide the router, domain, an array of endpoints and functions, and
// the port you would like the API accessible to.
func ServeApi(
	router *mux.Router,
	domain string,
	endpoint []string,
	function []func(http.ResponseWriter, *http.Request),
	port string,
) {
	// TODO: Add error handling
	apiR := router.Host(domain).Subrouter()

	for i := 0; i < len(endpoint); i++ {
		apiR.HandleFunc(endpoint[i], function[i])
	}

	apiSrv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      apiR,
	}

	go func() {
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	log.Println("API server is running on port " + port)
}
