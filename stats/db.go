package stats

//DecoyStats ...
type DecoyStats struct {
	Goodput, Badput int
	Leafs, Nonleafs int
	PagesTotal      int
	FinalDepths     []int
}

//PageStats ...
type PageStats struct {
	Hostname string
	Depth    int
	Final    bool
}

//ResourceStats ...
type ResourceStats struct {
	Hostname  string
	IsGoodput bool
	Size      int
}
