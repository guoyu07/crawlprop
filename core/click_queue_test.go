package core_test

import (
	"fmt"
	"testing"

	"github.com/millken/crawlprop/core"
)

func TestClickQueue(t *testing.T) {
	var s *core.ClickQueue
	s = core.NewClickQueue()

	s.Push("a", 1, 1)
	var count int = s.Count()
	if count != 1 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1, err := s.Pull()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", r1)

	s.Push("b", 2, 2)
	s.Push("c", 3, 3)
	count = s.Count()
	if count != 2 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1, err = s.Pull()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", r1)
	r1, err = s.Pull()
	if err != nil {
		t.Error(err)
	}
	r1, err = s.Pull()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\n", r1)

}
