// Package dfparser naivly reads through a dockerfile line by line, and maps it
// it a Dockerfile struct.
package dfparser

import (
	"bufio"
	"io"
	"strings"
)

// Dockerfile represents the parsed Dockerfile struct
type Dockerfile struct {
	Add         []string
	Arg         []string
	Cmd         []string
	Copy        []string
	Entrypoint  []string
	Env         []string
	Expose      []string
	From        []string
	Healthcheck []string
	Label       []string
	Maintainer  []string
	Onbuild     []string
	Run         []string
	Shell       []string
	StopSignal  []string
	User        []string
	Volume      []string
	Workdir     []string
}

// Parse takes a reader originated from a dockerfile and returns a Dockerfile struct.
func Parse(f io.Reader) (d Dockerfile, err error) {
	scanner := bufio.NewScanner(f)
	d = Dockerfile{}

	for scanner.Scan() {
		cmd, args := parseLine(scanner.Text())
		switch strings.ToUpper(cmd) {
		case "ADD":
			d.Add = append(d.Add, args)
		case "ARG":
			d.Arg = append(d.Arg, args)
		case "CMD":
			d.Cmd = append(d.Cmd, args)
		case "COPY":
			d.Copy = append(d.Copy, args)
		case "ENTRYPOINT":
			d.Entrypoint = append(d.Entrypoint, args)
		case "ENV":
			d.Env = append(d.Env, args)
		case "EXPOSE":
			d.Expose = append(d.Expose, args)
		case "FROM":
			d.From = append(d.From, args)
		case "HEALTHCHECK":
			d.Healthcheck = append(d.Healthcheck, args)
		case "LABEL":
			d.Label = append(d.Label, args)
		case "MAINTAINER":
			d.Maintainer = append(d.Maintainer, args)
		case "ONBUILD":
			d.Onbuild = append(d.Onbuild, args)
		case "RUN":
			d.Run = append(d.Run, args)
		case "SHELL":
			d.Shell = append(d.Shell, args)
		case "STOPSIGNAL":
			d.StopSignal = append(d.StopSignal, args)
		case "USER":
			d.User = append(d.User, args)
		case "VOLUME":
			d.Volume = append(d.Volume, args)
		case "WORKDIR":
			d.Workdir = append(d.Workdir, args)
		}
	}

	err = scanner.Err()

	return
}

func parseLine(s string) (cmd, args string) {
	ss := strings.Split(s, " ")
	cmd = ss[0]
	args = strings.Join(ss[1:], " ")

	return
}
