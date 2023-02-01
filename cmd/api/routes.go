package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Logger)

	mux.Use(middleware.Heartbeat("/ping"))

	// a single Point of entry that is short form for "sixty six"
	// this is so every request will be executing request 66
	mux.Post("/ss", app.ExecuteRequest)

	// add route for get and post balance
	mux.Get("/balance/{user}", app.GetBalance)
	mux.Post("/balance/{user}", app.UpdateBalance)

	mux.Get("/transaction/{user}", app.GetAllTransactions)

	return mux
}
