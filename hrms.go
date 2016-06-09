package main

import (
	"flag"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var web = flag.String("web", "web/dist", "Web UI directory")

func main() {
	flag.Parse()
	r := mux.NewRouter()
	n := negroni.New(negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(*web)))
	n.UseHandler(r)
	n.Run(":7080")
}
