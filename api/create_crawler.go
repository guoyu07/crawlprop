package api

import (
	"log"

	"github.com/millken/crawlprop/core"

	"github.com/gin-gonic/gin"
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
	log.Printf("%+v", c.URL)
	crawler := core.NewCrawler(c.URL.Get("name"), c.URL.Get("target"))
	crawler.Start()
	data = gin.H{"records": "23"}
	return
}
