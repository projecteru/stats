package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/etcd/client"
	"gitlab.ricebook.net/platform/eru-stats/config"
	"gitlab.ricebook.net/platform/eru-stats/router"
	"gitlab.ricebook.net/platform/eru-stats/versioninfo"
)

var (
	logLevel string
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

func initConfig() error {
	config.C = &config.Config{}

	config.C.Bind = ":6006"
	if os.Getenv("BIND") != "" {
		config.C.Bind = os.Getenv("BIND")
	}
	config.C.Etcd.AgentPrefix = "/agent2"
	if os.Getenv("AgentPrefix") != "" {
		config.C.Etcd.AgentPrefix = os.Getenv("AgentPrefix")
	}
	config.C.Etcd.CorePrefix = "/eru-core"
	if os.Getenv("CorePrefix") != "" {
		config.C.Etcd.CorePrefix = os.Getenv("CorePrefix")
	}
	nodeIP := os.Getenv("ERU_NODE_IP")
	etcdEnpoints := fmt.Sprintf("http://%s:2379", nodeIP)
	if os.Getenv("EtcdEndpoints") != "" {
		etcdEnpoints = os.Getenv("EtcdEndpoints")
	}
	etcdEnpoints = strings.Replace(etcdEnpoints, "\\", "", -1)
	config.C.Etcd.EtcdEndpoints = strings.Split(etcdEnpoints, ",")

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

	if err := initConfig(); err != nil {
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
