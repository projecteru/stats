package handler

import (
	"fmt"
	"strings"

	"github.com/deckarep/golang-set"
)

func DiffContainers() ([]string, []string, error) {

	coreContainers, err := CoreContainers()
	if err != nil {
		return nil, nil, err
	}

	_, agentContainers, err := AgentAllNodesAndContainers()
	if err != nil {
		return nil, nil, err
	}

	// containers
	coreContainerSet := mapset.NewSet()
	for _, c := range coreContainers {
		coreContainerSet.Add(c)
	}
	agentContainerSet := mapset.NewSet()
	for id := range agentContainers {
		agentContainerSet.Add(id)
	}

	// core - agent
	agentContainerLess := coreContainerSet.Difference(agentContainerSet)
	// agent - core
	agentContainerMore := agentContainerSet.Difference(coreContainerSet)

	// findout container nodes
	agentMoreContainers := []string{}
	for _, i := range agentContainerMore.ToSlice() {
		id := i.(string)
		agentMoreContainers = append(agentMoreContainers, fmt.Sprintf("%s__on_node__%s", id, agentContainers[id]))
	}

	agentLessContainers := []string{}
	s := []string{}
	for _, i := range agentContainerLess.ToSlice() {
		s = append(s, i.(string))
	}
	agentLessContainersInfo, err := ContainersInfo(s)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range agentLessContainersInfo {
		agentLessContainers = append(agentLessContainers, fmt.Sprintf("%s__on_node__%s", c.ID, c.Node))
	}

	return agentLessContainers, agentMoreContainers, nil
}

func DiffNodes() ([]interface{}, []interface{}, error) {
	coreNodes, _, err := CoreNodes()
	if err != nil {
		return nil, nil, err
	}

	agentNodes, _, err := AgentAllNodesAndContainers()
	if err != nil {
		return nil, nil, err
	}

	// nodes
	coreNodeSet := mapset.NewSet()
	for _, n := range coreNodes {
		coreNodeSet.Add(n.HostName)
	}

	agentNodeSet := mapset.NewSet()
	for _, n := range agentNodes {
		nodeShotName := strings.Split(n, ".")[0]
		agentNodeSet.Add(nodeShotName)
	}

	// core - agent
	agentNodeLess := coreNodeSet.Difference(agentNodeSet)
	// agent - core
	agentNodeMore := agentNodeSet.Difference(coreNodeSet)

	return agentNodeLess.ToSlice(), agentNodeMore.ToSlice(), nil
}
