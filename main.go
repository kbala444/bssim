package main

//  should validate config

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"
	blocks "github.com/ipfs/go-ipfs/blocks"
	key "github.com/ipfs/go-ipfs/blocks/key"
	bs "github.com/ipfs/go-ipfs/exchange/bitswap"
	tn "github.com/ipfs/go-ipfs/exchange/bitswap/testnet"
	splitter "github.com/ipfs/go-ipfs/importer/chunk"
	mocknet "github.com/ipfs/go-ipfs/p2p/net/mock"
	mockrouting "github.com/ipfs/go-ipfs/routing/mock"
	delay "github.com/ipfs/go-ipfs/thirdparty/delay"
	testutil "github.com/ipfs/go-ipfs/util/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var config map[string]string
var currLine int = 1
var net mocknet.Mocknet
var peers []bs.Instance
var deadline time.Duration
var dummy *DummyHandler
var recorder *Recorder

//  Map of files to the keys of the blocks that make it
var files = make(map[string][]key.Key)

func main() {
	fmt.Println(time.Now())
	var file *os.File
	var err error

	filename := flag.String("wl", "samples/star", "specifies the workload to run")
	prom := flag.Bool("prom", false, "if true, posts prometheus metrics to localhost:8080/metrics")
	override := getOptFlags()

	file, err = os.Open(*filename)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)

	//  parse first (config) line
	scanner.Scan()
	configure(scanner.Text(), override)
	currLine++
	fmt.Println(config)

	recorder = NewRecorder("data/metrics")

	net, peers = createTestNetwork()

	if *prom {
		go func() {
			http.Handle("/metrics", prometheus.Handler())
			http.ListenAndServe(":8080", nil)
		}()
	}

	//  execute commands
	for scanner.Scan() {
		err := execute(scanner.Text())
		check(err)
		currLine++
	}

	err = scanner.Err()
	check(err)

	//  Clean up dummy files if used
	if dummy != nil {
		dummy.DeleteFiles()
	}

	recorder.Close(file.Name())
	fmt.Println(time.Now())
}

//  Configure simulation based on first line of cmd file
func configure(cfgString string, override map[string]string) {
	//  Initialize config to default values
	config = map[string]string{
		"node_count":       "10",
		"visibility_delay": "0",
		"query_delay":      "0",
		"block_size":       strconv.Itoa(splitter.DefaultBlockSize),
		"deadline":         "600",
		"latency":          "0",
		"bandwidth":        "1000",
		"manual_links":		 "false",
	}

	if len(cfgString) > 1 {
		opts := strings.Split(cfgString, ",")
		for _, str := range opts {
			str = strings.TrimSpace(str)
			split := strings.Split(str, ":")
			if len(split) != 2 {
				log.Fatalf("Invalid config.")
			}
			k, v := strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
			config[k] = v
		}
	}

	//  override config with flags
	for k, v := range override {
		if v != "none" {
			config[k] = v
		}
	}

	d, err := strconv.ParseFloat(config["deadline"], 32)
	if err != nil {
		log.Fatalf("Invalid deadline.")
	}
	deadline = time.Duration(d) * time.Second
}

func execute(cmdString string) error {
	//  Check for comment
	if cmdString[0] == '#' {
		return nil
	}

	if strings.Contains(cmdString, "->") {
		return connectCmd(cmdString)
	}

	//  Check for dummy file command, could move some of this into a method in dummyfiles.go
	split := strings.Split(cmdString, " ")
	if split[0] == "create_dummy_files" {
		numfiles, err := strconv.Atoi(split[1])
		if err != nil {
			log.Fatalf("Line %d: Invalid argument for create_dummy_files.", currLine)
		}

		filesize, err := strconv.Atoi(split[2])
		if err != nil {
			log.Fatalf("Line %d: Invalid argument for create_dummy_files.", currLine)
		}

		createDummyFiles(numfiles, filesize)
		return nil
	}

	command := split[1]
	arg := split[2]

	//  Command in form "node# get/put/leave arg"
	nodes := getRange(split[0])
	switch command {
	case "putb":
		return putCmd(nodes, blocks.NewBlock([]byte(arg)))
	case "put":
		return putFileCmd(nodes, arg)
	case "getb":
		return getCmd(nodes, blocks.NewBlock([]byte(arg)))
	case "get":
		return getFileCmd(nodes, arg)
	case "leave":
		return leaveCmd(nodes, arg)
	default:
		return fmt.Errorf("Error on line %d: expected get/put/leave, found %s.", currLine, command)
	}
}

