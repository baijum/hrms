package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	jwtreq "github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/scrypt"
)

var (
	web        = flag.String("web", "web/dist", "Web UI directory")
	credential = flag.String("credential", "", "Credential with format: user:passwd")
	signingKey = []byte("yu786lklgfso32921lkasaskdhladsyg6")
	empdata    = "/tmp/emp.db"
	userdata   = "/tmp/users.db"
)

func getPasswdSalt(username string) (string, string) {
	out, _ := ioutil.ReadFile(userdata)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		text := strings.Split(line, ":")
		user, passwd, salt := text[0], text[1], text[2]
		if username == user {
			return passwd, salt
		}
	}
	return "", ""
}

// TokenAuth generate token for valid users
func TokenAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var pl map[string]string

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&pl)
	username := pl["username"]
	password := pl["password"]

	passwd, salt := getPasswdSalt(username)
	dk, _ := scrypt.Key([]byte(password), []byte(salt), 16384, 8, 1, 32)

	if passwd == string(dk) {
		claims := &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			Subject:   username,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		// Sign and get the complete encoded token as a string
		tokenString, _ := token.SignedString(signingKey)

		s := map[string]string{"token": tokenString}

		out, _ := json.Marshal(s)
		w.Write(out)
		w.WriteHeader(http.StatusOK)
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
	w.Header().Set("Content-Type", "application/json")
	var emp Employee

	ctx := r.Context()
	fmt.Printf("Context: %#v\n", ctx)

	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&emp)

	empline := fmt.Sprintf("id:%s;name:%s\n", emp.ID, emp.Name)

	out, _ := ioutil.ReadFile(empdata)

	ioutil.WriteFile(empdata, []byte(string(out)+empline), 0644)

	log.Printf("%#v\n", emp)

	w.Write([]byte("{}\n"))
}

// Authorize is a middleware for authorization
func Authorize(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token, err := jwtreq.ParseFromRequest(r, jwtreq.AuthorizationHeaderExtractor, keyFunc)

	fmt.Printf("Token: %#v\nError: %#v", token, err)
	if token.Valid {
		next(w, r)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func randomSalt() string {
	b := make([]byte, 20)
	rand.Read(b)
	return string(b)
}

func createUser(credential string) {
	textslice := strings.Split(credential, ":")
	username, passwd := textslice[0], textslice[1]

	salt := randomSalt()
	dk, _ := scrypt.Key([]byte(passwd), []byte(salt), 16384, 8, 1, 32)
	f, err := os.OpenFile(userdata, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(username + ":" + string(dk) + ":" + salt + "\n"); err != nil {
		panic(err)
	}

}

func main() {
	flag.Parse()
	if *credential != "" {
		createUser(*credential)
		os.Exit(1)
	}
	r := mux.NewRouter()

	ar := mux.NewRouter()

	r.HandleFunc("/api/token-auth/", TokenAuth).Methods("POST")
	ar.HandleFunc("/api/employees", AddEmployee).Methods("POST")
	r.PathPrefix("/api").Handler(
		negroni.New(negroni.HandlerFunc(Authorize), negroni.Wrap(ar)))
	n := negroni.New(negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(*web)))
	n.UseHandler(r)
	n.Run(":7080")

}
