package db

import (
	"fmt"

	"github.com/go-redis/redis"
)

type Db struct {
	taskID string
	client *redis.Client
}

func NewDb(taskID string, client *redis.Client) *Db {
	return &Db{
		taskID: taskID,
		client: client,
	}
}

func (c *Db) State(state string) error {
	field := "state"
	return c.SetInfo(field, state)
}

func (c *Db) SetInfo(field string, value string) error {
	key := fmt.Sprintf("%s:info", c.taskID)
	hSet := c.client.HSet(key, field, value)
	return hSet.Err()
}

func (c *Db) AddLink(u string) error {
	key := fmt.Sprintf("%s:link", c.taskID)
	sAdd := c.client.SAdd(key, u)
	return sAdd.Err()
}

func (c *Db) AddForm(action, method, data string) error {
	key := fmt.Sprintf("%s:form", c.taskID)
	u := fmt.Sprintf("%s\t%s\t%s", action, method, data)
	sAdd := c.client.SAdd(key, u)
	return sAdd.Err()
}
