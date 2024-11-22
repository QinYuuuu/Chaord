package chaord

import (
	"Chaord/internal/osv"
	"Chaord/pkg/crypto/commit/merkle"
	"Chaord/pkg/utils/polynomial"
	"bytes"
	"crypto/md5"
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
	distributeParam int
	percentage      float64

	rChan    chan merkle.Root
	r1Chan   chan merkle.Root
	bHatChan chan []*big.Int
	pHatChan chan []merkle.Witness

	r  merkle.Root
	r1 merkle.Root

	random *rand.Rand

	osv0  *osv.Node
	osv1  *osv.Node
	osvTX *osv.Node
}

func NewNode(param *Param, dP int, async bool) *Node {
	n := &Node{
		Param:           param,
		distributeParam: dP,
		random:          rand.New(rand.NewSource(1)),
	}
	if async {
		n.rChan = make(chan merkle.Root)
		n.r1Chan = make(chan merkle.Root)
		n.bHatChan = make(chan []*big.Int)
		n.pHatChan = make(chan []merkle.Witness)
	}
	return n
}

func (n *Node) step1() {
	// receive from BACSS.share

	// BDDG init

}

func (n *Node) sampleConstruct() {
	flag := make([]int, n.dataScale)
	N := float64(n.dataScale) * n.percentage
	var pi []byte
	bigD := new(big.Int).SetInt64(int64(n.dataScale))
	for i := 0; i < int(N); i++ {
		content := append(pi, []byte(strconv.Itoa(i))...)
		hasher := md5.New()
		index := new(big.Int).Mod(new(big.Int).SetBytes(hasher.Sum(content)), bigD)
		flag[index.Int64()] = 1
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

func shareSecret(secret, n, t int, p *big.Int) []*big.Int {
	poly, _ := polynomial.New(t)
	poly.Rand(p)
	err := poly.SetCoefficient(0, int64(secret))
	if err != nil {
		return nil
	}
	shares := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		shares[i] = poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
	}
	return shares
}

func shareSecretBig(secret *big.Int, n, t int, p *big.Int) []*big.Int {
	poly, _ := polynomial.New(t)
	poly.Rand(p)
	err := poly.SetCoefficientBig(0, secret)
	if err != nil {
		return nil
	}
	shares := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		shares[i] = poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
	}
	return shares
}

func reconstruct(x, y []*big.Int, p *big.Int) *big.Int {
	poly, err := polynomial.LagrangeInterpolation(x, y, p)
	if err != nil {
		return nil
	}
	s, _ := poly.GetCoefficient(0)
	return s
}

func batchDDGOffline(nodeNum, degree int, p *big.Int) (*big.Int, []*big.Int, []*big.Int) {
	// generate R and R^2 share
	R := rand.Int() % 2
	RShares := shareSecret(R, nodeNum, degree, p)
	R2Shares := shareSecret(R*R, nodeNum, degree, p)
	return new(big.Int).SetInt64(int64(R)), RShares, R2Shares
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
		poly.SetCoefficient(0, 0)
	}
	RShares := make([]*big.Int, n.dataScale)
	bBias := make([]int, n.dataScale)
	boShares := make([]*big.Int, n.dataScale)
	for i := 0; i < n.dataScale; i++ {
		bBias[i] = rand.Int() % 2
	}
	bo_add_R_shares := make([]*big.Int, n.dataScale)
	for i := 0; i < n.dataScale; i++ {
		bo_add_R_shares[i] = new(big.Int).Add(boShares[i], RShares[i])

	}
}

// Run start node
func (n *Node) Run() {

}
