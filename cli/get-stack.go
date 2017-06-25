package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/marwan-at-work/baghdad/utils"
	cli "gopkg.in/urfave/cli.v2"
)

func getStack(c *cli.Context) (err error) {
	tag := c.String("tag")
	if tag == "" {
		tag = "latest"
	}

	env := c.String("env")
	if env == "" {
		return errors.New("env flag is required")
	}

	domain := c.String("host")
	if domain == "" {
		return errors.New("host flag is required")
	}

	stackFile, err := ioutil.ReadFile("./stack-compose.yml")
	if err != nil {
		return
	}

	b, err := getBaghdad()
	if err != nil {
		return
	}

	t, err := newTag(tag)
	if err != nil {
		return
	}

	bts, err := utils.TagStackServices(stackFile, b, tag, t.BranchName, env, domain)
	fmt.Println(string(bts))
	return
}

func newTag(tag string) (Tag, error) {
	t := Tag{}
	r, _ := regexp.Compile(`([a-z0-9]+)-(\d+.\d+.\d+)-build\.(\d+)`)
	if !r.MatchString(tag) {
		return t, errors.New("invalid tag")
	}

	subs := r.FindStringSubmatch(tag)
	t.BranchName = subs[1]
	t.Version = subs[2]
	t.BuildNumber, _ = strconv.Atoi(subs[3])

	return t, nil
}

// Tag represents a baghdad tag: <branch>-<version>-build.<build-number>
// Potentailly moveable to be used by the core system.
type Tag struct {
	BranchName  string
	Version     string
	BuildNumber int
}
