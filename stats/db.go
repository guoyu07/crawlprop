package stats

type DecoyStats struct {
	goodput, badput int
	leafs, nonleafs int
	pages_total     int
	final_depths    []int
}

type PageStats struct {
	hostname string
	depth    int
	final    bool
}

type ResourceStats struct {
	hostname  string
	isGoodput bool
	size      int
}