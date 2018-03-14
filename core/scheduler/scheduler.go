package scheduler

type Scheduler interface {
	Push(url string)
	Poll() string
	Count() int
}
