package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/projecteru2/stats/config"
	"github.com/projecteru2/stats/router"
	"github.com/projecteru2/stats/versioninfo"
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
	app.Usage = "Run stats"
	app.Version = versioninfo.VERSION
	app.Action = func(c *cli.Context) error {
		run()
		return nil
	}

	app.Run(os.Args)
}
