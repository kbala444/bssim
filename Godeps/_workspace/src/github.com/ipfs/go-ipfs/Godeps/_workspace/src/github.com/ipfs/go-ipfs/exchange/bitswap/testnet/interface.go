package bitswap

import (
	bsnet "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/exchange/bitswap/network"
	peer "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/p2p/peer"
	"github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/util/testutil"
)

type Network interface {
	Adapter(testutil.Identity) bsnet.BitSwapNetwork

	HasPeer(peer.ID) bool
}
