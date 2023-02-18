package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/see-air-uh/finn-log-pose/auth"
	"github.com/see-air-uh/finn-log-pose/logs"
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

type BalanceResponse struct {
	Error   bool    `json:"error"`
	Message string  `json:"message"`
	Data    float32 `json:"data"`
}

type UpdateBalancePayload struct {
	TransactionAmount float32 `json:"transactionAmount"`
}

type Transaction struct {
	TransactionID          int     `json:"-"`
	UserID                 string  `json:"id"`
	TransactionAmount      float32 `json:"transactionAmount"`
	TransactionName        string  `json:"transactionName"`
	TransactionDescription string  `json:"transactionDescription"`
}

type TransactionResponse struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Data    []Transaction `json:"data"`
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

func (app *Config) logmessage(name string, msg string) error {
	conn, err := grpc.Dial(app.InspectorGadget, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("error. couldn't connect to logger")
	}
	defer conn.Close()

	l := logs.NewLogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = l.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: name,
			Data: msg,
		},
	})

	if err != nil {
		return fmt.Errorf("error. error when sending log, %s", err)
	}
	return nil
}

func (app *Config) create_user(w http.ResponseWriter, cruPayload CreateUserPayload) {

	conn, err := grpc.Dial(app.Ditto, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()
	a := auth.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	cruResp, err := a.CreateUser(ctx, &auth.CreateUserRequest{
		ArgUser: &auth.M_User{
			Email:     cruPayload.Email,
			Username:  cruPayload.Username,
			FirstName: cruPayload.FirstName,
			LastName:  cruPayload.LastName,
		},
		Password: cruPayload.Password,
	})

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.logmessage("created user", "created user with username "+cruResp.Username)

	payload := jsonResponse{
		Error:   false,
		Message: "created user " + cruResp.Username,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}

// a function that utilizes a GRPC conneciton to the auth
// service to validify an email and a password
func (app *Config) authenticate(w http.ResponseWriter, authPayload AuthPayload) {
	conn, err := grpc.Dial(app.Ditto, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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

	// logmessage("authed user", "authed user"+authResp.Username)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	// logmessage("authenticated user", fmt.Sprintf("authenticated user with username ", authResp.Username))
	payload := jsonResponse{
		Error:   false,
		Message: "authenticated user " + authResp.Username,
		Data:    authResp.PasetoToken,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}
func (app *Config) GetBalance(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user"){
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}


	resp, err := http.Get(fmt.Sprintf("%s/balance/%s",app.Mrkrabs, username))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data BalanceResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}
// TODO: Create a function that authenticates, then updates the balance
func (app *Config) UpdateBalance(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user"){
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	var requestPayload struct {
		TransactionAmount      float32 `json:"transactionAmount"`
		TransactionName        string  `json:"transactionName"`
		TransactionDescription string  `json:"transactionDescription"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	// values := map[string]float32{"transactionAmount": requestPayload.TransactionAmount, }
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%s/balance/%s",app.Mrkrabs, username), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data BalanceResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}
func (app *Config) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if u != r.Header.Get("user"){
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/transaction/%s",app.Mrkrabs ,u))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var trans TransactionResponse
	err = decoder.Decode(&trans)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, trans)
}
