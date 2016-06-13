package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var (
	web        = flag.String("web", "web/dist", "Web UI directory")
	signingKey = []byte("yu786lklgfso32921lkasaskdhladsyg6")
	users      = map[string]string{
		"mbaiju": "qwerty123",
		"somkar": "asdfgh123",
	}
)

func TokenAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var pl map[string]string

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&pl)
	username := pl["username"]
	password := pl["password"]

	if passwd, found := users[username]; found {
		if password == passwd {
			token := jwt.New(jwt.GetSigningMethod("HS256"))
			// FIXME: get the username from the request
			token.Claims["sub"] = username
			token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
			// Sign and get the complete encoded token as a string
			tokenString, _ := token.SignedString(signingKey)

			s := map[string]string{"token": tokenString}

			out, _ := json.Marshal(s)
			w.Write(out)
			w.WriteHeader(http.StatusOK)
		}
	}
	w.WriteHeader(http.StatusUnauthorized)

}

func main() {
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/api/token-auth/", TokenAuth).Methods("POST")
	n := negroni.New(negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(*web)))
	n.UseHandler(r)
	n.Run(":7080")

}
