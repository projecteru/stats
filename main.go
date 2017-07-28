package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"gitlab.ricebook.net/platform/eru-stats/config"
	"gitlab.ricebook.net/platform/eru-stats/router"
	"gitlab.ricebook.net/platform/eru-stats/versioninfo"
)

func run() {

	if err := config.InitConf(); err != nil {
		log.Fatal(err)
	}

	if err := router.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}

	app := cli.NewApp()
	app.Name = versioninfo.NAME
	app.Usage = "Run eru-stats"
	app.Version = versioninfo.VERSION
	app.Action = func(c *cli.Context) error {
		run()
		return nil
	}

	app.Run(os.Args)
}
