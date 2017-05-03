package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/marwan-at-work/baghdad/api/handlers"
	"github.com/marwan-at-work/baghdad/bus"
	"github.com/marwan-at-work/baghdad/utils"
)

func main() {
	godotenv.Load("/run/secrets/baghdad-vars")
	utils.ValidateEnvVars(getRequiredVars()...)
	r := mux.NewRouter()

	b := bus.NewBroker(os.Getenv("AMQP_URL"))
	registerRoutes(r, b)

	fmt.Println("[baghdad-api] 👂  on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func registerRoutes(r *mux.Router, b bus.Publisher) {
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello from baghdad 👻")
	})

	r.HandleFunc("/hooks/github", handlers.GithubHook(b)).Methods(http.MethodPost)

	r.HandleFunc("/projects/{owner}/{project}/deploy", handlers.Deploy(b)).Methods(http.MethodPost)

	r.HandleFunc("/projects/{project}/secrets", handlers.AddSecret(b))
}
