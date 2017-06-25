package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	cli "gopkg.in/urfave/cli.v2"
)

func register(c *cli.Context) error {
	s := c.String("server")
	if s == "" {
		return errors.New("server flag required")
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		return err
	}

	file := filepath.Join(homeDir, ".baghdad")
	err = ioutil.WriteFile(file, []byte(s), 0666)
	return err
}
