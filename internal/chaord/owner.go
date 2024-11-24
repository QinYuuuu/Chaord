package chaord

import (
	"Chaord/internal/abvss"
	"Chaord/pkg/crypto/commit/merkle"
	"Chaord/pkg/utils"
	"crypto/md5"
	"errors"
	"log"
	"math/big"
	"math/rand"
)

type Owner struct {
	*Param

	data        []*big.Int
	b           []*big.Int
	merkleProof []merkle.Witness

	bandwidth int

	// in BatchDDG
	batchCSS *abvss.ABVSS

	// async
	flagChan chan FlagMsg
	bfChan   chan BFMsg
}

func NewOwner(param *Param, async bool) *Owner {
	owner := &Owner{
		Param:       param,
		bandwidth:   0,
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

func (owner *Owner) SetData(data []*big.Int) error {
	if len(data) != owner.dataScale {
		return errors.New("data scale error")
	}
	owner.data = data
	vss, err := abvss.NewVSS(0, 0, owner.nodeNum, owner.degree, owner.dataScale, owner.nodeNum, owner.p, 1)
	if err != nil {
		log.Printf("owner NewVSS err:%v", err)
	}
	owner.batchCSS = vss
	return nil
}

func (owner *Owner) step1() ([][]*big.Int, [][]*big.Int, merkle.Root) {
	bByte := make([][]byte, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		owner.b[i] = utils.RandomNum(owner.p)
		bByte[i] = owner.b[i].Bytes()
	}
	hasher := md5.New()
	tree, err := merkle.NewMerkleTree(bByte, hasher.Sum)
	if err != nil {
		return nil, nil, nil
	}
	root := merkle.Commit(tree)
	for i := 0; i < owner.dataScale; i++ {
		owner.merkleProof[i], _ = merkle.CreateWitness(tree, i)
	}
	c := make([]*big.Int, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		c[i] = new(big.Int).Add(owner.data[i], owner.b[i])
	}

	cShares := make([][]*big.Int, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		cShares[i] = make([]*big.Int, owner.nodeNum)
		//cShares[i] = shareSecretBig(c[i], owner.nodeNum, owner.degree, owner.p)
	}
	err = owner.batchCSS.DistributorInit(nil, c)
	if err != nil {
		log.Printf("owner batchCSS init err:%v", err)
		return nil, nil, nil
	}
	err = owner.batchCSS.SamplePoly()
	if err != nil {
		log.Printf("owner batchCSS sample poly err:%v", err)
		return nil, nil, nil
	}
	vShares := make([][]*big.Int, owner.nodeNum)
	for i := 0; i < owner.nodeNum; i++ {
		vShares[i] = make([]*big.Int, owner.nodeNum)
		//cShares[i] = shareSecretBig(c[i], owner.nodeNum, owner.degree, owner.p)
	}
	for i := 0; i < owner.nodeNum; i++ {
		tmp, tmp2, _ := owner.batchCSS.GenerateRawShares(i)
		for j := 0; j < owner.dataScale; j++ {
			cShares[j][i] = tmp[j]
		}
		for j := 0; j < owner.nodeNum; j++ {
			vShares[j][i] = tmp2[j]
		}
	}

	owner.bandwidth += owner.nodeNum * owner.dataScale * len(cShares[0][0].Bytes())
	owner.bandwidth += len(root)
	return cShares, vShares, root
}

func (owner *Owner) step2(flag []int) ([]*big.Int, []merkle.Witness) {
	bHat := make([]*big.Int, owner.dataScale)
	pHat := make([]merkle.Witness, owner.dataScale)
	for i := 0; i < owner.dataScale; i++ {
		if flag[i] == 1 {
			bHat[i] = owner.b[i]
			pHat[i] = owner.merkleProof[i]
		} else {
			bHat[i] = big.NewInt(0)
			pHat[i] = merkle.Witness{}
		}
	}
	owner.bandwidth += owner.dataScale * len(bHat[0].Bytes())
	owner.bandwidth += owner.dataScale * len(pHat[0].Hash())
	return bHat, pHat
}

func (owner *Owner) step3() []*big.Int {

	owner.bandwidth += owner.dataScale * len(owner.b[0].Bytes())
	return owner.b
}

func (owner *Owner) batchDDGInit() (*big.Int, []*big.Int) {
	// share secret on Zp
	s := rand.Int() % 2
	boShares := shareSecret(s, owner.nodeNum, owner.degree, owner.p)
	owner.bandwidth += owner.nodeNum
	return new(big.Int).SetInt64(int64(s)), boShares
}
