package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	ctypes "github.com/projecteru2/core/types"
	cutils "github.com/projecteru2/core/utils"
	"github.com/projecteru2/stats/config"
	"github.com/projecteru2/stats/types"
)

func CorePods() (pods []string, err error) {
	key := fmt.Sprintf("%s/pod", config.C.Etcd.CorePrefix)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return pods, err
	}
	for _, node := range resp.Node.Nodes {
		t := strings.Split(node.Key, "/")
		pods = append(pods, t[len(t)-1])
	}
	return pods, nil
}

func nodeGetInfo(nodeKey string) (ctypes.Node, error) {
	nodeInfo := ctypes.Node{}
	key := fmt.Sprintf("%s/info", nodeKey)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return nodeInfo, err
	}
	if err := json.Unmarshal([]byte(resp.Node.Value), &nodeInfo); err != nil {
		return nodeInfo, err
	}
	return nodeInfo, nil
}

func podGetNodes(podname string) (nodes []Node, err error) {
	key := fmt.Sprintf("%s/pod/%s/node", config.C.Etcd.CorePrefix, podname)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return nodes, err
	}
	respNodes := resp.Node.Nodes

	var wg sync.WaitGroup
	wg.Add(len(respNodes))
	nodeInfoChan := make(chan Node)

	go func() {
		wg.Add(1)
		defer wg.Done()
		remaining := len(respNodes)
		for n := range nodeInfoChan {
			if n.HostName != "" {
				nodes = append(nodes, n)
			}
			if remaining--; remaining == 0 {
				close(nodeInfoChan)
			}
		}
	}()

	for _, node := range respNodes {
		go func(nodeKey string) {
			defer wg.Done()
			node, err := nodeGetInfo(nodeKey)
			if err != nil {
				log.Errorf("Get node info error: %s", err)
				nodeInfoChan <- Node{}
				return
			}
			n := Node{
				HostName: node.Name,
				PodName:  podname,
				Mem:      fmt.Sprintf("%d MB | %d GB", (node.MemCap / 1024 / 1024), (node.MemCap / 1024 / 1024 / 1024)),
			}
			nodeInfoChan <- n
		}(node.Key)
	}

	wg.Wait()

	return nodes, nil
}

func CoreNodes() (nodes []Node, podNodes map[string][]Node, err error) {
	pods, err := CorePods()
	if err != nil {
		log.Errorf("Get core all nodes error: %s", err)
		return nodes, podNodes, err
	}

	var wg sync.WaitGroup
	wg.Add(len(pods))
	nodesChan := make(chan []Node)

	go func() {
		wg.Add(1)
		defer wg.Done()
		remaining := len(pods)
		for c := range nodesChan {
			nodes = append(nodes, c...)
			if remaining--; remaining == 0 {
				close(nodesChan)
			}
		}
	}()

	podNodes = map[string][]Node{}
	for _, pod := range pods {
		go func(podname string) {
			defer wg.Done()
			nodes, err := podGetNodes(podname)
			if err != nil {
				log.Errorf("Pod get its nodes error: %s", err)
				return
			}
			podNodes[podname] = nodes
			nodesChan <- nodes
		}(pod)
	}

	wg.Wait()

	return nodes, podNodes, nil
}

func CoreContainers() (containers []string, err error) {
	key := fmt.Sprintf("%s/container", config.C.Etcd.CorePrefix)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return containers, err
	}

	for _, node := range resp.Node.Nodes {
		t := strings.Split(node.Key, "/")
		containers = append(containers, t[len(t)-1])
	}

	return containers, nil
}

func CoreStats() (int, int) {
	coreContainers, err := CoreContainers()
	if err != nil {
		return 0, 0
	}
	coreNodes, _, err := CoreNodes()
	if err != nil {
		return 0, 0
	}
	return len(coreNodes), len(coreContainers)
}

func coreGetContainerInfo(containerID string) (container types.Container, err error) {
	key := fmt.Sprintf("%s/container/%s", config.C.Etcd.CorePrefix, containerID)
	e := config.C.Etcd.Api
	resp, err := e.Get(context.Background(), key, &client.GetOptions{})
	if err != nil {
		return container, err
	}
	c := ctypes.Container{}
	if err := json.Unmarshal([]byte(resp.Node.Value), &c); err != nil {
		return container, err
	}

	appname, entrypoint, _, _ := cutils.ParseContainerName(c.Name)
	container = types.Container{
		ID:         containerID,
		AppName:    appname,
		Entrypoint: entrypoint,
		Memory:     c.Memory,
		CPU:        c.CPU,
		Pod:        c.Podname,
		Node:       c.Nodename,
	}

	return container, nil
}

func AppStats() (appStats map[string]*types.App, err error) {
	allContainers, err := CoreContainers()
	if err != nil {
		return appStats, err
	}

	// get all containers' info
	allContainerInfo, err := ContainersInfo(allContainers)
	if err != nil {
		return appStats, err
	}
	// 整理container信息
	appStats = map[string]*types.App{}

	for _, c := range allContainerInfo {
		if _, ok := appStats[c.AppName]; !ok {
			appStats[c.AppName] = &types.App{
				Entrypoints: map[string]*types.Entrypoint{},
			}
		}
		appStats[c.AppName].Count++
		appStats[c.AppName].MemTotal += c.Memory
		appStats[c.AppName].CPUTotal += c.CPU.Total()

		if _, ok := appStats[c.AppName].Entrypoints[c.Entrypoint]; !ok {
			appStats[c.AppName].Entrypoints[c.Entrypoint] = &types.Entrypoint{}
		}
		appStats[c.AppName].Entrypoints[c.Entrypoint].Count++
		appStats[c.AppName].Entrypoints[c.Entrypoint].Mem += c.Memory
	}

	for _, app := range appStats {
		app.Mem = fmt.Sprintf("%d MB", (app.MemTotal / 1024 / 1024))
	}

	return appStats, nil
}

func ContainersInfo(IDs []string) ([]types.Container, error) {
	var wg sync.WaitGroup
	wg.Add(len(IDs))
	containerChan := make(chan types.Container)
	allContainerInfo := []types.Container{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		remaining := len(IDs)
		for c := range containerChan {
			if c.ID != "" {
				allContainerInfo = append(allContainerInfo, c)
			}
			if remaining--; remaining == 0 {
				close(containerChan)
			}
		}
	}()
	for _, id := range IDs {
		go func(id string) {
			defer wg.Done()
			container, err := coreGetContainerInfo(id)
			if err != nil {
				log.Errorf("Get container info error: %s", err)
				containerChan <- types.Container{}
				return
			}
			containerChan <- container
		}(id)
	}
	wg.Wait()
	return allContainerInfo, nil
}
