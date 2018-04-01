package api

import (
	"log"

	"github.com/mattn/go-colorable"

	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/config"
)

/* gin app */
var app *gin.Engine

// Start REST API server
func Start(cfg config.ApiConfig) {

	if !cfg.Enabled {
		log.Printf("[INFO] API disabled")
		return
	}

	log.Printf("[INFO] Starting up API")
	gin.DefaultWriter = colorable.NewColorableStdout()
	//gin.DefaultWriter = ioutil.Discard
	gin.SetMode(gin.ReleaseMode)
	app = gin.Default()
	app.GET("/", serverInit)

	var err error
	err = app.Run(cfg.Bind)

	if err != nil {
		log.Fatal(err)
	}

}
