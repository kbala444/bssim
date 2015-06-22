package main

import (
	"fmt"
	"time"
	"encoding/json"
	"os"
	"log"
)

type Recorder struct {
	currID int
	times map[int]time.Time
	log *os.File
	data map[int]*stats
}

type stats struct {
	PeerID string
	File string
	Time float64
	//  downloaded from?
}

func NewRecorder() *Recorder{
	//var oldData stats
	fmt.Println("i want this package")
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
		fmt.Println(v)
		values = append(values, v)
	}
	err = encoder.Encode(values)
	if err != nil{
		log.Fatalf("Couldn't dump map to stats.json.", err)
	}
}

func (r *Recorder) StartTime(pid string, filename string) int{
	r.currID += 1	
	r.times[r.currID] = time.Now()
	r.data[r.currID] = &stats{PeerID: pid, File: filename}
	return r.currID
}

func (r *Recorder) EndTime(id int) {
	elapsed := time.Since(r.times[id])
	delete(r.times, id)
	curr, ok := r.data[id]
	if !ok{
		fmt.Println("No matching start time for end time.")
		return
	} 
	curr.Time = elapsed.Seconds()
}
