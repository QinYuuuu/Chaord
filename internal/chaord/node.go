package chaord

import (
	"Chaord/internal/osv"
	"Chaord/pkg/crypto/commit/merkle"
	"bytes"
	"crypto/md5"
	"math/big"
	"math/rand"
	"strconv"
)

type Node struct {
	dataScale  int
	percentage float64
	rChan      chan merkle.Root
	r1Chan     chan merkle.Root

	bHatChan chan []*big.Int
	pHatChan chan []merkle.Witness

	r  merkle.Root
	r1 merkle.Root

	random rand.Rand

	osv0  *osv.Node
	osv1  *osv.Node
	osvTX *osv.Node
}

func (n *Node) Init() {

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
	// RBC r1
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

func (n *Node) batchDDG() {
	// invoke verify s^2 - s = 0
}

func (n *Node) ODOLocal() {
	// generate shares of s and s^2

}

func (n *Node) Run() {

}
