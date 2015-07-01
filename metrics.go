package main

import (
	"errors"
	"fmt"
	"time"
	"encoding/json"
	"os"
	"log"
	bs "github.com/ipfs/go-ipfs/exchange/bitswap"
	"github.com/ipfs/go-ipfs/p2p/peer"
	prom "github.com/prometheus/client_golang/prometheus"
	//"net/http"
)

var (
	//  NewSummaryVec(opts, ["filename", "pid", "latency", "bandwidth"]
	fileTimes = prom.NewHistogramVec(prom.HistogramOpts{
		Name: "file_times_ms",
		Help: "Time for peer to get a file.",
		Buckets: prom.LinearBuckets(1, .25, 12),
	}, []string{"latency", "bandwidth"})
	
	blockTimes = prom.NewHistogramVec(prom.HistogramOpts{
		Name: "block_times_ms",
		Help: "Time for peer to get a block.",
		Buckets: prom.ExponentialBuckets(0.005, 10, 10),
		//Buckets: prom.LinearBuckets(0, .05, 100),
	}, []string{"latency", "bandwidth"})
)

func init() {
	prom.MustRegister(fileTimes)
	prom.MustRegister(blockTimes)
}

type Recorder struct {
	currID int
	times map[int]time.Time
	log *os.File
	data map[int]*stats
	//  block info per bs instance
	bi map[peer.ID]*blockInfo
}

type stats struct {
	PeerID string
	File string
	Time float64
	//  downloaded from?
}

type blockInfo struct {
	//  maybe keep array of all times to graph?
	//  totalBlocks int
	totalTime time.Duration
	max time.Duration
	min time.Duration
}

func NewRecorder() *Recorder{
	//var oldData stats
	logFile, err := os.Open("stats.json")
	if os.IsNotExist(err){
		logFile, err = os.Create("stats.json")
		if err != nil{
			log.Fatalf("Could not create stats.json.", err)
		}
	}
		
	return &Recorder{
		currID: 0,
		times: make(map[int]time.Time),
		log: logFile,
		data: make(map[int]*stats),
		bi: make(map[peer.ID]*blockInfo),
		//data: oldData,
	}
}

func (r *Recorder) Close(){
	r.log.Close()
	logFile, err := os.Create("stats.json")
	if err != nil{
		log.Fatalf("Could not open stats.json for writing.", err)
	}
	encoder := json.NewEncoder(logFile)
	values := make([]*stats, 0)
	for _, v := range r.data{
		values = append(values, v)
	}
	err = encoder.Encode(values)
	if err != nil{
		log.Fatalf("Couldn't dump map to stats.json.", err)
	}
}

//  Creates and starts a new timer.
//  Returns id of timer which should be given to an EndTime method to stop it and record the time.
func (r *Recorder) NewTimer() int{
	r.currID += 1	
	r.times[r.currID] = time.Now()
	return r.currID
}

//  Accepts a timer ID (given by StartFileTime) and records the elapsed time.
func (r *Recorder) EndFileTime(id int, pid string, filename string) {
	elapsed := time.Since(r.times[id])
	delete(r.times, id)
	r.data[r.currID] = &stats{PeerID: pid, File: filename, Time: elapsed.Seconds()}
	fileTimes.WithLabelValues(config["latency"], config["bandwidth"]).Observe(elapsed.Seconds() * 1000)
}

func (r *Recorder) EndBlockTime(id int, pid peer.ID) {
	elapsed := time.Since(r.times[id])
	delete(r.times, id)
	curr, ok := r.bi[pid]
	if !ok{
		r.bi[pid] = &blockInfo{totalTime:elapsed}
	} else {
		curr.totalTime += elapsed
	}
	blockTimes.WithLabelValues(config["latency"], config["bandwidth"]).Observe(elapsed.Seconds() * 1000)
}

//  Returns mean block request fulfillment time in ms of a bs instance
func (r *Recorder) MeanBlockTime(inst bs.Instance) (float64, error){
	s, err := inst.Exchange.Stat()
	if err != nil{
		return 0, fmt.Errorf("Couldn't get stats for peer %v.", inst.Peer)
	}
	if s.BlocksReceived == 0{
		return 0, errors.New("No blocks for peer.")
	}
	return (r.bi[inst.Peer].totalTime.Seconds() * 1000)/float64(s.BlocksReceived), nil
}

//  Returns mean block request fulfillment time in ms across all bs instances
func (r *Recorder) TotalMeanBlockTime(insts []bs.Instance) float64{
	var t time.Duration
	var b int
	for _, inst := range insts{
		s, err := inst.Exchange.Stat()
		if err != nil{
			fmt.Println("Couldn't get stats for peer ", inst.Peer)
			continue
		}
		if s.BlocksReceived == 0{
			continue
		}
		t += r.bi[inst.Peer].totalTime
		b += s.BlocksReceived
	}
	return (t.Seconds() * 1000)/float64(b)
}

func TotalBlocksReceived(peers []bs.Instance) int{
	var blocks int
	for _, p := range peers{
		pstat, err := p.Exchange.Stat()
		if err != nil{
			fmt.Println("Unable to get stats from peer ", p.Peer)
		}
		blocks += pstat.BlocksReceived
	}
	return blocks
}

func DupBlocksReceived(peers []bs.Instance) int{
	var blocks int
	for _, p := range peers{
		pstat, err := p.Exchange.Stat()
		if err != nil{
			fmt.Println("Unable to get stats from peer ", p.Peer)
		}
		blocks += pstat.DupBlksReceived
	}
	return blocks
}
