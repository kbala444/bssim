package coreunix

import (
	"io"

	core "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/core"
	path "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/path"
	uio "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/unixfs/io"
)

func Cat(n *core.IpfsNode, pstr string) (io.Reader, error) {
	p := path.FromString(pstr)
	dagNode, err := n.Resolver.ResolvePath(n.Context(), p)
	if err != nil {
		return nil, err
	}
	return uio.NewDagReader(n.Context(), dagNode, n.DAG)
}
