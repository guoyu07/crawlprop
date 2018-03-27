package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/core"
)

func attachServers(app *gin.RouterGroup) {

	app.GET("/", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, nil)
	})

	app.GET("/add", func(c *gin.Context) {
		name := c.DefaultQuery("name", "Guest")
		core.Scheduler().Push(name)
		c.JSON(200, gin.H{
			"status":  true,
			"message": "succefully to add to task",
			"url":     name,
		})
	})

}
