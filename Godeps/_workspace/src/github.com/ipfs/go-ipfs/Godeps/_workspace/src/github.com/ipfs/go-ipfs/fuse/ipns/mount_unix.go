// +build linux darwin freebsd
// +build !nofuse

package ipns

import (
	core "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/core"
	mount "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/fuse/mount"
)

// Mount mounts ipns at a given location, and returns a mount.Mount instance.
func Mount(ipfs *core.IpfsNode, ipnsmp, ipfsmp string) (mount.Mount, error) {
	cfg := ipfs.Repo.Config()
	allow_other := cfg.Mounts.FuseAllowOther

	fsys, err := NewFileSystem(ipfs, ipfs.PrivateKey, ipfsmp, ipnsmp)
	if err != nil {
		return nil, err
	}

	return mount.NewMount(ipfs.Process(), fsys, ipnsmp, allow_other)
}
