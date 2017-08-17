package router

import (
	"github.com/gin-gonic/gin"
	"github.com/projecteru2/stats/apiproxy"
	"github.com/projecteru2/stats/config"
	"github.com/projecteru2/stats/handler"
)

// Run stats server
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
	nodeStats, err := apiproxy.PodsMemCap()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"Agent": gin.H{
			"Containers": agentContainerNum,
			"Nodes":      agentNodeNum,
		},
		"Core": gin.H{
			"Containers": coreContainerNum,
			"Nodes":      coreNodeNum,
		},
		"NodesMemcap": nodeStats,
	})
}

func diff(c *gin.Context) {
	agentLessContainers, agentMoreContainers, err := handler.DiffContainers()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}

	agentNodeLess, agentNodeMore, err := handler.DiffNodes()
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"container": gin.H{
			"agentLess": agentLessContainers,
			"agentMore": agentMoreContainers,
		},
		"nodes": gin.H{
			"agentLess": agentNodeLess,
			"agentMore": agentNodeMore,
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
