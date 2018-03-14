package scheduler

import (
	"container/list"
	"crypto/md5"
	"net/url"
	"sync"
)

type QueueScheduler struct {
	locker *sync.Mutex
	rm     bool
	rmKey  map[[md5.Size]byte]bool
	queue  *list.List
}

func NewQueueScheduler(rmDuplicate bool) *QueueScheduler {
	queue := list.New()
	rmKey := make(map[[md5.Size]byte]bool)
	locker := new(sync.Mutex)
	return &QueueScheduler{rm: rmDuplicate, queue: queue, rmKey: rmKey, locker: locker}
}

func (c *QueueScheduler) Push(s string) {
	req := s
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	q := u.Query()
	m, _ := url.ParseQuery(u.RawQuery)
	for a, _ := range m {
		q.Set(a, "")
	}
	u.RawQuery = q.Encode()

	c.locker.Lock()
	var key [md5.Size]byte
	if c.rm {
		key = md5.Sum([]byte(u.String()))
		if _, ok := c.rmKey[key]; ok {
			c.locker.Unlock()
			return
		}
	}
	c.queue.PushBack(req)
	if c.rm {
		c.rmKey[key] = true
	}
	c.locker.Unlock()
}

func (c *QueueScheduler) Poll() string {
	c.locker.Lock()
	if c.queue.Len() <= 0 {
		c.locker.Unlock()
		return ""
	}
	e := c.queue.Front()
	s := e.Value.(string)
	c.queue.Remove(e)
	c.locker.Unlock()
	return s
}

func (c *QueueScheduler) Count() int {
	c.locker.Lock()
	len := c.queue.Len()
	c.locker.Unlock()
	return len
}
