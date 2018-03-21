package core

import (
	"container/list"
	"fmt"
	"sync"
)

type UrlPosition struct {
	URL  string
	X, Y int
}
type ClickQueue struct {
	mu    *sync.Mutex
	queue *list.List
}

func NewClickQueue() *ClickQueue {
	queue := list.New()
	mu := new(sync.Mutex)
	c := &ClickQueue{
		queue: queue,
		mu:    mu,
	}
	return c
}

func (c *ClickQueue) Push(clickUrl string, x, y int) {
	c.mu.Lock()
	up := &UrlPosition{
		URL: clickUrl,
		X:   x,
		Y:   y,
	}
	c.queue.PushBack(up)
	c.mu.Unlock()
}

func (c *ClickQueue) Pull() (*UrlPosition, error) {
	c.mu.Lock()
	if c.queue.Len() <= 0 {
		c.mu.Unlock()
		return nil, fmt.Errorf("the queue was empty")
	}
	e := c.queue.Front()
	s := e.Value.(*UrlPosition)
	c.queue.Remove(e)
	c.mu.Unlock()
	return s, nil
}

func (c *ClickQueue) Count() int {
	c.mu.Lock()
	len := c.queue.Len()
	c.mu.Unlock()
	return len
}
