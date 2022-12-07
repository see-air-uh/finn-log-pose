package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/see-air-uh/finn-log-pose/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action      string            `json:"action"`
	Auth        AuthPayload       `json:"auth,omitempty"`
	CreateUser  CreateUserPayload `json:"createuser,omitempty"`
	PasetoToken string            `json:"paseto,omitempty"`
}

type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserPayload struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
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
	case "createuser":
		app.create_user(w, requestPayload.CreateUser)
	case "auth":
		app.authenticate(w, requestPayload.Auth)

	default:
		app.errorJSON(w, fmt.Errorf("error. invalid option"))
	}

}

func (app *Config) create_user(w http.ResponseWriter, cruPayload CreateUserPayload) {
	log.Println("Done...")
	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	a := auth.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	log.Println("Done...")
	cruResp, err := a.CreateUser(ctx, &auth.CreateUserRequest{
		ArgUser: &auth.M_User{
			Email:     cruPayload.Email,
			Username:  cruPayload.Username,
			FirstName: cruPayload.FirstName,
			LastName:  cruPayload.LastName,
		},
		Password: cruPayload.Password,
	})

	log.Println("Done...")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	//  else if !authResp.Authed {
	// 	payload := jsonResponse{
	// 		Error:   true,
	// 		Message: "failed to authenticate " + authPayload.Username,
	// 	}
	// 	app.writeJSON(w, http.StatusUnauthorized, payload)
	// 	return
	// }

	payload := jsonResponse{
		Error:   false,
		Message: "created user " + cruResp.Username,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}

// a function that utilizes a GRPC conneciton to the auth
// service to validify an email and a password
func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {
	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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
			Password: authPayload.Password,

			Username: &authPayload.Username,
		},
	})

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	//  else if !authResp.Authed {
	// 	payload := jsonResponse{
	// 		Error:   true,
	// 		Message: "failed to authenticate " + authPayload.Username,
	// 	}
	// 	app.writeJSON(w, http.StatusUnauthorized, payload)
	// 	return
	// }

	payload := jsonResponse{
		Error:   false,
		Message: "authenticated user " + authResp.Username,
		Data:    authResp.PasetoToken,
	}
	log.Println("DOWN HERE")
	app.writeJSON(w, http.StatusAccepted, payload)
}
