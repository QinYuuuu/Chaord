package chaord

import (
	"Chaord/internal/reedsolomonP"
	"Chaord/pkg/crypto/commit/merkle"
	"crypto/md5"
	"hash"
	"math/big"
	"math/rand"
	"sync"
)

type Consumer struct {
	*Param

	b             []*big.Int
	dataDisturbed []*big.Int

	bandwidth int

	bfChan   chan BFMsg
	cHatChan chan CHatMsg
	hasher   hash.Hash

	rsCode *reedsolomonP.RSGFp
}

func NewConsumer(param *Param, async bool) *Consumer {
	consumer := &Consumer{
		Param:     param,
		hasher:    md5.New(),
		bandwidth: 0,
	}
	if async {
		consumer.cHatChan = make(chan CHatMsg, param.nodeNum)
	}
	return consumer
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

// use big.Int or byte
func (c *Consumer) step2(index []*big.Int, cHatShares [][]*big.Int, bHat []*big.Int) {
	if len(cHatShares) != c.dataScale {
		return
	}
	cHat := make([]*big.Int, c.dataScale)
	for i := 0; i < c.dataScale; i++ {
		if cHatShares[i] == nil {
			cHat[i] = big.NewInt(0)
			continue
		} else {
			cHat[i] = reconstruct(index, cHatShares[i], c.p)
		}

	}
	mHat := make([]*big.Int, c.dataScale)
	for i := 0; i < c.dataScale; i++ {
		mHat[i] = new(big.Int).Sub(cHat[i], bHat[i])
	}
}

func (c *Consumer) step3(b []*big.Int) merkle.Root {
	// receiver blinding factor
	c.b = b
	bByte := make([][]byte, c.dataScale)

	for i := 0; i < c.dataScale; i++ {
		bByte[i] = b[i].Bytes()
	}
	if len(b) != c.dataScale {
		return nil
	}
	tree, err := merkle.NewMerkleTree(bByte, c.hasher.Sum)
	if err != nil {
		return nil
	}
	r := merkle.Commit(tree)
	return r
}

func (c *Consumer) step4(index []*big.Int, cBarShares [][]*big.Int) {
	mBar := make([]*big.Int, c.dataScale)
	var wg sync.WaitGroup
	wg.Add(c.dataScale)
	for i := 0; i < c.dataScale; i++ {
		go func(i int) {
			tmp := reconstruct(index, cBarShares[i], c.p)
			mBar[i] = new(big.Int).Sub(tmp, c.b[i])
			mBar[i].Mod(mBar[i], c.p)
			wg.Done()
		}(i)
	}
	wg.Wait()
	c.dataDisturbed = mBar
}

func (c *Consumer) batchDDGInit() (*big.Int, []*big.Int) {
	// share secret on Zp
	bcShares := make([]*big.Int, c.dataScale)

	s := rand.Int() % 2
	bcShares = shareSecret(s, c.nodeNum, c.degree, c.p)

	return new(big.Int).SetInt64(int64(s)), bcShares
}
