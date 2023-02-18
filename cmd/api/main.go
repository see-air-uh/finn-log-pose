package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// const webPort = "80"

type Config struct {
	Environment string
	Ditto string
	Mrkrabs string
	InspectorGadget string
	DisableLogger string
	WebPort string
}

func main() {
	// check if there is an environment variable already set to production
	environment := os.Getenv("ENVIRONMENT")

	if environment == "" {
		// TODO: Load .env file
		log.Println("Loading environment variables from local .env file")
		godotenv.Load(".env")
	}

	// define app type
	app := setupConfig()

	fmt.Printf("%+v\n", app)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.WebPort),
		Handler: app.routes(),
	}

	// start http server
	log.Printf("Starting log-pose on port %s", app.WebPort)

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

// setup config with env variables
func setupConfig() *Config{

	return &Config{
		Environment: os.Getenv("ENVIRONMENT"),
		Ditto: os.Getenv("DITTO_CONNECTION_STRING"),
		Mrkrabs: os.Getenv("MRKRABS_CONNECTION_STRING"),
		InspectorGadget: os.Getenv("INSPECTOR_GADGET_CONNECTION_STRING"),
		DisableLogger: os.Getenv("DISABLE_LOGGER"),
		WebPort: os.Getenv("WEB_PORT"),
	}

}