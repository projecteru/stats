package config

import (
	"github.com/coreos/etcd/client"
)

type EtcdConf struct {
	EtcdEndpoints []string `yaml:"endpoints"`
	AgentPrefix   string   `yaml:"agentprefix"`
	CorePrefix    string   `yaml:"coreprefix"`
	Api           client.KeysAPI
}

type Config struct {
	Bind string   `yaml:"bind"`
	Etcd EtcdConf `yaml:"etcd"`
}

// global config
var C *Config
