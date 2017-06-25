package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/gorilla/mux"
)

// GetProjectServiceLogs retrieves service logs for a project
func GetProjectServiceLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "streams not supported")
		return
	}
	// traefik listens for this content type to push streams.
	w.Header().Set("Content-Type", "text/event-stream")

	c, err := docker.NewClient(os.Getenv("DOCKER_REMOTE_API_URL"), "1.29", nil, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not get docker client: %v\n", err)
		return
	}

	p := mux.Vars(r)["project"]
	env := r.URL.Query().Get("env")
	s := mux.Vars(r)["service"]
	sID := fmt.Sprintf("%v_%v_%v", p, env, s)
	rc, err := c.ServiceLogs(r.Context(), sID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
		Tail:       "20",
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not get logs from docker: %v\n", err)
		return
	}

	wf := ioutils.NewWriteFlusher(w)
	defer wf.Close()

	io.Copy(wf, rc)
}
