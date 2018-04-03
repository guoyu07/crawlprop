package api

import (
	"crypto/rand"
	"fmt"

	"github.com/millken/crawlprop/stored"

	"github.com/go-redis/redis"

	"github.com/gin-gonic/gin"

	"log"
	"net/url"
)

type Action interface {
	Response() (data gin.H, err error)
}

type ActionParam struct {
	URL url.Values
	Red *redis.Client
}

var Actions = map[string]func(*ActionParam) (Action, error){}

func Register(name string, actionFactory func(*ActionParam) (Action, error)) {
	if actionFactory == nil {
		panic(" actionFactory is nil")
	}
	if _, dup := Actions[name]; dup {
		panic(" Register called twice for " + name)
	}
	Actions[name] = actionFactory
}

func serverInit(c *gin.Context) {
	var data, response gin.H
	var err error
	get := url.Values{}
	c.Header("Access-Control-Allow-Origin", "*")
	get, _ = url.ParseQuery(string(c.Request.URL.RawQuery))

	action := get.Get("action")
	act, ok := Actions[action]
	if !ok {
		log.Printf("[ERROR] %s action not found", action)
		c.JSON(200, gin.H{"status": 404, "error": fmt.Sprintf("action not found: %s", action)})
		return
	}
	ap := &ActionParam{
		URL: get,
		Red: stored.RedisClient(),
	}
	a, _ := act(ap)
	response, err = a.Response()
	if err != nil {
		data = gin.H{
			"status": 501,
			"error":  fmt.Sprintf("%s", err),
		}
	} else {
		data = gin.H{
			"status": 200,
			"data":   response,
		}
	}
	c.JSON(200, data)

}

func UUID() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if n != len(b) {
		err = fmt.Errorf("Not enough entropy available")
	}
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
