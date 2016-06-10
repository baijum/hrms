package main

import (
	"flag"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var web = flag.String("web", "web/dist", "Web UI directory")

func TokenAuth(w http.ResponseWriter, r *http.Request) {
	// TODO: Complete this function
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
