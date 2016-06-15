package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	web        = flag.String("web", "web/dist", "Web UI directory")
	signingKey = []byte("yu786lklgfso32921lkasaskdhladsyg6")
	empdata    = "/tmp/emp.db"
	users      = map[string]string{
		"mbaiju": "qwerty123",
		"somkar": "asdfgh123",
	}
)

// TokenAuth generate token for valid users
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

// Employee represent an employee
type Employee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func keyFunc(*jwt.Token) (interface{}, error) {
	return signingKey, nil
}

// AddEmployee add employee details to the database
func AddEmployee(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.ParseFromRequest(r, keyFunc)

	fmt.Printf("Token: %#v\nError: %#v", token, err)
	if token.Valid {
		w.Header().Set("Content-Type", "application/json")
		var emp Employee

		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&emp)

		empline := fmt.Sprintf("id:%s;name:%s\n", emp.ID, emp.Name)

		out, _ := ioutil.ReadFile(empdata)

		ioutil.WriteFile(empdata, []byte(string(out)+empline), 0644)

		log.Printf("%#v\n", emp)

		w.Write([]byte("{}\n"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func main() {
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/api/token-auth/", TokenAuth).Methods("POST")
	r.HandleFunc("/api/employees", AddEmployee).Methods("POST")
	n := negroni.New(negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(*web)))
	n.UseHandler(r)
	n.Run(":7080")

}
