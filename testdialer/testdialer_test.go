// testdialer_test
package testdialer

import (
	"strconv"
	//"fmt"
	"net"
	"testing"
	"time"
	//ma "github.com/jbenet/go-multiaddr-net/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	"bytes"
	"sync"
)

func makeListener(port int, t *testing.T) net.Listener {
	addr := "127.0.0.1:" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal("failed to listen", err)
	}
	return listener
}

func TestDial(t *testing.T) {
	d := NewTestDialer()
	listener := makeListener(1234, t)

	c, err := d.DialTestAddr("/ip4/127.0.0.1/tcp/1234/caps/30mbps,50ms")
	if err != nil {
		t.Fatal("Failed to dial ", err)
	}
	if c.Bw != 30 {
		t.Fatal("Bandwidth not parsed correctly.")
	}
	if c.Lat != time.Millisecond*50 {
		t.Fatal("Latency not parsed correctly.")
	}
	listener.Close()
}

func TestWrite(t *testing.T) {
	d := NewTestDialer()
	listener := makeListener(1111, t)

	cA, _ := d.DialTestAddr("/ip4/127.0.0.1/tcp/1111/caps/50mbps,0s")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cB, err := listener.Accept()
		if err != nil {
			t.Fatal("failed to accept")
		}

		buf := make([]byte, 1024)
		for {
			_, err := cB.Read(buf)
			if err != nil {
				break
			}
			if !bytes.Equal(buf[:4], []byte("ping")) {
				t.Fatal("Incorrect message received.", string(buf))
			}
			cB.Write([]byte("pong"))
		}
		wg.Done()
	}()

	if _, err := cA.Write([]byte("ping")); err != nil {
		t.Fatal("failed to write:", err)
	}
	cA.Close()
	wg.Wait()
	listener.Close()
}

func TestLatency(t *testing.T) {
	d := NewTestDialer()
	listener := makeListener(1111, t)

	cA, _ := d.DialTestAddr("/ip4/127.0.0.1/tcp/1111/caps/50mbps,.5s")
	var rn time.Time
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cB, err := listener.Accept()
		if err != nil {
			t.Fatal("failed to accept")
		}

		buf := make([]byte, 1024)
		for {
			_, err := cB.Read(buf)
			if err != nil {
				break
			}
			if int(time.Since(rn).Seconds()*1000) != 250 {
				t.Fatal("Latency didn't work.  Time: ", time.Since(rn))
			}
			if !bytes.Equal(buf[:4], []byte("ping")) {
				t.Fatal("Incorrect message received.", string(buf))
			}
			cB.Write([]byte("pong"))
		}

		wg.Done()
	}()

	buf := make([]byte, 1024)
	rn = time.Now()
	//  1s
	if _, err := cA.Write([]byte("ping")); err != nil {
		t.Fatal("failed to write:", err)
	}
	if int(time.Since(rn).Seconds()*1000) != 500 {
		t.Fatal("Latency didn't work.  Time: ", time.Since(rn))
	}

	//  1s
	if _, err := cA.Read(buf); err != nil {
		t.Fatal("failed to read:", buf, err)
	}

	if !bytes.Equal(buf[:4], []byte("pong")) {
		t.Fatal("Incorrect message received.  Message: ", string(buf[:4]))
	}

	if int(time.Since(rn).Seconds()) != 1 {
		t.Fatal("Latency didn't work.  Time: ", time.Since(rn))
	}
	cA.Close()
	wg.Wait()
	listener.Close()
}

func TestBandwidth(t *testing.T) {
	d := NewTestDialer()
	listener := makeListener(1234, t)

	cA, _ := d.DialTestAddr("/ip4/127.0.0.1/tcp/1234/caps/.5mbps,0s")
	var rn time.Time
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cB, err := listener.Accept()
		if err != nil {
			t.Fatal("failed to accept")
		}

		// echo out
		buf := make([]byte, 1024*1024)
		for {
			_, err := cB.Read(buf)
			if err != nil {
				break
			}
		}
		wg.Done()
	}()

	//  1mb message at .5mbps should take 2 seconds
	buf := make([]byte, 1024*1024)
	buf = append([]byte("beep boop"), buf...)
	rn = time.Now()

	if _, err := cA.Write(buf); err != nil {
		t.Fatal("failed to write:", err)
	}

	if int(time.Since(rn).Seconds()) != 2 {
		t.Fatal("Bandwidth didn't work.  Time: ", time.Since(rn))
	}

	cA.Close()
	wg.Wait()
	listener.Close()
}
