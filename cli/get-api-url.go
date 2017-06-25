package main

import (
	"io/ioutil"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func getAPIURL() (url string, err error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return
	}

	apiURL, err := ioutil.ReadFile(filepath.Join(homeDir, ".baghdad"))
	if err != nil {
		return
	}

	return string(apiURL), err
}
