package chaord

import (
	"fmt"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"math/big"
)

func LocalTest(nodeNum, degree, dataScale, sampleScale int, p *big.Int) {
	//owner := NewOwner(nodeNum, degree, dataScale, sampleScale, false)
	curve := edwards25519.NewBlakeSHA256Ed25519()
	fmt.Print(curve.ScalarLen())
}
