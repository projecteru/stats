package router

import (
	"fmt"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/gin-gonic/gin"
	"gitlab.ricebook.net/platform/eru-stats/config"
	"gitlab.ricebook.net/platform/eru-stats/handler"
)

// Run eru-stats server
func Run() error {
	r := gin.Default()

	// get all routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"apis": r.Routes(),
		})
	})
	r.GET("/ping", ping)

	r.GET("/stats", statistics)

	r.GET("/diff", diff)

	r.GET("/pods", podstatus)

	r.GET("/apps", appstatus)

	return r.Run(config.C.Bind)
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"ping": "pong",
	})
}

func statistics(c *gin.Context) {
	agentNodeNum, agentContainerNum := handler.AgentStats()
	coreNodeNum, coreContainerNum := handler.CoreStats()
	c.JSON(200, gin.H{
		"Agent": gin.H{
			"Containers": agentContainerNum,
			"Nodes":      agentNodeNum,
		},
		"Core": gin.H{
			"Containers": coreContainerNum,
			"Nodes":      coreNodeNum,
		},
	})
}

func diff(c *gin.Context) {
	coreContainers, err := handler.CoreContainers()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}
	coreNodes, _, err := handler.CoreNodes()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}

	agentNodes, agentContainers, err := handler.AgentAllNodesAndContainers()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
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
	agentContainerLess := coreContainerSet.Difference(agentContainerSet)
	// agent - core
	agentContainerMore := agentContainerSet.Difference(coreContainerSet)

	// core - agent
	agentNodeLess := coreNodeSet.Difference(agentNodeSet)
	// agent - core
	agentNodeMore := agentNodeSet.Difference(coreNodeSet)

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
	agentLessContainersInfo, err := handler.ContainersInfo(s)
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}
	for _, c := range agentLessContainersInfo {
		agentLessContainers = append(agentLessContainers, fmt.Sprintf("%s__on_node__%s", c.ID, c.Node))
	}

	c.JSON(200, gin.H{
		"container": gin.H{
			"agentLess": agentLessContainers,
			"agentMore": agentMoreContainers,
		},
		"nodes": gin.H{
			"agentLess": agentNodeLess.ToSlice(),
			"agentMore": agentNodeMore.ToSlice(),
		},
	})

}

func podstatus(c *gin.Context) {
	_, podNodes, err := handler.CoreNodes()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(200, podNodes)
}

func appstatus(c *gin.Context) {
	appStats, err := handler.AppStats()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(200, appStats)
}
