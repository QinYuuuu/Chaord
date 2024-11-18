package chaord

import (
	"Chaord/internal/abvss"
	"Chaord/pkg/crypto/commit/merkle"
	"math/big"
	"math/rand"
)

type Owner struct {
	n           int
	dataScale   int
	sampleScale int
	flagChan    chan FlagMsg
	bfChan      chan BFMsg
	data        []*big.Int
	b           []*big.Int
	p           []merkle.Witness

	// in BatchDDG
	batchCSS *abvss.ABVSS
}

func NewOwner(n, dataScale, sampleScale int) *Owner {
	return &Owner{
		n:           n,
		dataScale:   dataScale,
		sampleScale: sampleScale,
		flagChan:    make(chan FlagMsg, n),
		bfChan:      make(chan BFMsg),
		data:        make([]*big.Int, dataScale),
		b:           make([]*big.Int, dataScale),
		p:           make([]merkle.Witness, dataScale),
	}
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
			pHat[i] = owner.p[i]
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

func (owner *Owner) batchDDG() {
	s := make([]*big.Int, owner.dataScale)
	// share s
	err := owner.batchCSS.DistributorInit(nil, s)
	if err != nil {
		return
	}
	for i := 0; i < owner.n; i++ {
		fshares, _, err := owner.batchCSS.GenerateRawShares(i)
		if err != nil {
			return
		}
	}
}
