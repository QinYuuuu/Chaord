package chaord

import (
	"Chaord/internal/abvss"
	"Chaord/pkg/crypto/commit/merkle"
	"math/big"
	"math/rand"
)

type Owner struct {
	nodeNum     int
	degree      int
	dataScale   int
	sampleScale int

	data        []*big.Int
	p           *big.Int
	b           []*big.Int
	merkleProof []merkle.Witness

	// in BatchDDG
	batchCSS *abvss.ABVSS

	// async
	flagChan chan FlagMsg
	bfChan   chan BFMsg
}

func NewOwner(nodeNum, degree, dataScale, sampleScale int, async bool) *Owner {
	owner := &Owner{
		nodeNum:     nodeNum,
		dataScale:   dataScale,
		sampleScale: sampleScale,

		data:        make([]*big.Int, dataScale),
		b:           make([]*big.Int, dataScale),
		merkleProof: make([]merkle.Witness, dataScale),
	}
	if async {
		owner.flagChan = make(chan FlagMsg, nodeNum)
		owner.bfChan = make(chan BFMsg)
	}
	return owner
}

func (owner *Owner) step1() {
	vec_b := make([]int, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		vec_b[i] = rand.Int()
	}
}

func (owner *Owner) step2() {
	flag := <-owner.flagChan
	bHat := make([]*big.Int, owner.dataScale)
	pHat := make([]merkle.Witness, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		if flag.Flag[i] {
			bHat[i] = owner.b[i]
			pHat[i] = owner.merkleProof[i]
		} else {
			bHat[i] = new(big.Int).SetInt64(0)
			pHat[i] = merkle.Witness{}
		}
	}
	// RBC bHat and pHat

	// clear flagChan
	for {
		<-owner.flagChan
	}
}

func (owner *Owner) step3() {
	b := make([][]byte, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		b[i] = owner.b[i].Bytes()
	}
}

func (owner *Owner) batchDDGInit() [][]*big.Int {
	// share secrets on Zp
	boShares := make([][]*big.Int, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		s := rand.Int() % 2
		boShares[i] = shareSecret(s, owner.nodeNum, owner.degree, owner.p)
	}
	return boShares
}
