// dummyfiles
package main

import (
	"fmt"
	bs "github.com/ipfs/go-ipfs/exchange/bitswap"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

type DummyHandler struct {
	//  Array of filenames handled by this DummyHandler
	files []string
	//  Size of each file in kb
	size int
}

//  Returns new dummy file handler with n files
func NewDummyHandler(n int, s int) *DummyHandler {
	dh := &DummyHandler{files: make([]string, 0), size: s}
	dh.CreateFiles(n, "/samples/dummy")
	return dh
}

// Creates n files with path /samples/dummy(i).  Files contents are from /dev/urandom.
func (dh *DummyHandler) CreateFiles(n int, suffix string) {
	bs := 64
	if dh.size < 64 {
		bs = dh.size
	}
	count := dh.size / bs
	dir, _ := os.Getwd()
	for i := 0; i < n; i++ {
		path := dir + suffix + strconv.Itoa(i)
		dd := "dd if=/dev/urandom of=" + path + " bs=" + strconv.Itoa(bs) + " count=" + strconv.Itoa(count)
		cmd := exec.Command("/bin/sh", "-c", dd)
		_, err := cmd.Output()
		if err != nil {
			log.Fatalf("Could not create dummy file", err)
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("Could not open new dummy file", i)
		}

		dh.files = append(dh.files, file.Name())
	}
}

func (dh *DummyHandler) distributeFiles(peers []bs.Instance) {
	dir, _ := os.Getwd()
	for _, file := range dh.files {
		//  get rel path
		file = file[len(dir)+1:]
		n := rand.Intn(len(peers))
		putFileCmd([]int{n}, file)
	}
}

//  Removes all files belonging to this handler
func (dh *DummyHandler) DeleteFiles() {
	for _, file := range dh.files {
		err := os.Remove(file)
		if err != nil {
			fmt.Println("Failed to delete file ", file)
		}
	}
}
