package reprovide_test

import (
	"testing"

	blocks "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/blocks"
	blockstore "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/blocks/blockstore"
	mock "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/routing/mock"
	testutil "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/util/testutil"
	ds "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	dssync "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	context "github.com/heems/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"

	. "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/exchange/reprovide"
)

func TestReprovide(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mrserv := mock.NewServer()

	idA := testutil.RandIdentityOrFatal(t)
	idB := testutil.RandIdentityOrFatal(t)

	clA := mrserv.Client(idA)
	clB := mrserv.Client(idB)

	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))

	blk := blocks.NewBlock([]byte("this is a test"))
	bstore.Put(blk)

	reprov := NewReprovider(clA, bstore)
	err := reprov.Reprovide(ctx)
	if err != nil {
		t.Fatal(err)
	}

	provs, err := clB.FindProviders(ctx, blk.Key())
	if err != nil {
		t.Fatal(err)
	}

	if len(provs) == 0 {
		t.Fatal("Should have gotten a provider")
	}

	if provs[0].ID != idA.ID() {
		t.Fatal("Somehow got the wrong peer back as a provider.")
	}
}
