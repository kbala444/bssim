package testdialer

import (
	"strconv"
	"strings"
	manet "github.com/jbenet/go-multiaddr-net"
	ma "github.com/jbenet/go-multiaddr-net/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	"log"
	"fmt"
	"math"
	"time"
	//"bytes"
)

//  Abstract interface can be implemented by mock_conn.conn or net.Conn
type conn interface{
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

//  Represents connection with simulated bandwidth cap and latency
type CappedConn struct{
	Conn conn
	Bw float64	
	Lat time.Duration	
}

func (cc *CappedConn) Read(p []byte) (n int, err error){
	return cc.delayedReadOrWrite(false, p)
}

func (cc *CappedConn) Write(p []byte) (n int, err error){
	return cc.delayedReadOrWrite(true, p)
}

func (cc *CappedConn) Close() error{
	return cc.Conn.Close()
}

func (cc *CappedConn) delayedReadOrWrite(write bool, b []byte) (n int, err error){
	var f func(p []byte) (int, error)
	if write{
		f = cc.Conn.Write
	} else {
		f = cc.Conn.Read
	}
	
	hlat := cc.Lat/2
	//  Simulate first half of latency
	time.Sleep(hlat)
	
	//  Simulate bandwidth
	chunk_size := int(cc.Bw * 1024 * 1024)
	chunks := float64(len(b))/(cc.Bw * 1024 * 1024)
	num_chunks := int(math.Ceil(chunks + .5))
	fmt.Println("num_chunks", num_chunks, "chunk_size", chunk_size, "chunks", chunks, "len", cap(b), "message", string(b[:5]))
	
	//  Read a bandwidth sized chunk every second (idk if this is how bandwidth caps work)
	for i := 0; i < num_chunks; i++{
		start := time.Now()
		var added int
		if i == num_chunks - 1{
			//  Just read everything left if on last chunk
			added, err = f(b[i * chunk_size:])
		} else {
			//  Read next chunk
			added, err = f(b[i * chunk_size:(i + 1) * chunk_size])
		}
		if err != nil{
			return n, err
		}
		n += added
		//  Sleep until 1 second is reached
		if (i != 0){
			time.Sleep(time.Second - time.Since(start))
		}
	}
	
	//  Simulate second half of latency
	time.Sleep(hlat)
	return n, err	
}

type TestDialer struct{
	*manet.Dialer
}

func NewTestDialer() *TestDialer{
	return &TestDialer{&manet.Dialer{}}
}

//  Dials addrs in the form /ip4/127.0.0.1/tcp/1234/caps/30mbps,50ms
func (td *TestDialer) DialTestAddr(remote string) (CappedConn, error) {
	i := strings.Index(remote, "caps")
	maddr, err := ma.NewMultiaddr(remote[:i - 1])
	if err != nil{
		log.Fatalf("Could not create normal multiaddr: ", err)
	}
	
	conn, err := td.Dial(maddr)
	if err != nil{
		log.Fatalf("Could not dial normal part of address.", err)
	}
	
	caps := strings.Split(remote[i + 5:], ",")
	fmt.Println(caps)
	lat, bw := parseCaps(caps[0], caps[1])
	
	return CappedConn{conn, lat, bw}, nil
}

//  for now assuming bandwidth is mbps
func parseCaps(bwString string, latString string) (float64, time.Duration){
	bw, err := strconv.ParseFloat(bwString[:len(bwString) - 4], 64)
	if err != nil{
		log.Fatal("Invalid bandwidth: ", err)
	}
	
	lat, err := time.ParseDuration(latString)
	if err != nil{
		log.Fatal("Invalid latency: ", err)
	}
	
	return bw, lat
}