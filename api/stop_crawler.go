package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/stored"
)

func init() {
	Register("stopCrawler", NewStopCrawler)
}

type StopCrawler struct {
	*ActionParam
}

func NewStopCrawler(v *ActionParam) (Action, error) {
	return &StopCrawler{
		v,
	}, nil
}

func (c *StopCrawler) Response() (data gin.H, err error) {
	taskID := c.URL.Get("task_id")
	err = stored.Delete(taskID)
	if err != nil {
		log.Printf("[INFO] failed to stop crawler : %s", err)
		data = gin.H{"stopped": false}
		return data, nil
	}
	data = gin.H{"stopped": true}
	return
}
