package commands

import (
	"bytes"
	"fmt"
	"io"

	key "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/blocks/key"
	cmds "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/commands"
	corerepo "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/core/corerepo"
	u "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/util"
)

var PinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Pin (and unpin) objects to local storage",
	},

	Subcommands: map[string]*cmds.Command{
		"add": addPinCmd,
		"rm":  rmPinCmd,
		"ls":  listPinCmd,
	},
}

type PinOutput struct {
	Pinned []key.Key
}

var addPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Pins objects to local storage",
		ShortDescription: `
Retrieves the object named by <ipfs-path> and stores it locally
on disk.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "Path to object(s) to be pinned").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.BoolOption("recursive", "r", "Recursively pin the object linked to by the specified object(s)"),
	},
	Type: PinOutput{},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.Context().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		// set recursive flag
		recursive, found, err := req.Option("recursive").Bool()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
		if !found {
			recursive = false
		}

		added, err := corerepo.Pin(n, req.Arguments(), recursive)
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		res.SetOutput(&PinOutput{added})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			added, ok := res.Output().(*PinOutput)
			if !ok {
				return nil, u.ErrCast()
			}

			var pintype string
			rec, _, _ := res.Request().Option("recursive").Bool()
			if rec {
				pintype = "recursively"
			} else {
				pintype = "directly"
			}

			buf := new(bytes.Buffer)
			for _, k := range added.Pinned {
				fmt.Fprintf(buf, "pinned %s %s\n", k, pintype)
			}
			return buf, nil
		},
	},
}

var rmPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Unpin an object from local storage",
		ShortDescription: `
Removes the pin from the given object allowing it to be garbage
collected if needed.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "Path to object(s) to be unpinned").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.BoolOption("recursive", "r", "Recursively unpin the object linked to by the specified object(s)"),
	},
	Type: PinOutput{},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.Context().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		// set recursive flag
		recursive, found, err := req.Option("recursive").Bool()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
		if !found {
			recursive = false // default
		}

		removed, err := corerepo.Unpin(n, req.Arguments(), recursive)
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		res.SetOutput(&PinOutput{removed})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			added, ok := res.Output().(*PinOutput)
			if !ok {
				return nil, u.ErrCast()
			}

			buf := new(bytes.Buffer)
			for _, k := range added.Pinned {
				fmt.Fprintf(buf, "unpinned %s\n", k)
			}
			return buf, nil
		},
	},
}

var listPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List objects pinned to local storage",
		ShortDescription: `
Returns a list of hashes of objects being pinned. Objects that are indirectly
or recursively pinned are not included in the list.
`,
		LongDescription: `
Returns a list of hashes of objects being pinned. Objects that are indirectly
or recursively pinned are not included in the list.

Use --type=<type> to specify the type of pinned keys to list. Valid values are:
    * "direct": pin that specific object.
    * "recursive": pin that specific object, and indirectly pin all its decendants
    * "indirect": pinned indirectly by an ancestor (like a refcount)
    * "all"

To see the ref count on indirect pins, pass the -count option flag.
Defaults to "direct".
`,
	},

	Options: []cmds.Option{
		cmds.StringOption("type", "t", "The type of pinned keys to list. Can be \"direct\", \"indirect\", \"recursive\", or \"all\". Defaults to \"direct\""),
		cmds.BoolOption("count", "n", "Show refcount when listing indirect pins"),
		cmds.BoolOption("quiet", "q", "Write just hashes of objects"),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		n, err := req.Context().GetNode()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}

		typeStr, found, err := req.Option("type").String()
		if err != nil {
			res.SetError(err, cmds.ErrNormal)
			return
		}
		if !found {
			typeStr = "direct"
		}

		switch typeStr {
		case "all", "direct", "indirect", "recursive":
		default:
			err = fmt.Errorf("Invalid type '%s', must be one of {direct, indirect, recursive, all}", typeStr)
			res.SetError(err, cmds.ErrClient)
		}

		keys := make(map[string]RefKeyObject)
		if typeStr == "direct" || typeStr == "all" {
			for _, k := range n.Pinning.DirectKeys() {
				keys[k.B58String()] = RefKeyObject{
					Type:  "direct",
					Count: 1,
				}
			}
		}
		if typeStr == "indirect" || typeStr == "all" {
			for k, v := range n.Pinning.IndirectKeys() {
				keys[k.B58String()] = RefKeyObject{
					Type:  "indirect",
					Count: v,
				}
			}
		}
		if typeStr == "recursive" || typeStr == "all" {
			for _, k := range n.Pinning.RecursiveKeys() {
				keys[k.B58String()] = RefKeyObject{
					Type:  "recursive",
					Count: 1,
				}
			}
		}

		res.SetOutput(&RefKeyList{Keys: keys})
	},
	Type: RefKeyList{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			typeStr, _, err := res.Request().Option("type").String()
			if err != nil {
				return nil, err
			}

			count, _, err := res.Request().Option("count").Bool()
			if err != nil {
				return nil, err
			}

			quiet, _, err := res.Request().Option("quiet").Bool()
			if err != nil {
				return nil, err
			}

			keys, ok := res.Output().(*RefKeyList)
			if !ok {
				return nil, u.ErrCast()
			}
			out := new(bytes.Buffer)
			if typeStr == "indirect" && count {
				for k, v := range keys.Keys {
					if quiet {
						fmt.Fprintf(out, "%s %d\n", k, v.Count)
					} else {
						fmt.Fprintf(out, "%s %s %d\n", k, v.Type, v.Count)
					}
				}
			} else {
				for k, v := range keys.Keys {
					if quiet {
						fmt.Fprintf(out, "%s\n", k)
					} else {
						fmt.Fprintf(out, "%s %s\n", k, v.Type)
					}
				}
			}
			return out, nil
		},
	},
}

type RefKeyObject struct {
	Type  string
	Count int
}

type RefKeyList struct {
	Keys map[string]RefKeyObject
}
