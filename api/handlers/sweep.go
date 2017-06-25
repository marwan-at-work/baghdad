package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/marwan-at-work/baghdad/bus"
)

// Sweep sends a sweep job to all docker-sweeper workers. Effectively removing
// all stopped containers, volumes and images that have no running containers.
// use with caution.
func Sweep(b bus.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fw, ok := w.(http.Flusher); ok {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintln(w, "this call will erase all stopped containers and non-running images.\nyou have 5 seconds to stop this process by pressing CTL+C")
			fw.Flush()
			ch := time.Tick(time.Second)
			for i := 5; i > 0; i-- {
				<-ch
				fmt.Fprintln(w, i)
				fw.Flush()
			}

			select {
			case <-r.Context().Done():
				return
			default:
			}

			err := b.Broadcast("sweep", "", nil)
			if err != nil {
				fmt.Fprintf(w, "could not send sweep job: %v\n", err)
				return
			}

			fmt.Fprintln(w, "done")
		} else {
			err := b.Broadcast("sweep", "", nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}
}
