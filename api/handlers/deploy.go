package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/marwan-at-work/baghdad"
	"github.com/marwan-at-work/baghdad/bus"
	"github.com/marwan-at-work/baghdad/utils"
)

type deployPostBody struct {
	Branch string `json:"branch"`
	Env    string `json:"environment"`
	Tag    string `json:"tag"`
}

// Deploy takes a deploy request, and if valid, forwrds it to the "deploy" queue.
func Deploy(b bus.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bts, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var d deployPostBody
		if err = json.Unmarshal(bts, &d); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p := mux.Vars(r)["project"]
		o := mux.Vars(r)["owner"]
		bgd, err := utils.GetBaghdad(utils.GetGithub(os.Getenv("ADMIN_TOKEN")), utils.GetBaghdadOpts{
			Ctx:      context.Background(),
			SHA:      d.Tag,
			Owner:    o,
			RepoName: p,
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dj := baghdad.DeployJob{
			Baghdad:    bgd,
			BranchName: d.Branch,
			Env:        d.Env,
			RepoName:   p,
			RepoOwner:  o,
			Tag:        d.Tag,
		}
		msg, _ := json.Marshal(dj)
		b.Publish("deploy-sync", msg)
	}
}
