package main

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/marwan-at-work/baghdad"

	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:  "baghdad",
		Usage: "automate your swarm",
		Commands: []*cli.Command{
			{
				Name:  "secret",
				Usage: "manage project secrets",
				Subcommands: []*cli.Command{
					{
						Name:   "create",
						Usage:  "deploy a secret to the swarm i.e. baghdad secret create <secret-name> <path-to-file>",
						Action: secretCreate,
					},
				},
			},
			{
				Name:  "register",
				Usage: "register your credentials to be used for subsequent commands",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "server",
						Usage: "Baghdad API's URL",
					},
				},
				Action: register,
			},
			{
				Name:   "logs",
				Usage:  "listen to your application build/app logs",
				Action: getLogs,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "service",
						Usage:   "get logs for a specific service",
						Aliases: []string{"s"},
					},
					&cli.StringFlag{
						Name:    "env",
						Usage:   "specify environment to which the service belongs",
						Aliases: []string{"e"},
					},
				},
			},
			{
				Name:  "generate",
				Usage: "generate configs to stdout",
				Subcommands: []*cli.Command{
					{
						Name:   "stack",
						Action: getStack,
						Usage:  "generate deployable stack compose file",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "tag",
								Aliases: []string{"t"},
								Usage:   "specify the tag that applies to each baghdad.toml service",
							},
							&cli.StringFlag{
								Name:    "env",
								Aliases: []string{"e"},
								Usage:   "specify the environment pipeline",
							},
							&cli.StringFlag{
								Name:  "host",
								Usage: "specify the host domain name",
							},
						},
					},
				},
			},
			{
				Name:   "deploy",
				Usage:  "deploy a project to your cluster",
				Action: deploy,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "tag",
						Usage:   "the build tag of the repository - required",
						Aliases: []string{"T"},
					},
					&cli.StringFlag{
						Name:    "env",
						Usage:   "the environment you want to deploy to",
						Aliases: []string{"E"},
					},
					&cli.StringFlag{
						Name:    "branch",
						Usage:   "the branch you want to deploy from",
						Aliases: []string{"B"},
					},
					&cli.StringFlag{
						Name:    "owner",
						Usage:   "the github username",
						Aliases: []string{"O"},
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func getBaghdad() (b baghdad.Baghdad, err error) {
	p, err := ioutil.ReadFile("./baghdad.toml")
	if err != nil {
		return
	}

	b = baghdad.Baghdad{}
	err = toml.Unmarshal(p, &b)
	return
}
