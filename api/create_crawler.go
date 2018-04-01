package api

import (
	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/stored"
)

func init() {
	Register("createCrawler", NewCreateCrawler)
}

type CreateCrawler struct {
	*ActionParam
}

func NewCreateCrawler(v *ActionParam) (Action, error) {
	return &CreateCrawler{
		v,
	}, nil
}

func (c *CreateCrawler) Response() (data gin.H, err error) {
	uuid := UUID()
	stored.Create(uuid, c.URL.Get("name"), c.URL.Get("target"))
	data = gin.H{"task_id": uuid}
	return
}
