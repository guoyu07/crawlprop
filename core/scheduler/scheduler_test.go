package scheduler_test

import (
	"fmt"
	"testing"

	"github.com/millken/crawlprop/core/scheduler"
)

func TestQueueScheduler(t *testing.T) {
	var s *scheduler.QueueScheduler
	s = scheduler.NewQueueScheduler(false)
	r := "http://a.com/?a=b&d=f"
	s.Push(r)
	var count int = s.Count()
	if count != 1 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1 := s.Poll()
	if r1 == "" {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)

	// remove duplicate
	s = scheduler.NewQueueScheduler(true)

	r2 := "http://b.com/"
	s.Push(r)
	s.Push(r2)
	s.Push(r)
	count = s.Count()
	if count != 2 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1 = s.Poll()
	if r1 == "" {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)
	r1 = s.Poll()
	if r1 == "" {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)

}
