package main

import "net/http"

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) ExecuteRequest(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// switch on different actions requested
	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	}
}

// a function that utilizes a GRPC conneciton to the auth
// service to validify an email and a password
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {

}
