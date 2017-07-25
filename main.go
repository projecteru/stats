package main

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
	"gitlab.ricebook.net/platform/eru-stats/config"
	"gitlab.ricebook.net/platform/eru-stats/router"
	"gitlab.ricebook.net/platform/eru-stats/versioninfo"
	"gopkg.in/yaml.v2"
)

var (
	configPath string
	logLevel   string
)

func setupLog(l string) error {
	level, err := log.ParseLevel(l)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	return nil
}

func initConfig(configPath string) error {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	config.C = &config.Config{}
	if err := yaml.Unmarshal(bytes, config.C); err != nil {
		return err
	}
	etcdClient, err := client.New(client.Config{Endpoints: config.C.Etcd.EtcdEndpoints})
	if err != nil {
		return err
	}
	config.C.Etcd.Api = client.NewKeysAPI(etcdClient)
	return nil
}

func serve() {
	if err := setupLog(logLevel); err != nil {
		log.Fatal(err)
	}

	if configPath == "" {
		log.Fatalf("Config path must be set")
	}

	if err := initConfig(configPath); err != nil {
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
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Value:       "/etc/eru/eru-stats.yaml",
			Usage:       "config file path for eru-stats, in yaml",
			Destination: &configPath,
			EnvVar:      "ERU_CONFIG_PATH",
		},
		cli.StringFlag{
			Name:        "log-level",
			Value:       "INFO",
			Usage:       "set log level",
			Destination: &logLevel,
			EnvVar:      "ERU_LOG_LEVEL",
		},
	}
	app.Action = func(c *cli.Context) error {
		serve()
		return nil
	}

	app.Run(os.Args)
}
