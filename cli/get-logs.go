package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/marwan-at-work/baghdad"
	cli "gopkg.in/urfave/cli.v2"
)

func getLogs(c *cli.Context) (err error) {
	var p string
	var b baghdad.Baghdad
	b, err = getBaghdad()
	if err != nil {
		// if baghdad doesn't exist, then it's a simple project with just a Dockerfile. Check for Dockerfile?
		p = getDirName()
	} else {
		p = b.Project
	}

	if p == "" {
		fmt.Println("missing baghdad.toml or incorrect wd")
		return
	}

	apiURL, err := getAPIURL()
	if err != nil {
		return fmt.Errorf("could not get .baghdad: %v", err)
	}

	url := fmt.Sprintf("%s/projects/%v/logs", apiURL, p) // env needs to be part of the login.
	service := c.String("service")
	if service != "" {
		env := c.String("env")
		if env == "" {
			return errors.New("you need to specify the env flag")
		}

		url = fmt.Sprintf(
			"%v/projects/%v/services/%v/logs?env=%v",
			apiURL,
			p,
			service,
			env,
		)
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)
	return
}

func getDirName() string {
	path, _ := os.Getwd()
	dirs := strings.Split(path, "/")
	return dirs[len(dirs)-1]
}
