package record

import (
	"bytes"
	"errors"
	"strings"

	key "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/blocks/key"
	ci "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/p2p/crypto"
	pb "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/routing/dht/pb"
	u "github.com/heems/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-ipfs/util"
)

// ValidatorFunc is a function that is called to validate a given
// type of DHTRecord.
type ValidatorFunc func(key.Key, []byte) error

// ErrBadRecord is returned any time a dht record is found to be
// incorrectly formatted or signed.
var ErrBadRecord = errors.New("bad dht record")

// ErrInvalidRecordType is returned if a DHTRecord keys prefix
// is not found in the Validator map of the DHT.
var ErrInvalidRecordType = errors.New("invalid record keytype")

// Validator is an object that helps ensure routing records are valid.
// It is a collection of validator functions, each of which implements
// its own notion of validity.
type Validator map[string]*ValidChecker

type ValidChecker struct {
	Func ValidatorFunc
	Sign bool
}

// VerifyRecord checks a record and ensures it is still valid.
// It runs needed validators
func (v Validator) VerifyRecord(r *pb.Record) error {
	// Now, check validity func
	parts := strings.Split(r.GetKey(), "/")
	if len(parts) < 3 {
		log.Infof("Record key does not have validator: %s", key.Key(r.GetKey()))
		return nil
	}

	val, ok := v[parts[1]]
	if !ok {
		log.Infof("Unrecognized key prefix: %s", parts[1])
		return ErrInvalidRecordType
	}

	return val.Func(key.Key(r.GetKey()), r.GetValue())
}

func (v Validator) IsSigned(k key.Key) (bool, error) {
	// Now, check validity func
	parts := strings.Split(string(k), "/")
	if len(parts) < 3 {
		log.Infof("Record key does not have validator: %s", k)
		return false, nil
	}

	val, ok := v[parts[1]]
	if !ok {
		log.Infof("Unrecognized key prefix: %s", parts[1])
		return false, ErrInvalidRecordType
	}

	return val.Sign, nil
}

// ValidatePublicKeyRecord implements ValidatorFunc and
// verifies that the passed in record value is the PublicKey
// that matches the passed in key.
func ValidatePublicKeyRecord(k key.Key, val []byte) error {
	keyparts := bytes.Split([]byte(k), []byte("/"))
	if len(keyparts) < 3 {
		return errors.New("invalid key")
	}

	pkh := u.Hash(val)
	if !bytes.Equal(keyparts[2], pkh) {
		return errors.New("public key does not match storage key")
	}
	return nil
}

var PublicKeyValidator = &ValidChecker{
	Func: ValidatePublicKeyRecord,
	Sign: false,
}

func CheckRecordSig(r *pb.Record, pk ci.PubKey) error {
	blob := RecordBlobForSig(r)
	good, err := pk.Verify(blob, r.Signature)
	if err != nil {
		return nil
	}
	if !good {
		return errors.New("invalid record signature")
	}
	return nil
}
