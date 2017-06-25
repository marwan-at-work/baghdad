package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type logGetter interface {
	GetLogs(id string) ([]string, error)
}

// RedisLog implements logGetter
type RedisLog struct{}

// GetLogs implements the logGetter for redis in production
func (r *RedisLog) GetLogs(id string) ([]string, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	logs, err := c.LRange(id, 0, -1).Result()
	if err != nil {
		return []string{}, err
	}

	results := []string{}

	for i := len(logs) - 1; i >= 0; i-- {
		results = append(results, logs[i])
	}

	return results, nil
}

// TemplateData for index.html
type TemplateData struct {
	Title string
	Logs  []string
}

// GetBuildLogs returns persisted logs based on the logID
func GetBuildLogs(l logGetter, t *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		project := mux.Vars(r)["project"]
		logID := mux.Vars(r)["logID"]

		logs, err := l.GetLogs(logID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err)
		}

		d := TemplateData{
			Title: project,
			Logs:  logs,
		}

		err = t.Lookup("index.html").Execute(w, d)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
		}
	}
}
