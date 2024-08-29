package main

import (
	"log"
	"net/http"

	"github.com/DemmyDemon/authplaceholder"
)

const bind = `127.0.0.1:8080`

const examplePage = `<!DOCTYPE html>
<html>
	<head>
		<title>Password protected prototype</title>
	</head>
	<body>
		<div>You totally got in!</div>
		<div><a href="/logout">Hop back out?</a></div>
	</body>
</html>`

func main() {
	mux := http.NewServeMux()
	auth, err := authplaceholder.New(mux, "example.json")
	if err != nil {
		log.Fatal("Oh gods, we totally just... " + err.Error())
	}
	mux.HandleFunc("GET /", auth.Wrap(TotallyUsefulHandler))

	log.Println("Starting went fine, will listen on http://" + bind + "/")
	if err := http.ListenAndServe(bind, mux); err != nil {
		log.Fatal("Serving ends: " + err.Error())
	}
}

// TotallyUsefulHandler does very useful things in this totally real and viable prototype.
func TotallyUsefulHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(examplePage))
}
