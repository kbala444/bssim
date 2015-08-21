// +build linux darwin

package utp

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ThomasRooney/gexpect"
	"github.com/h2so5/utp"
)

func TestUcatListen(t *testing.T) {
	child, err := gexpect.Spawn("libutp/ucat-static -l -p 8000")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	addr, err := utp.ResolveAddr("utp", "127.0.0.1:8000")
	if err != nil {
		t.Fatal(err)
	}
	c, err := utp.DialUTPTimeout("utp", nil, addr, 1000*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	var payload [128]byte
	_, err = rand.Read(payload[:])
	if err != nil {
		t.Fatal(err)
	}

	msg := hex.EncodeToString(payload[:])
	_, err = c.Write([]byte(msg + "\n"))
	if err != nil {
		t.Fatal(err)
	}

	err = child.ExpectTimeout(msg, 1000*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	child.SendLine(msg + "\n")

	err = c.SetDeadline(time.Now().Add(1000 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	var buf [1024]byte
	l, err := c.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}

	if string(buf[:l]) != msg+"\n" {
		t.Errorf("expected payload of %s; got %s", msg, string(buf[:l]))
	}

	c.Close()
	child.Wait()
}

func TestUcatConnect(t *testing.T) {
	addr, err := utp.ResolveAddr("utp", "127.0.0.1:9000")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := utp.Listen("utp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	err = ln.SetDeadline(time.Now().Add(1000 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	child, err := gexpect.Spawn("libutp/ucat-static 127.0.0.1 9000")
	if err != nil {
		t.Fatal(err)
	}

	c, err := ln.AcceptUTP()
	if err != nil {
		t.Fatal(err)
	}

	var payload [128]byte
	_, err = rand.Read(payload[:])
	if err != nil {
		t.Fatal(err)
	}

	msg := hex.EncodeToString(payload[:])
	_, err = c.Write([]byte(msg + "\n"))
	if err != nil {
		t.Fatal(err)
	}

	err = child.ExpectTimeout(msg, 1000*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	child.SendLine(msg + "\n")
	err = c.SetDeadline(time.Now().Add(1000 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	var buf [1024]byte
	l, err := c.Read(buf[:])
	if err != nil {
		t.Fatal(err)
	}

	if string(buf[:l]) != msg+"\n" {
		t.Errorf("expected payload of %s; got %s", msg, string(buf[:l]))
	}

	c.Close()
	child.Wait()
}
