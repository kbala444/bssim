package main
//  should validate config


import (
	"sync"
	tn "github.com/ipfs/go-ipfs/exchange/bitswap/testnet"
	mockrouting "github.com/ipfs/go-ipfs/routing/mock"
	delay "github.com/ipfs/go-ipfs/thirdparty/delay"
	blocks "github.com/ipfs/go-ipfs/blocks"
	key "github.com/ipfs/go-ipfs/blocks/key"
	bs "github.com/ipfs/go-ipfs/exchange/bitswap"
	context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"
	splitter "github.com/ipfs/go-ipfs/importer/chunk"
	mocknet "github.com/ipfs/go-ipfs/p2p/net/mock"
	testutil "github.com/ipfs/go-ipfs/util/testutil"
	//bsnet "github.com/ipfs/go-ipfs/exchange/bitswap/network"
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"errors"
	"log"
	"time"
)

var config map[string]string
var currLine int = 1

var net mocknet.Mocknet
var peers []bs.Instance
var deadline time.Duration
var dummy *DummyHandler

//  Map of files to the keys of the blocks that make it
var files = make(map[string][]key.Key)

func main() {
	var file *os.File
	var err error
	
	if len(os.Args) > 2{
		log.Fatalf("Too many arguments.")
	} else if len(os.Args) > 1{
		file, err = os.Open(os.Args[1])
	} else {
		file, err = os.Open("samples/lotsofiles")
	}
	
    check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	
	//  get first (config) line
	scanner.Scan()
	configure(scanner.Text())
	currLine++
	
	net, peers = createTestNetwork()
		
	for scanner.Scan() {
		err := execute(scanner.Text())
		check(err)
		currLine++
	}
	
	err = scanner.Err()
	check(err)
	
	//  Clean up if used
	if dummy != nil{
		dummy.DeleteFiles()
	}
}

//  Configure simulation based on first line of cmd file
func configure(cfgString string){
	//  Initialize config to default values
	config = map[string]string{
		"node_count" : "10",
		"visibility_delay" : "0",
		"query_delay" : "0",
		"block_size": strconv.Itoa(splitter.DefaultBlockSize),
		"deadline" : "60",
		//"message_delay" : "0",
		//"type" : "mock",
		//  add more options here later
	}
	
	opts := strings.Split(cfgString, ",")
	for _, str := range opts{
		str = strings.TrimSpace(str)
		split := strings.Split(str, ":")
		k, v := strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
		config[k] = v
	}
	
	d, err := strconv.Atoi(config["deadline"])
	if err != nil{
		log.Fatalf("Invalid deadline.")
	}
	deadline = time.Duration(d) * time.Minute
}

//  this is getting pretty bad
func execute(cmdString string) error{
	if cmdString[0] == '[' && strings.Contains(cmdString, "->"){
		return connectCmd(cmdString)
	}
	
	//  Check for comment
	if cmdString[0] == '#'{
		return nil
	}
	
	//  Check for dummy file command
	split := strings.Split(cmdString, " ")
	if (split[0] == "create_dummy_files"){
		numfiles, err := strconv.Atoi(split[1]);
		if err != nil{
			log.Fatalf("Line %d: Invalid argument for create_dummy_files.", currLine)
		}
		
		filesize, err := strconv.Atoi(split[2]);
		if err != nil{
			log.Fatalf("Line %d: Invalid argument for create_dummy_files.", currLine)
		}
		
		createDummyFiles(numfiles, filesize)
	}
	
	command := split[1]
	if node, err := strconv.Atoi(split[0]); err == nil{
		if len(split) < 3{
			if (command == "leave"){
				return leave([]int{node})
			} else {
				return fmt.Errorf("Line %d:  Expected leave, found %s.", currLine, command)
			}
		}
		arg := split[2]	
		//  Command in form "# cmd arg"
		switch command {
			case "putb": return putCmd(node, blocks.NewBlock([]byte(arg)))
			case "getb": return getCmd([]int{node}, blocks.NewBlock([]byte(arg)))
			case "put": return putFileCmd(node, arg)
			case "get": return getFileCmd([]int{node}, arg)
			default: return fmt.Errorf("Error on line %d: expected get or put, found %s.", currLine, command)
		}
	} else if cmdString[0] == '[' {
		//  Command in form "[#-#] get/leave arg"
		nodes, err := ParseRange(split[0])
		if err != nil {
			return err
		}
		switch command {
			case "getb": return getCmd(nodes, blocks.NewBlock([]byte(split[2])))
			case "get": return getFileCmd(nodes, split[2])
			case "leave": return leave(nodes)
			default: return fmt.Errorf("Error on line %d: expected get, found %s.", currLine, command)
		}
	}
	
	return nil
}

//  Unlinks peers from network
func leave(nodes []int) error{
	for _, n := range nodes{
		currQuitter := peers[n].Peer
		for _, p := range peers[n+1:]{
			err := net.UnlinkPeers(currQuitter, p.Peer)
			if err != nil{
				return err
			}
		}
	}
	return nil
}

