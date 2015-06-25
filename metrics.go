package main

import (
	"fmt"
	"time"
	"encoding/json"
	"os"
	"log"
	bs "github.com/ipfs/go-ipfs/exchange/bitswap"
)

type Recorder struct {
	currID int
	times map[int]time.Time
	log *os.File
	data map[int]*stats
	bi blockInfo
}

type stats struct {
	PeerID string
	File string
	Time float64
	//  downloaded from?
}

type blockInfo struct {
	totalBlocks int
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

func (r *Recorder) StartFileTime(pid string, filename string) int{
	r.currID += 1	
	r.times[r.currID] = time.Now()
	r.data[r.currID] = &stats{PeerID: pid, File: filename}
	return r.currID
}

func (r *Recorder) EndFileTime(id int) {
	elapsed := time.Since(r.times[id])
	delete(r.times, id)
	curr, ok := r.data[id]
	if !ok{
		fmt.Println("No matching start time for end time.")
		return
	} 
	curr.Time = elapsed.Seconds()
}

//  Returns mean block request fulfillment time in ms
func (r *Recorder) MeanBlockTime() float64{
	return (r.bi.totalTime.Seconds() * 1000)/float64(r.bi.totalBlocks)
}

func (r *Recorder) StartBlockTime() int{
	r.currID += 1
	r.times[r.currID] = time.Now()
	r.bi.totalBlocks += 1
	return r.currID
}

func (r *Recorder) EndBlockTime(id int) {
	t := time.Since(r.times[id])
	r.bi.totalTime += t
	if t > r.bi.max{
		r.bi.max = t
	} else if t < r.bi.min{
		r.bi.min = t
	}
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
