package router

import (
	"github.com/deckarep/golang-set"
	"github.com/gin-gonic/gin"
	"gitlab.ricebook.net/platform/stats/config"
	"gitlab.ricebook.net/platform/stats/handler"
)

// Run eruapp server
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
	}

	_, agentContainers, err := handler.AgentAllNodesAndContainers()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
	}

	coreContainerSet := mapset.NewSet()
	for _, c := range coreContainers {
		coreContainerSet.Add(c)
	}
	agentContainerSet := mapset.NewSet()
	for _, c := range agentContainers {
		agentContainerSet.Add(c)
	}

	// core - agent
	agentLess := coreContainerSet.Difference(agentContainerSet)

	// agent - core
	agentMore := agentContainerSet.Difference(coreContainerSet)

	c.JSON(200, gin.H{
		"agentLess": agentLess.ToSlice(),
		"agentMore": agentMore.ToSlice(),
	})

}

func podstatus(c *gin.Context) {
	_, podNodes, err := handler.CoreNodes()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
	}

	c.JSON(200, podNodes)
}

func appstatus(c *gin.Context) {
	appStats, err := handler.AppStats()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
	}

	c.JSON(200, appStats)
}
