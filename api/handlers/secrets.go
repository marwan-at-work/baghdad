package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/bus"
)

type secretPostBody struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// AddSecret will add a secret to the swarm clustor accessible only to the
// targeted project.
func AddSecret(b bus.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bd, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s := secretPostBody{}
		json.Unmarshal(bd, &s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sj := baghdad.SecretsJob{
			ProjectName: mux.Vars(r)["project"],
			SecretName:  s.Name,
			SecretBody:  []byte(s.Body),
		}

		msg, _ := baghdad.EncodeSecretsJob(sj)
		fmt.Println("about to publish secrets")
		err = b.Publish("secrets", msg)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