//  node1->node2 latency bw
func connectCmd(cmd string) error {
	split := strings.Split(cmd, " ")
	if len(split) != 3 {
		return fmt.Errorf("Line %d: Invalid number of arguments.", currLine)
	}
	nodes := strings.Split(split[0], "->")
	node1, err := ParseRange(nodes[0])
	if err != nil {
		return fmt.Errorf("Line %d: Invalid first node # or range.", currLine)
	}

	node2, err := strconv.Atoi(nodes[1])
	if err != nil {
		return fmt.Errorf("Line %d: Invalid second node # or range.", currLine)
	}
	
	if config["manual_links"] == "true"{
		link(node1, node2)
	}

	latencyFloat, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		return fmt.Errorf("Line %d: Invalid latency.", currLine)
	}

	latency := time.Millisecond * time.Duration(latencyFloat)

	bw, err := strconv.ParseFloat(split[2], 64)
	if err != nil {
		return fmt.Errorf("Line %d: Invalid bandwidth.", currLine)
	}

	//  convert bw from mbps to bps
	bw = bw * 1024 * 1024 / 8
	for _, node := range node1 {
		links := net.LinksBetweenPeers(peers[node].Peer, peers[node2].Peer)
		for _, link := range links {
			link.SetOptions(mocknet.LinkOptions{Bandwidth: bw, Latency: latency})
		}
	}
	return nil
}

func link(connecting []int, dest int){
	for _, node := range connecting {
		//  do i need the opposite command as well?
		net.LinkPeers(peers[node].Peer, peers[dest].Peer)
		//net.LinkPeers(peers[node].Peer, peers[dest].Peer)
	}
}

//  Unlinks peers from network
func leaveCmd(nodes []int, afterStr string) error {
	after, err := strconv.Atoi(afterStr)
	if err != nil {
		log.Fatalf("Line %d: Invalid argument to leave.", currLine)
	}

	time.AfterFunc(time.Second*time.Duration(after), func() {
		for _, n := range nodes {
			currQuitter := peers[n].Peer
			for _, p := range peers[n+1:] {
				err = net.UnlinkPeers(currQuitter, p.Peer)
				if err != nil {
					return
				}
			}
		}
	})
	return err
}

//  Chunks file into blocks and adds each block to exchange
func putFileCmd(nodes []int, file string) error {
	reader, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Line %d: Failed to open file '%s'.", currLine, file)
	}

	bsize, err := strconv.Atoi(config["block_size"])
	if err != nil {
		return fmt.Errorf("Invalid block size in config.")
	}
	chunks := (&splitter.SizeSplitter{Size: bsize}).Split(reader)

	files[file] = make([]key.Key, 0)
	//  waitgroup for chunks
	var wg sync.WaitGroup
	for chunk := range chunks {
		wg.Add(1)
		block := blocks.NewBlock(chunk)
		files[file] = append(files[file], block.Key())
		go func(block *blocks.Block) {
			err := putCmd(nodes, block)
			check(err)
			wg.Done()
		}(block)
	}
	wg.Wait()
	return nil
}