//  Chunks file into blocks and adds each block to exchange
func putFileCmd(node int, file string) error{
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
	var wg sync.WaitGroup
	for chunk := range chunks {
		wg.Add(1)
		block := blocks.NewBlock(chunk)
		files[file] = append(files[file], block.Key())
		go func(block *blocks.Block){
			err := putCmd(node, block)
			if err != nil{
				//  should i recover this maybe?  is this even how you're supposed to use panic?
				panic(err)
			}
			wg.Done()
		}(block)
	}
	wg.Wait()
	return nil
}

func getFileCmd(nodes []int, file string) error{
	blocks, ok := files[file]
	if !ok {
		return fmt.Errorf("Tried to get file, '%s', which has not been added.\n", file)
	}
	
	var wg sync.WaitGroup
	//  Get blocks and then Has them
	for _, node := range nodes{
		//  remove blocks peer already has or nah?
		//  I'm assuming that peers with the first block of the file have the whole file,
		//  which i think is ok for the simulation, but i might have to change this later
		alreadyhas, err := peers[node].Blockstore().Has(files[file][0])
		check(err)
		if (alreadyhas){
			continue;
		}
		wg.Add(1)
		go func(i int){
			ctx, _ := context.WithTimeout(context.Background(), deadline)
			received, _ := peers[i].Exchange.GetBlocks(ctx, blocks)

			for j := 0; j < len(blocks); j++{
				x := <-received
				if x == nil{
					wg.Done();
					return;
				}	
				fmt.Println(i, x, j)
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				err := peers[i].Exchange.HasBlock(ctx, x)
				if err != nil{
					fmt.Println("error when adding block", i, err)
				}
			}
			
			//	peers[i].Exchange.Close()			
			wg.Done()
		}(node)
	}

	wg.Wait()
	testGet(nodes, file)
	return nil
}

func testGet(nodes []int, file string){
	chunks, ok := files[file]
	if !ok {
		fmt.Println("Tried check file, '%s', which has not been added.", file)
		return
	}
	fmt.Println("checking...")
	var wg sync.WaitGroup
	for _, node := range nodes{
		for _, chunk := range chunks{
			wg.Add(1)
			go func(i int, block key.Key){
				has, err := peers[i].Blockstore().Has(block)
				check(err)
				if !has{
					fmt.Printf("Line %d: Node %d failed to get block %v\n", currLine, i, block)
				}
				wg.Done()
			}(node, chunk)
		}
	}
	wg.Wait()
	fmt.Println("done checking")
}

func connectCmd(cmd string) error{
	split := strings.Split(cmd, "->")
	connecting, err := ParseRange(split[0])
	if err != nil  {
		return fmt.Errorf("Error in line %d: %v\n", currLine, err)
	}
	
	target, err := strconv.Atoi(split[1])
	if err != nil {
		return fmt.Errorf("Invalid target node in line:", currLine)
	}
	
	
	fmt.Println(target, connecting)
	
	return nil
}

func putCmd(node int, block *blocks.Block) error{
	hasBlock := peers[node]
	//defer hasBlock.Exchange.Close()
	err := hasBlock.Exchange.HasBlock(context.Background(), block);
	return err
}

func getCmd(nodes []int, block *blocks.Block) error{
	var wg sync.WaitGroup
	for _, node := range nodes{
		wg.Add(1)
		go func(i int){
			ctx, _ := context.WithTimeout(context.Background(), deadline)
			peers[node].Exchange.GetBlock(ctx, block.Key())
			fmt.Printf("Gotem from node %d.\n", i)
			peers[node].Exchange.Close()			
			wg.Done()
		}(node)
	}

	wg.Wait()
	return nil
}

//  Create and distribute dummy files among existing peers
func createDummyFiles(n int, size int){
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
	mn, err := mocknet.FullMeshLinked(context.Background(), n)
	check(err)
	snet, err := tn.StreamNet(context.Background(), mn, mockrouting.NewServerWithDelay(delayCfg))
	check(err)
	instances := genInstances(n, &mn, &snet)
	return mn, instances
}

//  Adds random identities to the mocknet, creates bitswap instances for them, and links + connects them
func genInstances(n int, mn *mocknet.Mocknet, snet *tn.Network) []bs.Instance{
	instances := make([]bs.Instance, 0)
	for i := 0; i < n; i++{
		peer, err := testutil.RandIdentity()
		check(err)
		_, err = (*mn).AddPeer(peer.PrivateKey(), peer.Address())
		check(err)
		inst := bs.Session(context.Background(), *snet, peer)
		instances = append(instances, inst)
	}
	(*mn).LinkAll()
	//(*mn).ConnectAll()
	return instances
}

//  Converts config field to delay
func convertTimeField(field string) delay.D{
	val, err := strconv.Atoi(config[field])
	if err != nil {
		log.Fatalf("Invalid value for %s.", field)
	}
	return delay.Fixed(time.Duration(val) * time.Millisecond)
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

func check(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

//  Creates array of n instances using SessionGenerator g
func spawn(n int, g *bs.SessionGenerator) []bs.Instance {
	instances := make([]bs.Instance, 0)
	for j := 0; j < n; j++ {
		inst := g.Next()
		instances = append(instances, inst)
	}
	return instances
}
