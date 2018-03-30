package core

type Option struct {
	TabOpens  int      //allow open at the same time
	AllowHost []string //allow host for crawler
	Schedule  string
	MaxCrawls int //maximum grabbing number

}
