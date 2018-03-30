package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/config"
)

/* gin app */
var app *gin.Engine

/**
 * Initialize module
 */
func init() {
	gin.SetMode(gin.ReleaseMode)
}

/**
 * Starts REST API server
 */
func Start(cfg config.ApiConfig) {

	if !cfg.Enabled {
		log.Printf("[INFO] API disabled")
		return
	}

	log.Printf("[INFO] Starting up API")

	app = gin.Default()
	app.GET("/", serverInit)

	var err error
	err = app.Run(cfg.Bind)

	if err != nil {
		log.Fatal(err)
	}

}
