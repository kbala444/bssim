// test
package main

import (
	"time"
	"testing"
	key "github.com/ipfs/go-ipfs/blocks/key"
	//context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"
	//"time"
	blocks "github.com/ipfs/go-ipfs/blocks"
	//"fmt"
	"os"
)

func TestDelays(t *testing.T){
	
}

func TestPutBlockCmd(t *testing.T){
	configure("node_count:1")
	net, peers = createTestNetwork()
	
	testBlock := blocks.NewBlock([]byte("testblock"))
	err := putCmd([]int{0}, testBlock)
	check(err)
	
	if !checkHasBlock([]int{0}, testBlock.Key()){
		t.Error("Peer 0 never add block to blockstore.")
	}
}

func TestPutFileCmd(t *testing.T){
	configure("node_count:3")
	net, peers = createTestNetwork()
	
	//  test single node
	err := putFileCmd([]int{0}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{0}, block){
			t.Error("Peer 0 added file but it's not in its blockstore.")
		}
	}
	
	//  test multiple nodes
	err = putFileCmd([]int{1, 2}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{1, 2}, block){
			t.Error("Peers 1 and 2 added file but it's not in its blockstore.")
		}
	}
	
	//  test putting same file again
	err = putFileCmd([]int{0}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{0}, block){
			t.Error("Peer 0 added file but it's not in its blockstore.")
		}
	}
	
	//  test single block file
	err = putFileCmd([]int{0}, "samples/test.txt")
	check(err)
	for _, block := range files["samples/test.txt"] {
		if !checkHasBlock([]int{0}, block){
			t.Error("Peer 0 added file but it's not in its blockstore.")
		}
	}	
	
	//  try adding nonexistent file
	err = putFileCmd([]int{1}, "samples/ghost.mp3")
	//  expect path error (no such file or directory)
	if _, ok := err.(*os.PathError); ok{
		t.Error("Somehow put nonexistent file.");
	}
	
	//  ensure it hasn't been added to files map
	if _, ok := files["samples/ghost.mp3"]; ok{
		t.Error("Somehow added nonexistent file to files map.");
	}
}

func TestGetFileCmd(t *testing.T){
	configure("node_count:2")
	net, peers = createTestNetwork()
	
	err := putFileCmd([]int{0}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{0}, block){
			t.Error("Peer 0 added file but it's not in its blockstore.")
		}
	}
	
	err = getFileCmd([]int{1}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{1}, block){
			t.Error("Peer 0 got file but it's not in its blockstore.")
		}
	}
	
	//  try getting same file again
	err = getFileCmd([]int{1}, "samples/test.mp3")
	check(err)
	for _, block := range files["samples/test.mp3"] {
		if !checkHasBlock([]int{1}, block){
			t.Error("Peer 0 got file but it's not in its blockstore.")
		}
	}
	
	//  try getting nonexistent file
	err = getFileCmd([]int{1}, "samples/ghost.mp3")
	if err == nil{
		t.Error("Somehow got nonexistent file.");
	}
}

func TestLeaveCmd(t *testing.T){
	configure("node_count:2, deadline:0.25")
	net, peers = createTestNetwork()
	
	err := putFileCmd([]int{0}, "samples/test.txt")
	check(err)
	
	err = leaveCmd([]int{0}, "0")
	check(err)
	//  wait for node to unlink
	time.Sleep(time.Millisecond)
	err = getFileCmd([]int{1}, "samples/test.txt")
	check(err)		
	for _, block := range files["samples/test.txt"] {
		if checkHasBlock([]int{1}, block){
			t.Error("Peer 1 got file after peer 0 left.")
		}
	}
	
	configure("node_count:2, deadline:0.25")
	net, peers = createTestNetwork()
	
	err = putFileCmd([]int{0}, "samples/test.txt")
	check(err)
	
	err = leaveCmd([]int{0}, "5")
	check(err)
	//  peer 1 should still be able to get the file
	err = getFileCmd([]int{1}, "samples/test.txt")
	check(err)		
	for _, block := range files["samples/test.txt"] {
		if !checkHasBlock([]int{1}, block){
			t.Error("Peer 1 couldn't get file.")
		}
	}
}

func checkHasBlock(nodes []int, block key.Key) bool {
	for _, node := range nodes{
		has, err := peers[node].Blockstore().Has(block)
		check(err)
		if !has{
			return false
		}
	}
	return true
}