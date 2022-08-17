package main

import (
	"context"
	"net/http"
	"time"

	"github.com/see-air-uh/asxce-log-pose/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Username string `json:"username"`
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
func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {

	conn, err := grpc.Dial("toga:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer conn.Close()

	a := auth.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	authResp, err := a.AuthUser(ctx, &auth.AuthRequest{
		ArgUser: &auth.User{
			Username: authPayload.Username,
			Password: authPayload.Password,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	} else if !authResp.Authed {
		payload := jsonResponse{
			Error:   true,
			Message: "failed to authenticate " + authPayload.Username,
		}
		app.writeJSON(w, http.StatusUnauthorized, payload)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: "authenticated " + authPayload.Username,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}
