package config

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
)

type etcdConf struct {
	EtcdEndpoints []string
	AgentPrefix   string
	CorePrefix    string
	Api           client.KeysAPI
}

type citadelConf struct {
	Host string
	Auth string
}

type config struct {
	Bind     string
	Etcd     etcdConf
	LogLevel string
	Citadel  citadelConf
}

// global config
var C *config

func InitConf() error {
	C = &config{}

	C.Bind = ":6006"
	C.Etcd.AgentPrefix = "/eru-agent"
	C.Etcd.CorePrefix = "/eru-core"
	C.LogLevel = "info"
	C.Citadel.Host = "http://citadel.ricebook.net"
	C.Citadel.Auth = "hello"

	// bind port
	if os.Getenv("PORT_BIND") != "" {
		C.Bind = os.Getenv("PORT_BIND")
	}

	// etcd prefix
	if os.Getenv("AGENT_PREFIX") != "" {
		C.Etcd.AgentPrefix = os.Getenv("AGENT_PREFIX")
	}
	if os.Getenv("CORE_PREFIX") != "" {
		C.Etcd.CorePrefix = os.Getenv("CORE_PREFIX")
	}

	// set citadel config
	if os.Getenv("CITADEL_URL") != "" {
		host := os.Getenv("CITADEL_URL")
		C.Citadel.Host = strings.TrimRight(host, "/")
	}
	if os.Getenv("CITADEL_AUTH_TOKEN") != "" {
		C.Citadel.Auth = os.Getenv("CITADEL_AUTH_TOKEN")
	}

	// set etcd endpoints
	nodeIP := os.Getenv("ERU_NODE_IP")
	etcdEnpoints := fmt.Sprintf("http://%s:2379", nodeIP)
	if os.Getenv("ETCD_ENDPOINTS") != "" {
		etcdEnpoints = os.Getenv("ETCD_ENDPOINTS")
	}
	etcdEnpoints = strings.Replace(etcdEnpoints, "\\", "", -1)
	C.Etcd.EtcdEndpoints = strings.Split(etcdEnpoints, ",")

	// init etcdclient
	etcdClient, err := client.New(client.Config{Endpoints: C.Etcd.EtcdEndpoints})
	if err != nil {
		return err
	}
	C.Etcd.Api = client.NewKeysAPI(etcdClient)

	// set log
	if os.Getenv("LOG_LEVEL") != "" {
		C.LogLevel = os.Getenv("LOG_LEVEL")
	}
	if err := setupLog(C.LogLevel); err != nil {
		return err
	}

	return nil
}

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
