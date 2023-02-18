package main

import (
	"context"
	"net/http"
	"time"

	"github.com/see-air-uh/finn-log-pose/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AuthorizeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request_token := r.Header.Get("Token")

		r.Header.Del("user")

		// reject if no token was supplied in the header
		if(request_token == ""){
			authFailed(w)
			return
		}

		// check if token is correct
		conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			connFailed(w)
			return
		}
		defer conn.Close()

		// create new connection to auth server
		a := auth.NewAuthServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		token_check_response, err := a.CheckToken(ctx, &auth.CheckTokenRequest{
			PasetoToken: request_token,
		})

		// if token is invalid, reject request
		if err != nil {
			authFailed(w)
			return
		}

		// append user to request header for rest of application to use
		r.Header.Add("user", token_check_response.GetUsername())
		next.ServeHTTP(w, r)
	})
}

func authFailed(w http.ResponseWriter) {
	// w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}

func connFailed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}