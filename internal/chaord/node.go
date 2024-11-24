package chaord

import (
	"Chaord/internal/abvss"
	"Chaord/internal/osv"
	"Chaord/pkg/crypto/commit/merkle"
	"Chaord/pkg/utils/polynomial"
	"bytes"
	"crypto/md5"
	"log"
	"math/big"
	"math/rand"
	"strconv"
)

type Param struct {
	nodeNum     int
	degree      int
	dataScale   int
	sampleScale int
	p           *big.Int
}

func NewParam(nodeNum, degree, dataScale, sampleScale int, p *big.Int) *Param {
	return &Param{
		nodeNum:     nodeNum,
		degree:      degree,
		dataScale:   dataScale,
		sampleScale: sampleScale,
		p:           p,
	}
}

type Node struct {
	*Param
	id              int
	distributeParam int

	bandwidth int

	rChan    chan merkle.Root
	r1Chan   chan merkle.Root
	bHatChan chan []*big.Int
	pHatChan chan []merkle.Witness

	r  merkle.Root
	r1 merkle.Root

	batchCSS *abvss.ABVSS

	random *rand.Rand

	osv0  *osv.Node
	osv1  *osv.Node
	osvTX *osv.Node
}

func NewNode(param *Param, id, dP int, async bool) *Node {
	n := &Node{
		Param:           param,
		id:              id,
		distributeParam: dP,
		random:          rand.New(rand.NewSource(1)),
		bandwidth:       0,
	}
	if async {
		n.rChan = make(chan merkle.Root)
		n.r1Chan = make(chan merkle.Root)
		n.bHatChan = make(chan []*big.Int)
		n.pHatChan = make(chan []merkle.Witness)
	}
	return n
}

func (n *Node) step1(fShares, gShares []*big.Int, r merkle.Root) {
	vss, err := abvss.NewVSS(0, n.id, n.nodeNum, n.degree, n.dataScale, n.nodeNum, n.p, 1)
	if err != nil {
		log.Printf("owner NewVSS err:%v", err)
	}
	vss.ReceiverInit(nil)
	n.batchCSS = vss
	// receive from BACSS.share and generate proof
	n.batchCSS.SetShares(fShares, gShares)
	n.batchCSS.GetLCM()

	// store merkle root
	n.r = r
}

func (n *Node) step2Sample(cShares [][]*big.Int) ([][]*big.Int, []int) {
	flag := make([]int, n.dataScale)
	N := n.sampleScale
	var pi []byte
	bigD := new(big.Int).SetInt64(int64(n.dataScale))
	hasher := md5.New()
	for i := 0; i < N; i++ {
		content := append(pi, []byte(strconv.Itoa(i))...)
		index := new(big.Int).Mod(new(big.Int).SetBytes(hasher.Sum(content)), bigD)
		flag[index.Int64()] = 1
	}
	cHat := make([][]*big.Int, n.dataScale)
	for i := 0; i < n.dataScale; i++ {
		if flag[i] == 1 {
			cHat[i] = cShares[i]
		} else {
			cHat[i] = nil
		}
	}
	return cHat, flag
}

func (n *Node) step2Forward(bHat []*big.Int, pHat []merkle.Witness) {
	hasher := md5.New()
	for i := 0; i < n.dataScale; i++ {
		if bHat[i] != nil {
			_, err := merkle.Verify(n.r, pHat[i], bHat[i].Bytes(), hasher.Sum)
			if err != nil {
				return
			}
		}
	}
}

func (n *Node) step3() {
	n.r1 = <-n.r1Chan
	if bytes.Compare(n.r1, n.r) == 0 {
		n.osv0.Init()
		n.osv1.Init()
	}

}

func (n *Node) step4() {
	for {
		select {
		case <-n.osv0.OutPut:
			// refund P to consumer Pc
		case <-n.osv1.OutPut:
			// initiate BDDG
		case <-n.osvTX.OutPut:
			// transfer P to owner Po
		}
	}

}

func (n *Node) coin() {
	coinPre := make([]int, n.distributeParam)
	for i := 0; i < n.distributeParam; i++ {
		coinPre[i] = n.random.Int() % 2
	}
}

func (n *Node) odoLocal() {
	// generate shares of s and s^2
	for i := 0; i < n.dataScale; i++ {
		poly, _ := polynomial.New(n.degree)
		_ = poly.SetCoefficient(0, 0)
	}
	RShares := make([]*big.Int, n.dataScale)
	bBias := make([]int, n.dataScale)
	boShares := make([]*big.Int, n.dataScale)
	for i := 0; i < n.dataScale; i++ {
		bBias[i] = rand.Int() % 2
	}
	boAddRShares := make([]*big.Int, n.dataScale)
	for i := 0; i < n.dataScale; i++ {
		boAddRShares[i] = new(big.Int).Add(boShares[i], RShares[i])

	}
}

// Run start node
func (n *Node) Run() {

}
