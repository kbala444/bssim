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
	//peer "github.com/ipfs/go-ipfs/p2p/peer"
	//p2putil "github.com/ipfs/go-ipfs/p2p/test/util"
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
	ds "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	ds_sync "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs/blocks/blockstore"
	datastore2 "github.com/ipfs/go-ipfs/util/datastore2"
	//"github.com/ipfs/go-ipfs/util"
)

var config map[string]string
var currLine int = 1

var peers []bs.Instance

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
		file, err = os.Open("samples/star")
	}
	
    check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	
	//  get first (config) line
	scanner.Scan()
	configure(scanner.Text())
	currLine++
	
	net := createTestNetwork()
	g := bs.NewTestSessionGenerator(net)
	
	n, err := strconv.Atoi(config["n"])
	if err != nil {
		log.Fatal("Invalid number of nodes.")
	}
	
	peers = spawn(n, &g)
	for scanner.Scan() {
		err := execute(scanner.Text())
		check(err)
		currLine++
	}
	
	err = scanner.Err()
	check(err)
}

//  Configure simulation based on first line of cmd file
func configure(cfgString string){
	//  Initialize config to default values
	config = map[string]string{
		"n" : "10",
		"vv" : "0",
		"q" : "0",
		"md" : "0",
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
}

func execute(cmdString string) error{
	if cmdString[0] == '[' && strings.Contains(cmdString, "->"){
		return connectCmd(cmdString)
	}
	
	if cmdString[0] == '#'{
		return nil
	}
	
	//  get/put command
	split := strings.Split(cmdString, " ")
	if len(split) != 3 {
		return fmt.Errorf("Error on line %d: too few arguments.", currLine)
	}
	
	command := split[1]
	arg := split[2]
	if node, err := strconv.Atoi(split[0]); err == nil{
		//  Command in form "# cmd arg"
		switch command {
			case "putb": putCmd(node, blocks.NewBlock([]byte(arg)))
			case "getb": getCmd([]int{node}, blocks.NewBlock([]byte(arg)))
			case "put": putFileCmd(node, arg)
			case "get": getFileCmd([]int{node}, arg)
			default: return fmt.Errorf("Error on line %d: expected get or put, found %s.", currLine, command)
		}
	} else if cmdString[0] == '[' {
		//  Command in form "[#-#] getcmd arg"
		nodes, err := ParseRange(split[0])
		if err != nil {
			return err
		}
		switch command {
			case "getb": getCmd(nodes, blocks.NewBlock([]byte(arg)))
			case "get": getFileCmd(nodes, arg)
			default: return fmt.Errorf("Error on line %d: expected get, found %s.", currLine, command)
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

	chunks := splitter.DefaultSplitter.Split(reader)
	
	files[file] = make([]key.Key, 0)
	for chunk := range chunks {
		block := blocks.NewBlock(chunk)
		files[file] = append(files[file], block.Key())
		go func(block *blocks.Block){
			err := putCmd(node, block)
			if err != nil{
				//  should i recover this maybe?  is this even how you're supposed to use panic?
				panic(err)
			}
		}(block)
	}
	return nil
}

func getFileCmd(nodes []int, file string) error{
	blocks, ok := files[file]
	if !ok {
		fmt.Println("Tried to get file, '%s', which has not been added.", file)
		return nil
	}
	var wg sync.WaitGroup
	for _, node := range nodes{
		//  Get blocks and then Has them
		wg.Add(1)
		go func(i int){
			ctx, _ := context.WithTimeout(context.Background(), time.Minute)
			received, err := peers[i].Exchange.GetBlocks(ctx, blocks)
			if err != nil{
				panic(err)
			}
			for j := 0; j < len(blocks); j++{
				ctx, _ := context.WithTimeout(context.Background(), time.Minute)
				x := <-received
				fmt.Println(i, x, j)
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
	fmt.Println("checking")
	var wg sync.WaitGroup
	for _, node := range nodes{
		for _, chunk := range chunks{
			wg.Add(1)
			go func(i int, block key.Key){
				has, err := peers[i].Blockstore().Has(block)
				check(err)
				if !has{
					fmt.Println("Node %d failed to get block %v", i, block)
				}
				wg.Done()
			}(node, chunk)
		}
	}
	wg.Wait()
	fmt.Println("it's over")
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
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			peers[node].Exchange.GetBlock(ctx, block.Key())
			fmt.Printf("Gotem from node %d.\n", i)
			peers[node].Exchange.Close()			
			wg.Done()
		}(node)
	}

	wg.Wait()
	return nil
}

//  Creates test network using delays in config
func createTestNetwork() tn.Network {
	vv := convertTimeField("vv")
	q := convertTimeField("q")
	md := convertTimeField("md")
	fmt.Println(md)
		
	delayCfg := mockrouting.DelayConfig{ValueVisibility: vv, Query: q}
	n, err := strconv.Atoi(config["n"])
	check(err)
	mn, err := mocknet.FullMeshLinked(context.Background(), n)
	check(err)
	router := mockrouting.NewServerWithDelay(delayCfg)
	snet, err := tn.StreamNet(context.Background(), mn, router)
	check(err)
	peers = genInstances(n, &mn, &snet)
	return snet
	//return tn.VirtualNetwork(mockrouting.NewServerWithDelay(delayCfg), md)
}

func genInstances(n int, mn *mocknet.Mocknet, snet *tn.Network) []bs.Instance{
	instances := make([]bs.Instance, 0)
	for i := 0; i < n; i++{
		peer, err := testutil.RandIdentity()
		check(err)
		_, err = (*mn).AddPeer(peer.PrivateKey(), peer.Address())
		check(err)
		bsdelay := delay.Fixed(0)
		const kWriteCacheElems = 100
	
		adapter := (*snet).Adapter(peer)
		dstore := ds_sync.MutexWrap(datastore2.WithDelay(ds.NewMapDatastore(), bsdelay))
	
		bstore, err := blockstore.WriteCached(blockstore.NewBlockstore(ds_sync.MutexWrap(dstore)), kWriteCacheElems)
		check(err)
		
		xchg := bs.New(context.Background(), peer.ID(), adapter, bstore, true).(*bs.Bitswap)
	
		inst := bs.NewInstance(peer.ID(), xchg, bstore, bsdelay)
		//i := bs.Session(context.Background(), snet, peer)
		instances = append(instances, inst)
	}
	(*mn).ConnectAll()
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
