package internal

import (
	"Chaord/pkg/crypto/commit/merkle"
	"crypto/hasher"
	"math/big"
	"strconv"
)

type Node struct {
	scale          int
	percentage     float64
	merkleRootChan chan merkle.Root
}

func (n *Node) Init() {

}

func (n *Node) step1() {
	// receive from BACSS.share

	// BDDG init

}

func (n *Node) step2() {
	flag := make([]int, n.scale)
	N := float64(n.scale) * n.percentage
	var pi []byte
	bigD := new(big.Int).SetInt64(int64(n.scale))
	for i := 0; i < int(N); i++ {
		content := append(pi, []byte(strconv.Itoa(i))...)
		index := new(big.Int).Mod(new(big.Int).SetBytes(hasher.MD5Hasher(content)), bigD)
		flag[index.Int64()] = 1
	}
}

func (n *Node) Run() {

}
