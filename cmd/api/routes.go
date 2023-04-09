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
		r.Get("/balance/{user}/{account}", app.GetBalance)
		r.Post("/balance/{user}/{account}", app.UpdateBalance)

		r.Get("/accounts/{user}", app.GetAccounts)
		r.Post("/accounts/add/{user}/{account}", app.AddAccount)
		r.Post("/accounts/add_user/{user}/{account}/{user2}", app.AddUserToAccount)

		r.Get("/recurring/{user}/{account}", app.GetReccurringPayments)
		r.Post("/recurring/add/{user}/{account}", app.AddReccurringPayment)
		r.Get("/recurring/history/{user}/{recurring_id}", app.GetPaymentHistory)

		r.Get("/transaction/{user}/{account}", app.GetAllTransactions)
		r.Get("/transaction/{user}/{account}/category/{category}", app.GetAllTransactionsOfCategory)

		r.Get("/transaction/{user}/{account}/category", app.GetCategories)
		r.Post("/transaction/{user}/{account}/category", app.UpdateTransactionCategory)

		r.Get("/debt/{user}/{account}", app.GetAllDebts)
		r.Post("/debt/{user}/{account}", app.CreateDebt)
		r.Get("/debt/{user}/{account}/{debtID}", app.GetDebtByID)
		r.Post("/debt/{user}/{account}/{debtID}", app.MakeDebtPayment)
	})

	// a single Point of entry that is short form for "sixty six"
	// this is so every request will be executing request 66
	mux.Route("/ss", func(r chi.Router) {
		r.Post("/", app.ExecuteRequest)
	})

	return mux
}
