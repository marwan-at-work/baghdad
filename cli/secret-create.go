package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	cli "gopkg.in/urfave/cli.v2"
)

func secretCreate(c *cli.Context) (err error) {
	b, err := getBaghdad()
	if err != nil {
		return
	}

	args := c.Args()
	if args.Len() != 2 {
		return errors.New("must have two arguments: name of secret and path to secret file")
	}

	secretName := args.Get(0)
	secretPath := args.Get(1)

	bts, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return fmt.Errorf("could not read file path %v: %v", secretPath, err)
	}

	apiURL, err := getAPIURL()
	if err != nil {
		return
	}

	body := map[string]string{
		"name": secretName,
		"body": string(bts),
	}

	bodyBytes, _ := json.Marshal(body)
	u, err := url.Parse(apiURL)
	if err != nil {
		return
	}

	ref, _ := url.Parse(path.Join("projects", b.Project, "secrets"))
	fullURL := u.ResolveReference(ref).String()

	resp, err := http.Post(fullURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}

	fmt.Println(resp.Status)
	return
}
