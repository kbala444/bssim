// class to handle prometheus metric recording
package main

import (
	"fmt"
	prom "github.com/heems/bssim/Godeps/_workspace/src/github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	fields = []string{"latency", "bandwidth", "block_size"}
	//  NewSummaryVec(opts, ["filename", "pid", "latency", "bandwidth"]
	fileTimes = prom.NewHistogramVec(prom.HistogramOpts{
		Name:    "file_times_ms",
		Help:    "Time for peer to get a file.",
		Buckets: prom.LinearBuckets(1, .25, 12),
	}, fields)

	blockTimes = prom.NewHistogramVec(prom.HistogramOpts{
		Name:    "block_times_ms",
		Help:    "Time for peer to get a block.",
		Buckets: prom.ExponentialBuckets(0.005, 10, 10),
		//Buckets: prom.LinearBuckets(0, .05, 100),
	}, fields)

	dupBlocks = prom.NewGaugeVec(prom.GaugeOpts{
		Name: "dup_blocks_count",
		Help: "Count of total duplicate blocks received.",
	}, fields)
)

type PromHandler struct {
	currLables prom.Labels
}

func (p *PromHandler) Start() {
	prom.MustRegister(fileTimes)
	prom.MustRegister(blockTimes)
	prom.MustRegister(dupBlocks)
}

func (p *PromHandler) Observe(field *prom.HistogramVec, elapsed time.Duration) {
	labels := p.getLabels()
	field.With(labels).Observe(elapsed.Seconds() * 1000)
	if field == fileTimes {
		p.updateDupBlocks()
	}
}

func (p *PromHandler) getLabels() prom.Labels {
	if p.currLables == nil {
		currLables := make(map[string]string, 0)
		for _, field := range fields {
			currLables[field] = config[field]
		}
		p.currLables = currLables
	}

	return p.currLables
}

func (p *PromHandler) updateDupBlocks() {
	var blocks int
	for _, p := range peers {
		pstat, err := p.Exchange.Stat()
		if err != nil {
			fmt.Println("Unable to get stats from peer ", p.Peer)
		}
		blocks += pstat.BlocksReceived
	}
	l := p.getLabels()
	dupBlocks.With(l).Set(float64(blocks))
}
