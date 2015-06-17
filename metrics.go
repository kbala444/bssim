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
	data stats
}

type stats map[string][]float64

func NewRecorder() *Recorder{
	var oldData stats
	logFile, err := os.Open("stats.json")
	if os.IsNotExist(err){
		logFile, err = os.Create("stats.json")
		if err != nil{
			log.Fatalf("Could not create stats.json.", err)
		}
		oldData = stats{}
	} else {
		decoder := json.NewDecoder(logFile)
		decoder.Decode(&oldData)
		if oldData == nil{
			fmt.Println("here")
			oldData = stats{}
		}
	}
	
	return &Recorder{
		currID: 0,
		times: make(map[int]time.Time),
		log: logFile,
		data: oldData,
	}
}

func (r *Recorder) Close(){
	r.log.Close()
	logFile, err := os.Create("stats.json")
	if err != nil{
		log.Fatalf("Could not open stats.json for writing.", err)
	}
	encoder := json.NewEncoder(logFile)
	err = encoder.Encode(r.data)
	if err != nil{
		log.Fatalf("Couldn't dump map to stats.json.", err)
	}
}

func (r *Recorder) StartTime() int{
	r.currID += 1	
	r.times[r.currID] = time.Now()
	return r.currID
}

func (r *Recorder) EndTime(id int, field string) {
	elapsed := time.Since(r.times[id])
	delete(r.times, id)
	_, ok := r.data[field]
	if !ok{
		r.data[field] = make([]float64, 0)
	} 
	r.data[field] = append(r.data[field], elapsed.Seconds())
}
