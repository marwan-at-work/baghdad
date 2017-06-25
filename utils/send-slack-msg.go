package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// SendSlackMessage attempts to send a given msg to the given url.
func SendSlackMessage(url, msg string) {
	body := map[string]string{"text": msg}
	var bts bytes.Buffer
	json.NewEncoder(&bts).Encode(body)

	http.Post(url, "application/json", &bts)
}
