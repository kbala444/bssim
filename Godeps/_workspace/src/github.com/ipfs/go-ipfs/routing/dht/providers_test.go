package dht

import (
	"testing"

	key "github.com/heems/bssim/Godeps/_workspace/src/github.com/ipfs/go-ipfs/blocks/key"
	peer "github.com/heems/bssim/Godeps/_workspace/src/github.com/ipfs/go-ipfs/p2p/peer"

	context "github.com/heems/bssim/Godeps/_workspace/src/golang.org/x/net/context"
)

func TestProviderManager(t *testing.T) {
	ctx := context.Background()
	mid := peer.ID("testing")
	p := NewProviderManager(ctx, mid)
	a := key.Key("test")
	p.AddProvider(ctx, a, peer.ID("testingprovider"))
	resp := p.GetProviders(ctx, a)
	if len(resp) != 1 {
		t.Fatal("Could not retrieve provider.")
	}
	p.proc.Close()
}
