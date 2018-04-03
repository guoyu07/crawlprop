package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/millken/crawlprop/core"
	"github.com/millken/crawlprop/stored"
	"github.com/millken/crawlprop/utils"
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
	var option core.Option
	name := c.URL.Get("name")
	target := c.URL.Get("target")
	concurrent := utils.StrToInt(c.URL.Get("concurrent"))
	if concurrent > 0 {
		option.TabOpens = concurrent
	}

	allowHost := c.URL.Get("allow_host")
	if allowHost != "" {
		option.AllowHost = strings.Split(allowHost, ",")
	}

	uuid := UUID()
	err = stored.Create(uuid, name, target, option)
	data = gin.H{"task_id": uuid}
	return data, err
}
