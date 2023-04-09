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
	TransactionID          int     `json:"transaction_id"`
	UserID                 string  `json:"user_id"`
	TransactionAmount      float32 `json:"transactionAmount"`
	TransactionName        string  `json:"transactionName"`
	TransactionDescription string  `json:"transactionDescription"`
	TransactionCategory    string  `json:"transactionCategory"`
}

type TransactionResponse struct {
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Data    []Transaction `json:"data"`
}

type AccountsResponse struct {
	Error   bool      `json:"error"`
	Message string    `json:"message"`
	Data    []Account `json:"data"`
}

type Account struct {
	AccountID   int    `json:"id"`
	AccountName string `json:"accountname"`
	Email       string `json:"email"`
	IsPrimary   bool   `json:"isprimary"`
}

type AddAccountResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type AddUserToAccountResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type RecurringPayment struct {
	PaymentID          int     `json:"paymentid"`
	UserName           string  `json:"username"`
	PaymentAmount      float32 `json:"amount"`
	PaymentName        string  `json:"paymentName"`
	PaymentDescription string  `json:"paymentDescription"`
	PaymentDate        string  `json:"paymentDate"`
	PaymentType        string  `json:"paymentType"`
}

type RecurringPaymentResponse struct {
	Error   bool               `json:"error"`
	Message string             `json:"message"`
	Data    []RecurringPayment `json:"data"`
}

type PaymentHistory struct {
	PaymentHistoryID     int    `json:"paymenthistoryid"`
	PaymentID            int    `json:"paymentID"`
	PaymentHistoryDate   string `json:"paymentHistoryDate"`
	PaymentHistoryStatus bool   `json:"paymentHistoryStatus"`
}

type PaymentHistoryPayload struct {
	Error   bool             `json:"error"`
	Message string           `json:"message"`
	Data    []PaymentHistory `json:"data"`
}

type DebtResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    []Debt `json:"data"`
}

type Debt struct {
	DebtID            int     `json:"debtID"`
	UserID            string  `json:"user_id"`
	TotalOwing        float32 `json:"total_owing"`
	TotalDebtPayments float32 `json:"total_payments"`
}

type DebtPayment struct {
	PaymentID     int `json:"payment_id"`
	TransactionID int `json:"transaction_id"`
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
	if app.DisableLogger == "true" {
		return nil
	}
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
	fmt.Print("Here")
	conn, err := grpc.Dial(app.Ditto, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	fmt.Print("THere")
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
	account := chi.URLParam(r, "account")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%sbalance/%s/%s", app.Mrkrabs, username, account))
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
	account := chi.URLParam(r, "account")
	fmt.Print("before here payload\n")
	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}
	fmt.Print("before reading payload\n")
	var requestPayload struct {
		TransactionAmount      float32 `json:"transactionAmount"`
		TransactionName        string  `json:"transactionName"`
		TransactionDescription string  `json:"transactionDescription"`
		TransactionCategory    string  `json:"transactionCategory"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	fmt.Print("after reading payload\n")
	// values := map[string]float32{"transactionAmount": requestPayload.TransactionAmount, }
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%sbalance/%s/%s", app.Mrkrabs, username, account), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	fmt.Print("after api ping\n")
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

func (app *Config) GetAccounts(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%saccounts/%s", app.Mrkrabs, username))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data AccountsResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) AddAccount(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Post(fmt.Sprintf("%saccounts/add/%s/%s", app.Mrkrabs, username, account), "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data AddAccountResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) AddUserToAccount(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")
	u2 := chi.URLParam(r, "user2")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Post(fmt.Sprintf("%saccounts/add/%s/%s/%s", app.Mrkrabs, username, account, u2), "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data AddUserToAccountResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) GetReccurringPayments(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%srecurring/%s/%s", app.Mrkrabs, username, account))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data RecurringPaymentResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) AddReccurringPayment(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}
	var requestPayload struct {
		PaymentAmount      float32 `json:"amount"`
		PaymentName        string  `json:"paymentName"`
		PaymentDescription string  `json:"paymentDescription"`
		PaymentDate        string  `json:"paymentDate"`
		PaymentType        string  `json:"paymentType"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%srecurring/add/%s/%s", app.Mrkrabs, u, account), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var data struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    int32  `json:"data"`
	}

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, data)

}

func (app *Config) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	r_id := chi.URLParam(r, "recurring_id")

	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%srecurring/history/%s/%s", app.Mrkrabs, u, r_id))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data PaymentHistoryPayload

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, data)

}

func (app *Config) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%stransaction/%s/%s", app.Mrkrabs, u, account))
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
func (app *Config) GetAllTransactionsOfCategory(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")
	c := chi.URLParam(r, "category")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/transaction/%s/%s/category/%s", app.Mrkrabs, u, account, c))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var transactions TransactionResponse
	err = decoder.Decode(&transactions)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, transactions)

}
func (app *Config) UpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")

	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}
	var requestPayload struct {
		TransactionID int    `json:"transactionID"`
		Category      string `json:"transactionCategory"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%s/transaction/%s/category", app.Mrkrabs, u), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var data struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Updated Category")

}

func (app *Config) GetCategories(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%s/transaction/%s/category", app.Mrkrabs, u))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	/*
			    "error": false,
		    "message": "successfully grabbed all categories for user2",
		    "data": [
		        "",
		        "Scuba diving"
		    ]
	*/

	var categories struct {
		Error   bool     `json:"error"`
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	fmt.Print(categories)
	fmt.Print("\ncategories\n")
	fmt.Print("\n")
	err = decoder.Decode(&categories)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, categories.Data)
}

func (app *Config) GetAllDebts(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%sdebt/%s/%s", app.Mrkrabs, username, account))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data DebtResponse

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) CreateDebt(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")

	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	var requestPayload struct {
		TotalOwing float32 `json:"total_owing"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%sdebt/%s/%s", app.Mrkrabs, u, account), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var data struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    int    `json:"data"`
	}

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Added Debt")

}

func (app *Config) GetDebtByID(w http.ResponseWriter, r *http.Request) {

	username := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")
	debtIDString := chi.URLParam(r, "debtID")

	// check if username passed in is same for the requested user balance
	// TODO: Change this if statement to get the users associated with this account
	if username != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	resp, err := http.Get(fmt.Sprintf("%sdebt/%s/%s/%s", app.Mrkrabs, username, account, debtIDString))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    Debt   `json:"data"`
	}

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, data)
}

func (app *Config) MakeDebtPayment(w http.ResponseWriter, r *http.Request) {
	u := chi.URLParam(r, "user")
	account := chi.URLParam(r, "account")
	debtIDString := chi.URLParam(r, "debtID")

	if u != r.Header.Get("user") {
		w.WriteHeader(http.StatusUnauthorized)
		app.errorJSON(w, fmt.Errorf("error. Insufficient access"))
		return
	}

	var requestPayload struct {
		Amount float32 `json:"amount"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	json_data, err := json.Marshal(&requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp, err := http.Post(fmt.Sprintf("%sdebt/%s/%s/%s", app.Mrkrabs, u, account, debtIDString), "application/json", bytes.NewBuffer((json_data)))

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var data struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}

	err = decoder.Decode(&data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Updated debt")

}
