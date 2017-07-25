package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	atypes "gitlab.ricebook.net/platform/agent/types"
	ctypes "gitlab.ricebook.net/platform/core/types"
	"gitlab.ricebook.net/platform/stats/config"
)

// Node running agent
type Node struct {
	HostName string
	PodName  string
	Mem      string
	ctypes.CPUMap
}

func (n *Node) allContainers() ([]string, error) {
	key := fmt.Sprintf("%s/%s/containers", config.C.Etcd.AgentPrefix, n.HostName)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return nil, err
	}

	var containers []string
	for _, node := range resp.Node.Nodes {
		t := strings.Split(node.Key, "/")
		containers = append(containers, t[len(t)-1])
	}

	return containers, nil
}

func (n *Node) getContainer(cid string) (atypes.Container, error) {
	key := fmt.Sprintf("%s/%s/container/%s", config.C.Etcd.AgentPrefix, n.HostName, cid)
	container := atypes.Container{}
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return atypes.Container{}, err
	}
	if err := json.Unmarshal([]byte(resp.Node.Value), &container); err != nil {
		return atypes.Container{}, err
	}
	return container, nil
}

// AgentAllNodesAndContainers get all nodes and all containers in etcd(agent keys)
func AgentAllNodesAndContainers() (nodes, containers []string, err error) {
	key := fmt.Sprintf("%s", config.C.Etcd.AgentPrefix)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	for _, n := range resp.Node.Nodes {
		t := strings.Split(n.Key, "/")
		nodes = append(nodes, t[len(t)-1])
	}

	var wg sync.WaitGroup
	containerChan := make(chan []string)

	go func() {
		wg.Add(1)
		defer wg.Done()
		remaining := len(nodes)
		for c := range containerChan {
			containers = append(containers, c...)
			if remaining--; remaining == 0 {
				close(containerChan)
			}
		}
	}()

	wg.Add(len(nodes))
	for _, n := range nodes {
		go func(hostname string) {
			defer wg.Done()
			node := Node{HostName: hostname}
			nodeContainers, err := node.allContainers()
			if err != nil {
				log.Errorf("CountContainers error: %s", err)
				return
			}
			containerChan <- nodeContainers
		}(n)
	}
	wg.Wait()

	return nodes, containers, nil
}

// AgentStats returns how many nodes and containers in agent's etcd
func AgentStats() (int, int) {
	nodes, containers, err := AgentAllNodesAndContainers()
	if err != nil {
		log.Errorf("Get agent status error: %s", err)
	}

	return len(nodes), len(containers)
}
