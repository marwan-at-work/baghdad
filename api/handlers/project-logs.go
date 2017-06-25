package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/gorilla/mux"
	"github.com/marwan-at-work/baghdad/worker"
	"github.com/streadway/amqp"
)

// GetProjectLogs connect to rabbitmq's logs channel,
func GetProjectLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "streams not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	routingKey := mux.Vars(r)["project"]
	wrkr, msgs, err := getMsgs(routingKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "could not get rabbitmq msgs: %v", err)
		return
	}
	defer wrkr.Close()

	wf := ioutils.NewWriteFlusher(w)
	defer wf.Close()

	fmt.Fprintf(wf, "startings log stream for %v\n", routingKey)
	wf.Flush()

	ctx := r.Context()
	streaming := true
	for streaming {
		select {
		case <-ctx.Done():
			streaming = false
			break
		case msg, ok := <-msgs:
			if !ok {
				streaming = false
				break
			}

			if !strings.HasPrefix(msg.RoutingKey, routingKey) {
				continue
			}

			_, err := wf.Write(msg.Body)
			if err != nil {
				fmt.Println(err)
			}
			wf.Flush()

			msg.Ack(false)
		}
	}
}

func getMsgs(routingKey string) (w *worker.Worker, msgs <-chan amqp.Delivery, err error) {
	w, err = worker.NewWorker(os.Getenv("AMQP_URL"))
	if err != nil {
		return
	}

	err = w.EnsureExchanges("logs")
	if err != nil {
		return
	}

	q, err := w.QueueDeclare(worker.QueueOpts{
		Durable:   true,
		Exclusive: true,
	})
	if err != nil {
		return
	}

	err = w.Ch.QueueBind(q.Name, routingKey, "logs", false, nil)
	if err != nil {
		return
	}

	msgs, err = w.Consume(worker.ConsumeOpts{
		Queue: q.Name,
	})

	return
}
