package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"net/http"

	cli "gopkg.in/urfave/cli.v2"
)

func deploy(c *cli.Context) (err error) {
	t := c.String("tag")
	if t == "" {
		return errors.New("missing tag")
	}

	e := c.String("env")
	if e == "" {
		return errors.New("missing env")
	}

	b := c.String("branch")
	if b == "" {
		return errors.New("missing branch")
	}

	// hard coded for now.
	owner := "marwan-at-work"
	apiURL, err := getAPIURL()
	if err != nil {
		return err
	}

	url := fmt.Sprintf(
		"%v/projects/%v/%v/deploy",
		apiURL,
		owner,
		getDirName(),
	)

	dpb := deployPostBody{
		Branch: b,
		Env:    e,
		Tag:    t,
	}
	var bts bytes.Buffer
	err = json.NewEncoder(&bts).Encode(dpb)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", &bts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("deploy job sent")
	} else {
		fmt.Println("deploy job responded with", resp.StatusCode)
	}

	return nil
}

type deployPostBody struct {
	Branch string `json:"branch"`
	Env    string `json:"environment"`
	Tag    string `json:"tag"`
}
