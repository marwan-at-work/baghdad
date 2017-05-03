package main

import (
	"fmt"
	"os"

	"github.com/marwan-at-work/dfparser"
)

func getDockerfileWorkdir(path string) (workdir string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	d, err := dfparser.Parse(f)
	if err != nil {
		return
	}

	if len(d.Workdir) == 0 {
		return "/", nil
	}

	workdir = d.Workdir[len(d.Workdir)-1]
	if workdir == "" {
		err = fmt.Errorf("dockefile, %v, returned an empty workdir", path)
	}

	return
}
