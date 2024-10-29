package kzg

import (
	//"fmt"
	"github.com/Nik-U/pbc"
	"math/big"
	//"TMAABE/tmaabe"
	//"crypto/sha256"
)

type PK struct{
	pairing	*pbc.Pairing
	t		int
	g		[]*pbc.Element
	N		*big.Int
}

func (pk *PK) SetPairing(pairing *pbc.Pairing){
	pk.pairing = pairing
}

func (pk *PK) Getg()([]*pbc.Element){
	return pk.g
}

func (pk *PK) SetT(t int){
	pk.t = t
}
func (pk *PK) GetT()(int){
	return pk.t
}


func (pk *PK) SetGenerateG(g *pbc.Element){
	pk.g = make([]*pbc.Element, pk.t+1)
	pk.g[0] = g
}

func NewPublicKeyFromBytes(pairing *pbc.Pairing, g[][]byte)(*PK){
	pk := new(PK)
	pk.pairing = pairing
	pk.g = make([]*pbc.Element, len(g))
	for i:=0;i<len(g);i++{
		pk.g[i] = pairing.NewG1().SetBytes(g[0])
	}
	pk.t = len(g)-1

	n0, _ := new(big.Int).SetString("1363895147340162124487750544377566700025348452567", 10)
	n1, _ := new(big.Int).SetString("1257354545315887944833595666025792933231792977521", 10)
	n2, _ := new(big.Int).SetString("1296657106138026641358592699056954007605324218609", 10)
	n := new(big.Int)
	n.Mul(n0, n1)
	n.Mul(n, n2)

	pk.N = n
	return pk
}