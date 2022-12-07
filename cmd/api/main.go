package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct {
}

func main() {

	// define app type
	app := Config{}

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start http server
	log.Printf("Starting log-pose on port %s", webPort)

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}
