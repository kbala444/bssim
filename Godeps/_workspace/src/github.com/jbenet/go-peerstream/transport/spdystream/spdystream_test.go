package peerstream_spdystream

import (
	"testing"

	psttest "github.com/heems/bssim/Godeps/_workspace/src/github.com/jbenet/go-peerstream/transport/test"
)

func TestSpdyStreamTransport(t *testing.T) {
	psttest.SubtestAll(t, Transport)
}
