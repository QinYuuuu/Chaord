package chaord

import (
	"Chaord/internal/abvss"
	"Chaord/pkg/crypto/commit/merkle"
	"math/big"
	"math/rand"
)

type Owner struct {
	*Param

	data        []*big.Int
	b           []*big.Int
	merkleProof []merkle.Witness

	// in BatchDDG
	batchCSS *abvss.ABVSS

	// async
	flagChan chan FlagMsg
	bfChan   chan BFMsg
}

func NewOwner(param *Param, async bool) *Owner {
	owner := &Owner{
		Param: param,

		data:        make([]*big.Int, param.dataScale),
		b:           make([]*big.Int, param.dataScale),
		merkleProof: make([]merkle.Witness, param.dataScale),
	}

	if async {
		owner.flagChan = make(chan FlagMsg, param.nodeNum)
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

func (owner *Owner) batchDDGInit() (*big.Int, []*big.Int) {
	// share secret on Zp
	s := rand.Int() % 2
	boShares := shareSecret(s, owner.nodeNum, owner.degree, owner.p)
	return new(big.Int).SetInt64(int64(s)), boShares
}
