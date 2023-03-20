package main

import (
	"log"
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

	mux.Route("/app", func(r chi.Router) {
		log.Println("entering in app??")
		r.Use(app.AuthorizeRequest)
		// add route for get and post balance
		r.Get("/balance/{user}", app.GetBalance)
		r.Post("/balance/{user}", app.UpdateBalance)

		r.Get("/transaction/{user}", app.GetAllTransactions)

		r.Get("/transaction/{user}/category", app.GetCategories)
		r.Post("/transaction/{user}/category", app.UpdateTransactionCategory)
	})

	// a single Point of entry that is short form for "sixty six"
	// this is so every request will be executing request 66
	mux.Route("/ss", func (r chi.Router) {
		r.Post("/", app.ExecuteRequest)
	})




	return mux
}