func putCmd(nodes []int, block *blocks.Block) error {
	for _, node := range nodes {
		err := peers[node].Exchange.HasBlock(context.Background(), block)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFileCmd(nodes []int, file string) error {
	blocks, ok := files[file]
	if !ok {
		return fmt.Errorf("Tried to get file, '%s', which has not been added.\n", file)
	}
	var wg sync.WaitGroup
	//  Get blocks and then Has them
	for _, node := range nodes {
		//  remove blocks peer already has or nah?
		//  I'm assuming that peers with the first block of the file have the whole file,
		//  which i think is ok for the simulation, but i might have to change this later
		alreadyhas, err := peers[node].Blockstore().Has(files[file][0])
		check(err)

		if alreadyhas {
			continue
		}
		wg.Add(1)
		go func(i int) {
			timer := recorder.NewTimer()
			ctx, _ := context.WithTimeout(context.Background(), deadline)
			received, _ := peers[i].Exchange.GetBlocks(ctx, blocks)

			for j := 0; j < len(blocks); j++ {
				blockTimer := recorder.NewTimer()
				x := <-received
				if x == nil {
					wg.Done()
					return
				}
				recorder.EndBlockTime(blockTimer, peers[i].Peer.Pretty())
				fmt.Println(i, x, j)
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				err := peers[i].Exchange.HasBlock(ctx, x)
				if err != nil {
					fmt.Println("error when adding block", i, err)
				}
			}
			recorder.EndFileTime(timer, peers[i].Peer.Pretty(), file)

			//	peers[i].Exchange.Close()
			wg.Done()
		}(node)
	}

	wg.Wait()
	testGet(nodes, file)
	return nil
}

func testGet(nodes []int, file string) {
	chunks, ok := files[file]
	if !ok {
		fmt.Printf("Tried check file, '%s', which has not been added.\n", file)
		return
	}
	fmt.Println("checking...")
	var wg sync.WaitGroup
	for _, node := range nodes {
		for _, chunk := range chunks {
			wg.Add(1)
			go func(i int, block key.Key) {
				has, err := peers[i].Blockstore().Has(block)
				check(err)
				if !has {
					fmt.Printf("Line %d: Node %d failed to get block %v\n", currLine, i, block)
				}
				wg.Done()
			}(node, chunk)
		}
	}
	wg.Wait()
	fmt.Println("done checking")
}

func getCmd(nodes []int, block *blocks.Block) error {
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(i int) {
			ctx, _ := context.WithTimeout(context.Background(), deadline)
			peers[i].Exchange.GetBlock(ctx, block.Key())
			fmt.Printf("Gotem from node %d.\n", i)
			peers[i].Exchange.Close()
			wg.Done()
		}(node)
	}

	wg.Wait()
	return nil
}

//  Create and distribute dummy files among existing peers
func createDummyFiles(n int, size int) {
	dummy = NewDummyHandler(n, size)
	dummy.distributeFiles(peers)
}

//  Creates test network using delays in config
//  Returns a fully connected mocknet and an array of the instances in the network
func createTestNetwork() (mocknet.Mocknet, []bs.Instance) {
	vv := convertTimeField("visibility_delay")
	q := convertTimeField("query_delay")
	//md := convertTimeField("message_delay")

	delayCfg := mockrouting.DelayConfig{ValueVisibility: vv, Query: q}
	n, err := strconv.Atoi(config["node_count"])
	check(err)
	mn := mocknet.New(context.Background())
	snet, err := tn.StreamNet(context.Background(), mn, mockrouting.NewServerWithDelay(delayCfg))
	check(err)
	instances := genInstances(n, &mn, &snet)
	return mn, instances
}

//  Adds random identities to the mocknet, creates bitswap instances for them, and links + connects them
func genInstances(n int, mn *mocknet.Mocknet, snet *tn.Network) []bs.Instance {
	instances := make([]bs.Instance, 0)
	for i := 0; i < n; i++ {
		peer, err := testutil.RandIdentity()
		check(err)
		_, err = (*mn).AddPeer(peer.PrivateKey(), peer.Address())
		check(err)
		inst := bs.Session(context.Background(), *snet, peer)
		instances = append(instances, inst)
	}

	bps, err := strconv.ParseFloat(config["bandwidth"], 64)
	if err != nil {
		log.Fatalf("Invalid bandwidth in config.")
	}
	//  Convert bandwidth from megabits/s to bytes/s
	bps = bps * 1024 * 1024 / 8

	lat, err := strconv.ParseFloat(config["latency"], 64)
	if err != nil {
		log.Fatalf("Invalid latency in config.")
	}
	(*mn).SetLinkDefaults(mocknet.LinkOptions{Latency: time.Duration(lat) * time.Millisecond, Bandwidth: bps})
	if config["manual_links"] == "false"{
		(*mn).LinkAll()
	}
	return instances
}

//  Converts config field to delay
func convertTimeField(field string) delay.D {
	val, err := strconv.Atoi(config[field])
	if err != nil {
		log.Fatalf("Invalid value for %s.", field)
	}
	return delay.Fixed(time.Duration(val) * time.Millisecond)
}

func getRange(s string) []int {
	nodes, err := ParseRange(s)
	if err != nil {
		log.Fatalf("Line %d: %v.", currLine, err)
	}
	//  todo: refactor all of these node_count conversions into something cleverer
	n, err := strconv.Atoi(config["node_count"])
	if err != nil {
		log.Fatalf("Invalid node_count.")
	}
	if nodes[len(nodes)-1] > n-1 || n < 0 {
		log.Fatalf("Line %d: Range out of bounds (max node number is %d).", currLine, n-1)
	}
	return nodes
}

//  I should probably find a way to not copy paste this
//  lifted from dhtHell
func ParseRange(s string) ([]int, error) {
	if len(s) == 0 {
		return nil, errors.New("no input given")
	}
	if s[0] == '[' && s[len(s)-1] == ']' {
		parts := strings.Split(s[1:len(s)-1], "-")
		if len(parts) == 0 {
			return nil, errors.New("No value in range!")
		}
		if len(parts) == 1 {
			n, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			return []int{n}, nil
		}
		low, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		high, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		var out []int
		for i := low; i <= high; i++ {
			out = append(out, i)
		}
		return out, nil

	} else {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return []int{n}, nil
	}
}

func getOptFlags() map[string]string {
	bw := flag.String("bw", "none", "overrides workload bandwidth")
	lat := flag.String("lat", "none", "overrides workload latency")
	flag.Parse()
	d := make(map[string]string, 0)
	d["bandwidth"] = *bw
	d["latency"] = *lat
	return d
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
