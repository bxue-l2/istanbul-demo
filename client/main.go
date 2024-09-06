package main

import (
	"encoding/binary"
	"fmt"
	"os"

	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	program_client "github.com/ethereum-optimism/optimism/op-program/client"
)

type rawHint string

func (rh rawHint) Hint() string {
	return string(rh)
}

func main() {
	_, _ = os.Stderr.Write([]byte("started!"))

	po := preimage.NewOracleClient(program_client.CreatePreimageChannel())
	hinter := preimage.NewHintWriter(program_client.CreateHinterChannel())
	preHash := *(*[32]byte)(po.Get(preimage.LocalIndexKey(0)))
	diffHash := *(*[32]byte)(po.Get(preimage.LocalIndexKey(1)))
	claimData := *(*[8]byte)(po.Get(preimage.LocalIndexKey(2)))

	fmt.Printf("preHash %x\n", preHash)
	fmt.Printf("diffHash %x\n", diffHash)
	fmt.Printf("claimData %x\n", claimData)

	// Hints are used to indicate which things the program will access,
	// so the server can be prepared to serve the corresponding pre-images.
	hinter.Hint(rawHint(fmt.Sprintf("fetch-state %x", preHash)))
	pre := po.Get(preimage.Keccak256Key(preHash))

	// Multiple pre-images may be fetched based on a hint.
	// E.g. when we need all values of a merkle-tree.
	hinter.Hint(rawHint(fmt.Sprintf("fetch-diff %x", diffHash)))
	diff := po.Get(preimage.Keccak256Key(diffHash))
	diffPartA := po.Get(preimage.Keccak256Key(*(*[32]byte)(diff[:32])))
	diffPartB := po.Get(preimage.Keccak256Key(*(*[32]byte)(diff[32:])))

	// Example state-transition function: s' = s*a + b
	s := binary.BigEndian.Uint64(pre)
	a := binary.BigEndian.Uint64(diffPartA)
	b := binary.BigEndian.Uint64(diffPartB)
	fmt.Printf("computing %d * %d + %d\n", s, a, b)
	sOut := s*a + b

	sClaim := binary.BigEndian.Uint64(claimData[:])
	if sOut != sClaim {
		fmt.Printf("claim %d is bad! Correct result is %d\n", sOut, sClaim)
		os.Exit(1)
	} else {
		fmt.Printf("claim %d is good!\n", sOut)
		os.Exit(0)
	}
}
