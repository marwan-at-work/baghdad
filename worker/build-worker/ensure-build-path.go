package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/marwan-at-work/baghdad"
)

func ensureBuildPath() {
	pr := strconv.Itoa(baghdad.PullRequestEvent)
	err := os.MkdirAll(filepath.Join(WorkPath, pr), 0666)
	if err != nil {
		panic(fmt.Errorf("could not make pr dir: %v", err))
	}

	push := strconv.Itoa(baghdad.PushEvent)
	err = os.MkdirAll(filepath.Join(WorkPath, push), 0666)
	if err != nil {
		panic(fmt.Errorf("could not make push dir: %v", err))
	}

	err = os.MkdirAll(filepath.Join(WorkPath, "tars"), 0666)
	if err != nil {
		panic(fmt.Errorf("could not make push dir: %v", err))
	}

	err = os.MkdirAll(filepath.Join(WorkPath, "artifacts"), 0666)
	if err != nil {
		panic(fmt.Errorf("could not make push dir: %v", err))
	}
}
