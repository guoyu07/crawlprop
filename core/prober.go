package core

import (
	"fmt"

	"github.com/millken/crawlprop/stats"
)

const maxResourceBufferSize = int(16 * 1024 * 1024)
const maxTotalBufferSize = maxResourceBufferSize * 4
const maxPostDataSize = 10240
const pageLimit = 100

type Prober struct {
	stats             map[string]*stats.DecoyStats
	resourceStatsChan chan (*stats.ResourceStats)
	pageStatsChan     chan (*stats.PageStats)
}

func NewProber() *Prober {
	p := &Prober{
		stats:             make(map[string]*stats.DecoyStats),
		resourceStatsChan: make(chan *stats.ResourceStats),
		pageStatsChan:     make(chan *stats.PageStats),
	}
	return p
}

func (p *Prober) Start() {

	done := make(chan struct{})

	go func() {
		for {
			select {
			case resStats := <-p.resourceStatsChan:
				if _, ok := p.stats[resStats.Hostname]; !ok {
					p.stats[resStats.Hostname] = &stats.DecoyStats{}
				}
				if resStats.IsGoodput {
					p.stats[resStats.Hostname].Goodput += resStats.Size
					p.stats[resStats.Hostname].Leafs++
				} else {
					p.stats[resStats.Hostname].Badput += resStats.Size
					p.stats[resStats.Hostname].Nonleafs++
				}
			case pageStat := <-p.pageStatsChan:
				if _, ok := p.stats[pageStat.Hostname]; !ok {
					p.stats[pageStat.Hostname] = &stats.DecoyStats{}
				}
				p.stats[pageStat.Hostname].PagesTotal++
				if pageStat.Final {
					p.stats[pageStat.Hostname].FinalDepths =
						append(p.stats[pageStat.Hostname].FinalDepths, pageStat.Depth)
				}
			case <-done:
				return
			}
		}
	}()

	decoys := []string{"http://demo.aisec.cn/demo/aisec/"}
	for _, decoy := range decoys {
		p.Decoy(decoy)
	}
	done <- struct{}{}

}

func (p *Prober) Decoy(decoyUrl string) {
}

func (p *Prober) Results() {
	for k, v := range p.stats {
		allput := v.Goodput + v.Badput
		ratio := float64(0)
		if allput != 0 {
			ratio = float64(v.Goodput) / float64(allput) * 100
		}
		fmt.Printf("[%s] total_bytes: %v goodput_bytes: %v ratio: %v%% "+
			" pages: %v"+
			" leafs: %v non-leafs: %v"+
			" depths: %v\n",
			k, allput, v.Goodput, ratio,
			v.PagesTotal,
			v.Leafs, v.Nonleafs,
			v.FinalDepths)
	}
}
