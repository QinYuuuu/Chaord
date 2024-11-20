package chaord

import (
	"Chaord/internal/reedsolomonP"
	"Chaord/pkg/crypto/commit/merkle"
	"math/rand"

	"hash"
	"math/big"
)

type Consumer struct {
	nodeNum     int
	degree      int
	dataScale   int
	sampleScale int
	b           []*big.Int
	bfChan      chan BFMsg
	cHatChan    chan CHatMsg
	hasher      hash.Hash

	rsCode *reedsolomonP.RSGFp
}

func NewConsumer(n, dataScale int, sampleScale int) *Consumer {
	return &Consumer{
		nodeNum:     n,
		dataScale:   dataScale,
		sampleScale: sampleScale,
		cHatChan:    make(chan CHatMsg, n),
	}
}

func (c *Consumer) foretaste() {
	//	receive (<cHat>_i, bHat) then
	//	m = OEC([<cHat>_i]) − bHat
	cHat := make([]*big.Int, c.dataScale)
	mHat := make([]*big.Int, c.dataScale)
	bHat := make([]*big.Int, c.dataScale)

	for {
		done := true
		cHatShareList := make([][]reedsolomonP.Share, c.dataScale)
		cHatMsg := <-c.cHatChan
		index := cHatMsg.FromID
		for i := 0; i < c.dataScale; i++ {
			cHatShare := reedsolomonP.Share{
				Number: index,
				Data:   new(big.Int).SetBytes(cHatMsg.CHat[i]),
			}
			cHatShareList[i] = append(cHatShareList[i], cHatShare)
		}
		for i := 0; i < c.dataScale; i++ {
			correct, err := c.rsCode.Correct(cHatShareList[i])
			if err != nil {
				done = false
				break
			}
			cHatCoef, err := c.rsCode.Decode(correct)
			if err != nil {
				done = false
				break
			}
			cHat[i] = cHatCoef[0]
		}
		if done == true {
			break
		}
	}
	for i := 0; i < c.dataScale; i++ {
		mHat[i] = new(big.Int).Sub(cHat[i], bHat[i])
	}

	// verify mHat meets privacy price
	/*
	 */
}

func (c *Consumer) step3() {
	// receiver blinding factor
	b := <-c.bfChan
	if len(b.B) == c.dataScale {
		tree, err := merkle.NewMerkleTree(b.B, c.hasher.Sum)
		if err != nil {
			return
		}
		merkle.Commit(tree)
	}
}

func (c *Consumer) batchDDGInit() [][]*big.Int {
	// share secrets on Z2
	bcShares := make([][]*big.Int, c.dataScale)
	for i := 0; i < c.dataScale; i++ {
		s := rand.Int() % 2
		bcShares[i] = shareSecret(s, c.nodeNum, c.degree, new(big.Int).SetInt64(2))
	}
	return bcShares
}
